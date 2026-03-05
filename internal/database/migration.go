package database

import "database/sql"

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Run() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT UNIQUE NOT NULL,
			avatar_color TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS cards (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			priority INTEGER NOT NULL CHECK (priority >= 1 AND priority <= 5),
			column_name TEXT NOT NULL CHECK (column_name IN ('todo', 'doing', 'done')),
			sort_order INTEGER NOT NULL DEFAULT 0,
			reporter_id UUID NOT NULL REFERENCES users(id),
			assignee_id UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS subtasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			completed BOOLEAN NOT NULL DEFAULT FALSE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT UNIQUE NOT NULL,
			color TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS card_tags (
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (card_id, tag_id)
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			card_id UUID NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id),
			message TEXT NOT NULL,
			card_id UUID REFERENCES cards(id) ON DELETE SET NULL,
			read BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS activity_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			action TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id UUID NOT NULL,
			user_id UUID NOT NULL REFERENCES users(id),
			details TEXT NOT NULL DEFAULT '',
			card_id UUID REFERENCES cards(id) ON DELETE SET NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS idempotency_keys (
			key TEXT PRIMARY KEY,
			response_status INTEGER NOT NULL,
			response_body BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cards_column ON cards(column_name)`,
		`CREATE INDEX IF NOT EXISTS idx_cards_assignee ON cards(assignee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_cards_reporter ON cards(reporter_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subtasks_card ON subtasks(card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_card ON comments(card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_card ON activity_log(card_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_user ON activity_log(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_action ON activity_log(action)`,
	}

	for _, stmt := range statements {
		if _, err := m.db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}
