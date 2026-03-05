package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type CardStore struct {
	db *sql.DB
}

func NewCardStore(db *sql.DB) *CardStore {
	return &CardStore{db: db}
}

func (s *CardStore) FindAll(filter model.CardFilter) ([]model.Card, error) {
	query := `SELECT c.id, c.title, c.description, c.priority, c.column_name, c.sort_order,
		c.reporter_id, c.assignee_id, c.created_at, c.updated_at,
		r.id, r.name, r.avatar_color, r.created_at, r.updated_at,
		a.id, a.name, a.avatar_color, a.created_at, a.updated_at,
		COALESCE((SELECT COUNT(*) FROM subtasks WHERE card_id = c.id), 0),
		COALESCE((SELECT COUNT(*) FROM subtasks WHERE card_id = c.id AND completed = true), 0)
		FROM cards c
		JOIN users r ON c.reporter_id = r.id
		JOIN users a ON c.assignee_id = a.id`

	var conditions []string
	var args []any
	argIdx := 1

	if filter.Column != "" {
		conditions = append(conditions, fmt.Sprintf("c.column_name = $%d", argIdx))
		args = append(args, filter.Column)
		argIdx++
	}
	if filter.Assignee != "" {
		conditions = append(conditions, fmt.Sprintf("c.assignee_id = $%d", argIdx))
		args = append(args, filter.Assignee)
		argIdx++
	}
	if filter.Reporter != "" {
		conditions = append(conditions, fmt.Sprintf("c.reporter_id = $%d", argIdx))
		args = append(args, filter.Reporter)
		argIdx++
	}
	if filter.Priority > 0 {
		conditions = append(conditions, fmt.Sprintf("c.priority = $%d", argIdx))
		args = append(args, filter.Priority)
		argIdx++
	}
	if filter.Tag != "" {
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM card_tags ct JOIN tags t ON ct.tag_id = t.id WHERE ct.card_id = c.id AND t.id = $%d)", argIdx))
		args = append(args, filter.Tag)
		argIdx++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY c.column_name, c.sort_order, c.created_at"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []model.Card
	for rows.Next() {
		var c model.Card
		var reporter model.User
		var assignee model.User
		if err := rows.Scan(
			&c.ID, &c.Title, &c.Description, &c.Priority, &c.Column, &c.SortOrder,
			&c.ReporterID, &c.AssigneeID, &c.CreatedAt, &c.UpdatedAt,
			&reporter.ID, &reporter.Name, &reporter.AvatarColor, &reporter.CreatedAt, &reporter.UpdatedAt,
			&assignee.ID, &assignee.Name, &assignee.AvatarColor, &assignee.CreatedAt, &assignee.UpdatedAt,
			&c.SubtaskTotal, &c.SubtaskCompleted,
		); err != nil {
			return nil, err
		}
		c.Reporter = &reporter
		c.Assignee = &assignee
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (s *CardStore) FindByID(id string) (model.Card, error) {
	var c model.Card
	var reporter model.User
	var assignee model.User
	err := s.db.QueryRow(
		`SELECT c.id, c.title, c.description, c.priority, c.column_name, c.sort_order,
			c.reporter_id, c.assignee_id, c.created_at, c.updated_at,
			r.id, r.name, r.avatar_color, r.created_at, r.updated_at,
			a.id, a.name, a.avatar_color, a.created_at, a.updated_at,
			COALESCE((SELECT COUNT(*) FROM subtasks WHERE card_id = c.id), 0),
			COALESCE((SELECT COUNT(*) FROM subtasks WHERE card_id = c.id AND completed = true), 0)
		FROM cards c
		JOIN users r ON c.reporter_id = r.id
		JOIN users a ON c.assignee_id = a.id
		WHERE c.id = $1`, id,
	).Scan(
		&c.ID, &c.Title, &c.Description, &c.Priority, &c.Column, &c.SortOrder,
		&c.ReporterID, &c.AssigneeID, &c.CreatedAt, &c.UpdatedAt,
		&reporter.ID, &reporter.Name, &reporter.AvatarColor, &reporter.CreatedAt, &reporter.UpdatedAt,
		&assignee.ID, &assignee.Name, &assignee.AvatarColor, &assignee.CreatedAt, &assignee.UpdatedAt,
		&c.SubtaskTotal, &c.SubtaskCompleted,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrNotFound
	}
	if err != nil {
		return c, err
	}
	c.Reporter = &reporter
	c.Assignee = &assignee
	return c, nil
}

func (s *CardStore) Create(title, description string, priority int, col model.Column, sortOrder int, reporterID, assigneeID string) (model.Card, error) {
	var c model.Card
	err := s.db.QueryRow(
		`INSERT INTO cards (title, description, priority, column_name, sort_order, reporter_id, assignee_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, title, description, priority, column_name, sort_order, reporter_id, assignee_id, created_at, updated_at`,
		title, description, priority, string(col), sortOrder, reporterID, assigneeID,
	).Scan(&c.ID, &c.Title, &c.Description, &c.Priority, &c.Column, &c.SortOrder,
		&c.ReporterID, &c.AssigneeID, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (s *CardStore) Update(id, title, description string, priority int, col model.Column, sortOrder int, reporterID, assigneeID string) (model.Card, error) {
	var c model.Card
	err := s.db.QueryRow(
		`UPDATE cards SET title = $2, description = $3, priority = $4, column_name = $5,
			sort_order = $6, reporter_id = $7, assignee_id = $8, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, title, description, priority, column_name, sort_order, reporter_id, assignee_id, created_at, updated_at`,
		id, title, description, priority, string(col), sortOrder, reporterID, assigneeID,
	).Scan(&c.ID, &c.Title, &c.Description, &c.Priority, &c.Column, &c.SortOrder,
		&c.ReporterID, &c.AssigneeID, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrNotFound
	}
	return c, err
}

func (s *CardStore) UpdateColumn(id string, col model.Column) (model.Card, error) {
	var c model.Card
	err := s.db.QueryRow(
		`UPDATE cards SET column_name = $2, updated_at = NOW() WHERE id = $1
		 RETURNING id, title, description, priority, column_name, sort_order, reporter_id, assignee_id, created_at, updated_at`,
		id, string(col),
	).Scan(&c.ID, &c.Title, &c.Description, &c.Priority, &c.Column, &c.SortOrder,
		&c.ReporterID, &c.AssigneeID, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrNotFound
	}
	return c, err
}

func (s *CardStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM cards WHERE id = $1`, id)
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

func (s *CardStore) NextSortOrder(col model.Column) (int, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(sort_order) FROM cards WHERE column_name = $1`, string(col)).Scan(&maxOrder)
	if err != nil {
		return 0, err
	}
	if maxOrder.Valid {
		return int(maxOrder.Int64) + 1, nil
	}
	return 0, nil
}
