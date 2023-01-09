package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Directors struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Surname string   `json:"surname"`
	Awards  []string `json:"awards"`
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type DirectorModel struct {
	DB *sql.DB
}

func (d DirectorModel) Insert(director *Directors) error {
	query := `
		INSERT INTO directors(name, surname, awards)
		VALUES ($1, $2, $3)
		RETURNING id,name`
	return d.DB.QueryRow(query, &director.Name, &director.Surname, pq.Array(&director.Awards)).Scan(&director.ID, &director.Name)
}

func (d DirectorModel) GetByName(name string) (*Directors, error) {
	var director Directors
	if name == "" {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id, name, surname, awards FROM directors WHERE name = $1`

	result := d.DB.QueryRow(query, name).Scan(
		&director.ID,
		&director.Name,
		&director.Surname,
		&director.Awards,
	)

	if result != nil {
		switch {
		case errors.Is(result, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, result
		}
	}

	return &director, nil
}

func (d DirectorModel) Get(name string, filters Filters) ([]*Directors, error) {

	query := fmt.Sprintf(`
	SELECT id, name, surname, awards
	FROM directors
	WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	ORDER BY %s %s, id ASC`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := d.DB.QueryContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	directors := []*Directors{}

	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var director Directors
		err := rows.Scan(
			&director.ID,
			&director.Name,
			&director.Surname,
			pq.Array(&director.Awards),
		)
		if err != nil {
			return nil, err
		}
		directors = append(directors, &director)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return directors, nil
}
