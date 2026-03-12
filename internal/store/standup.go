package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/d6o/aiboard/internal/model"
)

type StandupStore struct {
	db *sql.DB
}

func NewStandupStore(db *sql.DB) *StandupStore {
	return &StandupStore{db: db}
}

func (s *StandupStore) GetConfig() (model.StandupConfig, error) {
	var cfg model.StandupConfig
	err := s.db.QueryRow(
		`SELECT interval_hours, enabled FROM standup_config LIMIT 1`,
	).Scan(&cfg.IntervalHours, &cfg.Enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return model.StandupConfig{IntervalHours: 24, Enabled: false}, nil
	}
	return cfg, err
}

func (s *StandupStore) SaveConfig(intervalHours int, enabled bool) (model.StandupConfig, error) {
	var cfg model.StandupConfig
	err := s.db.QueryRow(
		`INSERT INTO standup_config (id, interval_hours, enabled)
		 VALUES (1, $1, $2)
		 ON CONFLICT (id) DO UPDATE SET interval_hours = $1, enabled = $2
		 RETURNING interval_hours, enabled`,
		intervalHours, enabled,
	).Scan(&cfg.IntervalHours, &cfg.Enabled)
	return cfg, err
}

func (s *StandupStore) FindAll(limit int) ([]model.Standup, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT id, number, start_time, end_time, created_at
		 FROM standups ORDER BY number DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var standups []model.Standup
	for rows.Next() {
		var st model.Standup
		if err := rows.Scan(&st.ID, &st.Number, &st.StartTime, &st.EndTime, &st.CreatedAt); err != nil {
			return nil, err
		}
		standups = append(standups, st)
	}
	return standups, rows.Err()
}

func (s *StandupStore) FindByID(id string) (model.Standup, error) {
	var st model.Standup
	err := s.db.QueryRow(
		`SELECT id, number, start_time, end_time, created_at FROM standups WHERE id = $1`, id,
	).Scan(&st.ID, &st.Number, &st.StartTime, &st.EndTime, &st.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return st, model.ErrNotFound
	}
	return st, err
}

func (s *StandupStore) FindLatest() (model.Standup, error) {
	var st model.Standup
	err := s.db.QueryRow(
		`SELECT id, number, start_time, end_time, created_at
		 FROM standups ORDER BY number DESC LIMIT 1`,
	).Scan(&st.ID, &st.Number, &st.StartTime, &st.EndTime, &st.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return st, model.ErrNotFound
	}
	return st, err
}

func (s *StandupStore) Create(number int, startTime, endTime time.Time) (model.Standup, error) {
	var st model.Standup
	err := s.db.QueryRow(
		`INSERT INTO standups (number, start_time, end_time) VALUES ($1, $2, $3)
		 RETURNING id, number, start_time, end_time, created_at`,
		number, startTime, endTime,
	).Scan(&st.ID, &st.Number, &st.StartTime, &st.EndTime, &st.CreatedAt)
	return st, err
}

func (s *StandupStore) FindEntriesByStandupID(standupID string) ([]model.StandupEntry, error) {
	rows, err := s.db.Query(
		`SELECT e.id, e.standup_id, e.user_id, e.content, e.created_at,
			u.id, u.name, u.avatar_color, u.created_at, u.updated_at
		 FROM standup_entries e JOIN users u ON e.user_id = u.id
		 WHERE e.standup_id = $1 ORDER BY e.created_at`, standupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.StandupEntry
	for rows.Next() {
		var e model.StandupEntry
		var u model.User
		if err := rows.Scan(&e.ID, &e.StandupID, &e.UserID, &e.Content, &e.CreatedAt,
			&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		e.User = &u
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *StandupStore) CreateEntry(standupID, userID, content string) (model.StandupEntry, error) {
	var e model.StandupEntry
	err := s.db.QueryRow(
		`INSERT INTO standup_entries (standup_id, user_id, content) VALUES ($1, $2, $3)
		 RETURNING id, standup_id, user_id, content, created_at`,
		standupID, userID, content,
	).Scan(&e.ID, &e.StandupID, &e.UserID, &e.Content, &e.CreatedAt)
	return e, err
}

func (s *StandupStore) NextNumber() (int, error) {
	var maxNum sql.NullInt64
	err := s.db.QueryRow(`SELECT MAX(number) FROM standups`).Scan(&maxNum)
	if err != nil {
		return 0, err
	}
	if maxNum.Valid {
		return int(maxNum.Int64) + 1, nil
	}
	return 1, nil
}
