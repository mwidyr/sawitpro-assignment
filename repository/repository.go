package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Repository struct {
	Db *sql.DB
}

type NewRepositoryOptions struct {
	Dsn string
}

func NewRepository(opts NewRepositoryOptions) *Repository {
	db, err := sql.Open("postgres", opts.Dsn)
	if err != nil {
		panic(err)
	}

	// verify connection
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	// init schema
	schema := `
	CREATE TABLE IF NOT EXISTS test (
	id serial PRIMARY KEY,
	name VARCHAR ( 50 ) UNIQUE NOT NULL
);
	`
	if _, err := db.Exec(schema); err != nil {
		panic(fmt.Errorf("failed to init schema: %w", err))
	}

	return &Repository{
		Db: db,
	}
}
