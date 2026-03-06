package store

import (
	"database/sql"
	"errors"

	"github.com/d6o/aiboard/internal/model"
)

type FileStore struct {
	db *sql.DB
}

func NewFileStore(db *sql.DB) *FileStore {
	return &FileStore{db: db}
}

func (s *FileStore) FindByCardID(cardID string) ([]model.File, error) {
	rows, err := s.db.Query(
		`SELECT id, card_id, filename, content_type, size, created_at
		 FROM files WHERE card_id = $1 ORDER BY created_at`, cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.File
	for rows.Next() {
		var f model.File
		if err := rows.Scan(&f.ID, &f.CardID, &f.Filename, &f.ContentType, &f.Size, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (s *FileStore) FindByID(id string) (model.File, error) {
	var f model.File
	err := s.db.QueryRow(
		`SELECT id, card_id, filename, content_type, size, created_at FROM files WHERE id = $1`, id,
	).Scan(&f.ID, &f.CardID, &f.Filename, &f.ContentType, &f.Size, &f.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return f, model.ErrNotFound
	}
	return f, err
}

func (s *FileStore) Create(cardID, filename, contentType string, size int64) (model.File, error) {
	var f model.File
	err := s.db.QueryRow(
		`INSERT INTO files (card_id, filename, content_type, size) VALUES ($1, $2, $3, $4)
		 RETURNING id, card_id, filename, content_type, size, created_at`,
		cardID, filename, contentType, size,
	).Scan(&f.ID, &f.CardID, &f.Filename, &f.ContentType, &f.Size, &f.CreatedAt)
	return f, err
}

func (s *FileStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM files WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return model.ErrNotFound
	}
	return nil
}
