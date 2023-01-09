package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// By default, the keys in the JSON object are equal to the field names in the struct ( ID,
// CreatedAt, Title and so on).
type Movie struct {
	ID        int64     `json:"id"`                       // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                        // Timestamp for when the movie is added to our database, "-" directive, hidden in response
	Title     string    `json:"title"`                    // Movie title
	Year      int32     `json:"year,omitempty"`           // Movie release year, "omitempty" - hide from response if empty
	Runtime   int32     `json:"runtime,omitempty,string"` // Movie runtime (in minutes), "string" - convert int to string
	Genres    []string  `json:"genres,omitempty"`         // Slice of genres for the movie (romance, comedy, etc.)
	Version   int32     `json:"version"`                  // The version number starts at 1 and will be incremented each
	// time the movie information is updated
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// method for inserting a new record in the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies(title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	return m.DB.QueryRow(query, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres)).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// func (d DirectorModel) InsertDirector(directors *Directors) error {
// 	query := `
// 		INSERT INTO directors(name, year, DOB)
// 		VALUES ($1, $2, $3)
// 		RETURNING id, name`

// 	return d.DB.QueryRow(query, &directors.Name, &directors.Surname).Scan(&directors.ID, &directors.Name)
// }

// method for fetching a specific record from the movies table.
func (m MovieModel) GetById(id int64) (*Movie, error) {
	var movie Movie
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id, created_at, title, year, runtime, genres, version FROM movies WHERE id = $1`

	result := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	if result != nil {
		switch {
		case errors.Is(result, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, result
		}
	}

	return &movie, nil
}

func (m MovieModel) Get(title string, genres []string, filters Filters) ([]*Movie, error) {

	query := fmt.Sprintf(`
	SELECT id, created_at, title, year, runtime, genres, version
FROM movies
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (genres @> $2 OR $2 = '{}')
ORDER BY %s %s, id ASC
LIMIT 5 OFFSET 10`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, pq.Array(genres))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []*Movie{}

	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return movies, nil
}

// method for updating a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	query := `UPDATE movies SET title = $1, year = $2, runtime = $3, genres = $4 WHERE id = $5`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// method for deleting a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM movies WHERE id=$1`

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
