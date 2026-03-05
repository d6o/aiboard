package store

import (
	"database/sql"
	"errors"

	"github.com/d6o/aiboard/internal/model"
)

type TagStore struct {
	db *sql.DB
}

func NewTagStore(db *sql.DB) *TagStore {
	return &TagStore{db: db}
}

func (s *TagStore) FindAll() ([]model.Tag, error) {
	rows, err := s.db.Query(`SELECT id, name, color, created_at FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *TagStore) FindByID(id string) (model.Tag, error) {
	var t model.Tag
	err := s.db.QueryRow(`SELECT id, name, color, created_at FROM tags WHERE id = $1`, id).
		Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return t, model.ErrNotFound
	}
	return t, err
}

func (s *TagStore) Create(name, color string) (model.Tag, error) {
	var t model.Tag
	err := s.db.QueryRow(
		`INSERT INTO tags (name, color) VALUES ($1, $2) RETURNING id, name, color, created_at`,
		name, color,
	).Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return t, model.ErrDuplicateName
		}
		return t, err
	}
	return t, nil
}

func (s *TagStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM tags WHERE id = $1`, id)
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

func (s *TagStore) AttachToCard(cardID, tagID string) error {
	_, err := s.db.Exec(`INSERT INTO card_tags (card_id, tag_id) VALUES ($1, $2)`, cardID, tagID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrTagAlreadyAttached
		}
		return err
	}
	return nil
}

func (s *TagStore) DetachFromCard(cardID, tagID string) error {
	result, err := s.db.Exec(`DELETE FROM card_tags WHERE card_id = $1 AND tag_id = $2`, cardID, tagID)
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

func (s *TagStore) FindByCardID(cardID string) ([]model.Tag, error) {
	rows, err := s.db.Query(
		`SELECT t.id, t.name, t.color, t.created_at FROM tags t
		 JOIN card_tags ct ON t.id = ct.tag_id WHERE ct.card_id = $1 ORDER BY t.name`, cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}
