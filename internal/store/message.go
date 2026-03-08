package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/d6o/aiboard/internal/model"
)

type MessageStore struct {
	db *sql.DB
}

func NewMessageStore(db *sql.DB) *MessageStore {
	return &MessageStore{db: db}
}

func (s *MessageStore) FindAll(limit int, before string) ([]model.Message, error) {
	query := `SELECT m.id, m.user_id, m.content, m.created_at,
		u.id, u.name, u.avatar_color, u.created_at, u.updated_at
		FROM messages m JOIN users u ON m.user_id = u.id`
	var args []any
	if before != "" {
		query += ` WHERE m.id < $1 ORDER BY m.created_at DESC LIMIT $2`
		args = append(args, before, limit)
	} else {
		query += ` ORDER BY m.created_at DESC LIMIT $1`
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var m model.Message
		var u model.User
		if err := rows.Scan(&m.ID, &m.UserID, &m.Content, &m.CreatedAt,
			&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		m.User = &u
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (s *MessageStore) FindByID(id string) (model.Message, error) {
	var m model.Message
	err := s.db.QueryRow(
		`SELECT id, user_id, content, created_at FROM messages WHERE id = $1`, id,
	).Scan(&m.ID, &m.UserID, &m.Content, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return m, model.ErrNotFound
	}
	return m, err
}

func (s *MessageStore) Create(userID, content string) (model.Message, error) {
	var m model.Message
	err := s.db.QueryRow(
		`INSERT INTO messages (user_id, content) VALUES ($1, $2)
		 RETURNING id, user_id, content, created_at`,
		userID, content,
	).Scan(&m.ID, &m.UserID, &m.Content, &m.CreatedAt)
	return m, err
}

func (s *MessageStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM messages WHERE id = $1`, id)
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

func (s *MessageStore) UnreadCount(userID string) (int, error) {
	var lastRead sql.NullTime
	err := s.db.QueryRow(
		`SELECT last_read_at FROM message_reads WHERE user_id = $1`, userID,
	).Scan(&lastRead)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	var count int
	if lastRead.Valid {
		err = s.db.QueryRow(
			`SELECT COUNT(*) FROM messages WHERE created_at > $1`, lastRead.Time,
		).Scan(&count)
	} else {
		err = s.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&count)
	}
	return count, err
}

func (s *MessageStore) MarkRead(userID string) error {
	_, err := s.db.Exec(
		`INSERT INTO message_reads (user_id, last_read_at) VALUES ($1, $2)
		 ON CONFLICT (user_id) DO UPDATE SET last_read_at = $2`,
		userID, time.Now(),
	)
	return err
}
