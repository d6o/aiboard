package model

import "time"

type Subtask struct {
	ID        string    `json:"id"`
	CardID    string    `json:"card_id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
