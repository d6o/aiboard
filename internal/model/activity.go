package model

import "time"

type ActivityEntry struct {
	ID           string    `json:"id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	UserID       string    `json:"user_id"`
	Details      string    `json:"details"`
	CardID       string    `json:"card_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ActivityFilter struct {
	CardID string
	UserID string
	Action string
}
