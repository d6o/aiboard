package store

import (
	"database/sql"
	"errors"

	"github.com/d6o/aiboard/internal/model"
)

type CommentStore struct {
	db *sql.DB
}

func NewCommentStore(db *sql.DB) *CommentStore {
	return &CommentStore{db: db}
}

func (s *CommentStore) FindByCardID(cardID string) ([]model.Comment, error) {
	rows, err := s.db.Query(
		`SELECT c.id, c.card_id, c.user_id, c.content, c.created_at,
			u.id, u.name, u.avatar_color, u.created_at, u.updated_at
		 FROM comments c JOIN users u ON c.user_id = u.id
		 WHERE c.card_id = $1 ORDER BY c.created_at`, cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		var u model.User
		if err := rows.Scan(&c.ID, &c.CardID, &c.UserID, &c.Content, &c.CreatedAt,
			&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		c.User = &u
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

func (s *CommentStore) FindByID(id string) (model.Comment, error) {
	var c model.Comment
	err := s.db.QueryRow(
		`SELECT id, card_id, user_id, content, created_at FROM comments WHERE id = $1`, id,
	).Scan(&c.ID, &c.CardID, &c.UserID, &c.Content, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrNotFound
	}
	return c, err
}

func (s *CommentStore) Create(cardID, userID, content string) (model.Comment, error) {
	var c model.Comment
	err := s.db.QueryRow(
		`INSERT INTO comments (card_id, user_id, content) VALUES ($1, $2, $3)
		 RETURNING id, card_id, user_id, content, created_at`,
		cardID, userID, content,
	).Scan(&c.ID, &c.CardID, &c.UserID, &c.Content, &c.CreatedAt)
	return c, err
}

func (s *CommentStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM comments WHERE id = $1`, id)
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
