package model

import "time"

type File struct {
	ID          string    `json:"id"`
	CardID      string    `json:"card_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	RawURL      string    `json:"raw_url"`
	CreatedAt   time.Time `json:"created_at"`
}
