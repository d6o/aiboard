package database

import "database/sql"

type Seeder struct {
	db *sql.DB
}

func NewSeeder(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) Run() error {
	return s.seedTags()
}

func (s *Seeder) seedTags() error {
	tags := []struct {
		name  string
		color string
	}{
		{"bug", "#EF4444"},
		{"feature", "#3B82F6"},
		{"enhancement", "#8B5CF6"},
		{"urgent", "#F97316"},
		{"design", "#EC4899"},
		{"backend", "#10B981"},
		{"frontend", "#06B6D4"},
	}

	for _, t := range tags {
		_, err := s.db.Exec(
			`INSERT INTO tags (name, color) VALUES ($1, $2) ON CONFLICT (name) DO NOTHING`,
			t.name, t.color,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
