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
		message_reads,
		messages,
		standup_entries,
		standups,
		standup_config,
		activity_log,
		notifications,
		comments,
		files,
		card_tags,
		cards,
		tags,
		users
		CASCADE`)
	return err
}
