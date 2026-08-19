package application

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Application represents an individual request
type Application struct {
	// NOTE: json:"..." - what is the field called in json
	ID int `json:"id"` 
	Client string `json:"client"` // WHY: with a capital, so that the field is visible from outside the package
	Amount int `json:"amount"` // WHY: without the tag, the JSON would be Amount with a capital
	Term int `json:"term"`
	Status string `json:"status"`
}

// Store represents a simple memory cell
type Store struct {
	db *sql.DB
}

// Close the Store, since it owns the connection, it is also responsible for closing it
func (s *Store) Close() error {
	return s.db.Close()  // WHY: Whoever opened the resource provides the way to close it
}

// NewStore create and returns an exemplar of Store
func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// Add accept the request and saves it to memory, returning the completed request
func (s *Store) Add(app Application) (Application, error) { // WHY: `*` is used so that the original is modified, rather than a copy
	app.Status = "new"

	err := s.db.QueryRow(
		`INSERT INTO applications (client, amount, term, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		app.Client, app.Amount, app.Term, app.Status,
	).Scan(&app.ID)

	if err != nil {
		return Application{}, err
	}

	return app, nil
}