package store

import "database/sql"

type BoardStore struct {
	db *sql.DB
}

func NewBoardStore(db *sql.DB) *BoardStore {
	return &BoardStore{db: db}
}

func (s *BoardStore) Reset() error {
	_, err := s.db.Exec(`TRUNCATE
		idempotency_keys,
		activity_log,
		notifications,
		comments,
		card_tags,
		subtasks,
		cards,
		tags,
		users
		CASCADE`)
	return err
}
