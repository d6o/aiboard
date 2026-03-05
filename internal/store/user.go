package store

import (
	"database/sql"
	"errors"

	"github.com/d6o/aiboard/internal/model"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) FindAll() ([]model.User, error) {
	rows, err := s.db.Query(`SELECT id, name, avatar_color, created_at, updated_at FROM users ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *UserStore) FindByID(id string) (model.User, error) {
	var u model.User
	err := s.db.QueryRow(
		`SELECT id, name, avatar_color, created_at, updated_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, model.ErrNotFound
	}
	return u, err
}

func (s *UserStore) FindByName(name string) (model.User, error) {
	var u model.User
	err := s.db.QueryRow(
		`SELECT id, name, avatar_color, created_at, updated_at FROM users WHERE LOWER(name) = LOWER($1)`, name,
	).Scan(&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, model.ErrNotFound
	}
	return u, err
}

func (s *UserStore) Create(name, avatarColor string) (model.User, error) {
	var u model.User
	err := s.db.QueryRow(
		`INSERT INTO users (name, avatar_color) VALUES ($1, $2)
		 RETURNING id, name, avatar_color, created_at, updated_at`,
		name, avatarColor,
	).Scan(&u.ID, &u.Name, &u.AvatarColor, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return u, model.ErrDuplicateName
		}
		return u, err
	}
	return u, nil
}
