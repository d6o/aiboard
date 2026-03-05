package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Connection struct {
	DB *sql.DB
}

func NewConnection(databaseURL string) (*Connection, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Connection{DB: db}, nil
}

func (c *Connection) Close() error {
	return c.DB.Close()
}
