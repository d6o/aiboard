package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type ActivityStore struct {
	db *sql.DB
}

func NewActivityStore(db *sql.DB) *ActivityStore {
	return &ActivityStore{db: db}
}

func (s *ActivityStore) Find(filter model.ActivityFilter) ([]model.ActivityEntry, error) {
	query := `SELECT id, action, resource_type, resource_id, user_id, details, COALESCE(card_id::text, ''), created_at
		FROM activity_log`

	var conditions []string
	var args []any
	argIdx := 1

	if filter.CardID != "" {
		conditions = append(conditions, fmt.Sprintf("card_id = $%d", argIdx))
		args = append(args, filter.CardID)
		argIdx++
	}
	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, filter.UserID)
		argIdx++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIdx))
		args = append(args, filter.Action)
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC LIMIT 200"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.ActivityEntry
	for rows.Next() {
		var e model.ActivityEntry
		if err := rows.Scan(&e.ID, &e.Action, &e.ResourceType, &e.ResourceID, &e.UserID, &e.Details, &e.CardID, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *ActivityStore) Create(action, resourceType, resourceID, userID, details, cardID string) (model.ActivityEntry, error) {
	var e model.ActivityEntry
	var err error
	if cardID == "" {
		err = s.db.QueryRow(
			`INSERT INTO activity_log (action, resource_type, resource_id, user_id, details)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, action, resource_type, resource_id, user_id, details, COALESCE(card_id::text, ''), created_at`,
			action, resourceType, resourceID, userID, details,
		).Scan(&e.ID, &e.Action, &e.ResourceType, &e.ResourceID, &e.UserID, &e.Details, &e.CardID, &e.CreatedAt)
	} else {
		err = s.db.QueryRow(
			`INSERT INTO activity_log (action, resource_type, resource_id, user_id, details, card_id)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 RETURNING id, action, resource_type, resource_id, user_id, details, COALESCE(card_id::text, ''), created_at`,
			action, resourceType, resourceID, userID, details, cardID,
		).Scan(&e.ID, &e.Action, &e.ResourceType, &e.ResourceID, &e.UserID, &e.Details, &e.CardID, &e.CreatedAt)
	}
	return e, err
}
