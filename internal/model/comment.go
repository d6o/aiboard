package model

import "time"

type Comment struct {
	ID        string    `json:"id"`
	CardID    string    `json:"card_id"`
	UserID    string    `json:"user_id"`
	User      *User     `json:"user,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
