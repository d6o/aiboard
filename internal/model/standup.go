package model

import "time"

type StandupConfig struct {
	IntervalHours int  `json:"interval_hours"`
	Enabled       bool `json:"enabled"`
}

type Standup struct {
	ID        string         `json:"id"`
	Number    int            `json:"number"`
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Entries   []StandupEntry `json:"entries,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type StandupEntry struct {
	ID        string    `json:"id"`
	StandupID string    `json:"standup_id"`
	UserID    string    `json:"user_id"`
	User      *User     `json:"user,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
