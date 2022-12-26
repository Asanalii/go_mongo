package data

import (
	"database/sql"
	"time"
)

type Directors struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Surname string    `json:"runtime,omitempty,string"`
	DOB     time.Time `json:"dob"`
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type DirectorModel struct {
	DB *sql.DB
}

func (d DirectorModel) Insert(director *Directors) error {
	query := `
		INSERT INTO directors(name, surname, DOB)
		VALUES ($1, $2, $3)
		RETURNING id,name`

	return d.DB.QueryRow(query, &director.Name, &director.Surname, &director.DOB).Scan(&director.ID, &director.Name)
}
