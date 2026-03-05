package store

import (
	"database/sql"
	"errors"
)

type IdempotencyRecord struct {
	Key            string
	ResponseStatus int
	ResponseBody   []byte
}

type IdempotencyStore struct {
	db *sql.DB
}

func NewIdempotencyStore(db *sql.DB) *IdempotencyStore {
	return &IdempotencyStore{db: db}
}

func (s *IdempotencyStore) Find(key string) (IdempotencyRecord, bool, error) {
	var r IdempotencyRecord
	err := s.db.QueryRow(
		`SELECT key, response_status, response_body FROM idempotency_keys WHERE key = $1`, key,
	).Scan(&r.Key, &r.ResponseStatus, &r.ResponseBody)
	if errors.Is(err, sql.ErrNoRows) {
		return r, false, nil
	}
	if err != nil {
		return r, false, err
	}
	return r, true, nil
}

func (s *IdempotencyStore) Save(key string, status int, body []byte) error {
	_, err := s.db.Exec(
		`INSERT INTO idempotency_keys (key, response_status, response_body) VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO NOTHING`,
		key, status, body,
	)
	return err
}
