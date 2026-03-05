package store

import (
	"database/sql"
	"errors"

	"github.com/d6o/aiboard/internal/model"
)

type SubtaskStore struct {
	db *sql.DB
}

func NewSubtaskStore(db *sql.DB) *SubtaskStore {
	return &SubtaskStore{db: db}
}

func (s *SubtaskStore) FindByCardID(cardID string) ([]model.Subtask, error) {
	rows, err := s.db.Query(
		`SELECT id, card_id, title, completed, sort_order, created_at, updated_at
		 FROM subtasks WHERE card_id = $1 ORDER BY sort_order, created_at`, cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subtasks []model.Subtask
	for rows.Next() {
		var st model.Subtask
		if err := rows.Scan(&st.ID, &st.CardID, &st.Title, &st.Completed, &st.SortOrder, &st.CreatedAt, &st.UpdatedAt); err != nil {
			return nil, err
		}
		subtasks = append(subtasks, st)
	}
	return subtasks, rows.Err()
}

func (s *SubtaskStore) FindByID(id string) (model.Subtask, error) {
	var st model.Subtask
	err := s.db.QueryRow(
		`SELECT id, card_id, title, completed, sort_order, created_at, updated_at FROM subtasks WHERE id = $1`, id,
	).Scan(&st.ID, &st.CardID, &st.Title, &st.Completed, &st.SortOrder, &st.CreatedAt, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return st, model.ErrNotFound
	}
	return st, err
}

func (s *SubtaskStore) Create(cardID, title string, sortOrder int) (model.Subtask, error) {
	var st model.Subtask
	err := s.db.QueryRow(
		`INSERT INTO subtasks (card_id, title, sort_order) VALUES ($1, $2, $3)
		 RETURNING id, card_id, title, completed, sort_order, created_at, updated_at`,
		cardID, title, sortOrder,
	).Scan(&st.ID, &st.CardID, &st.Title, &st.Completed, &st.SortOrder, &st.CreatedAt, &st.UpdatedAt)
	return st, err
}

func (s *SubtaskStore) Update(id, title string, completed bool) (model.Subtask, error) {
	var st model.Subtask
	err := s.db.QueryRow(
		`UPDATE subtasks SET title = $2, completed = $3, updated_at = NOW() WHERE id = $1
		 RETURNING id, card_id, title, completed, sort_order, created_at, updated_at`,
		id, title, completed,
	).Scan(&st.ID, &st.CardID, &st.Title, &st.Completed, &st.SortOrder, &st.CreatedAt, &st.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return st, model.ErrNotFound
	}
	return st, err
}

func (s *SubtaskStore) Delete(id string) error {
	result, err := s.db.Exec(`DELETE FROM subtasks WHERE id = $1`, id)
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

func (s *SubtaskStore) CountByCardID(cardID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM subtasks WHERE card_id = $1`, cardID).Scan(&count)
	return count, err
}

func (s *SubtaskStore) CountIncompleteByCardID(cardID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM subtasks WHERE card_id = $1 AND completed = false`, cardID).Scan(&count)
	return count, err
}

func (s *SubtaskStore) HasDuplicateTitle(cardID, title, excludeID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM subtasks WHERE card_id = $1 AND LOWER(title) = LOWER($2)`
	args := []any{cardID, title}
	if excludeID != "" {
		query += ` AND id != $3`
		args = append(args, excludeID)
	}
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count > 0, err
}

func (s *SubtaskStore) Reorder(cardID string, ids []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range ids {
		_, err := tx.Exec(
			`UPDATE subtasks SET sort_order = $1, updated_at = NOW() WHERE id = $2 AND card_id = $3`,
			i, id, cardID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SubtaskStore) NextSortOrder(cardID string) (int, error) {
	var maxOrder sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(sort_order) FROM subtasks WHERE card_id = $1`, cardID).Scan(&maxOrder)
	if err != nil {
		return 0, err
	}
	if maxOrder.Valid {
		return int(maxOrder.Int64) + 1, nil
	}
	return 0, nil
}
