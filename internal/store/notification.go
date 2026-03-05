package store

import (
	"database/sql"

	"github.com/d6o/aiboard/internal/model"
)

type NotificationStore struct {
	db *sql.DB
}

func NewNotificationStore(db *sql.DB) *NotificationStore {
	return &NotificationStore{db: db}
}

func (s *NotificationStore) FindByUserID(userID string, unreadOnly bool) ([]model.Notification, error) {
	query := `SELECT id, user_id, message, COALESCE(card_id::text, ''), read, created_at
		FROM notifications WHERE user_id = $1`
	if unreadOnly {
		query += ` AND read = false`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Message, &n.CardID, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, rows.Err()
}

func (s *NotificationStore) Create(userID, message, cardID string) (model.Notification, error) {
	var n model.Notification
	var err error
	if cardID == "" {
		err = s.db.QueryRow(
			`INSERT INTO notifications (user_id, message) VALUES ($1, $2)
			 RETURNING id, user_id, message, COALESCE(card_id::text, ''), read, created_at`,
			userID, message,
		).Scan(&n.ID, &n.UserID, &n.Message, &n.CardID, &n.Read, &n.CreatedAt)
	} else {
		err = s.db.QueryRow(
			`INSERT INTO notifications (user_id, message, card_id) VALUES ($1, $2, $3)
			 RETURNING id, user_id, message, COALESCE(card_id::text, ''), read, created_at`,
			userID, message, cardID,
		).Scan(&n.ID, &n.UserID, &n.Message, &n.CardID, &n.Read, &n.CreatedAt)
	}
	return n, err
}

func (s *NotificationStore) MarkRead(id string) error {
	result, err := s.db.Exec(`UPDATE notifications SET read = true WHERE id = $1`, id)
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

func (s *NotificationStore) MarkAllRead(userID string) error {
	_, err := s.db.Exec(`UPDATE notifications SET read = true WHERE user_id = $1 AND read = false`, userID)
	return err
}
