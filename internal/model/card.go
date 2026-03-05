package model

import "time"

type Card struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Priority         int       `json:"priority"`
	Column           Column    `json:"column"`
	SortOrder        int       `json:"sort_order"`
	ReporterID       string    `json:"reporter_id"`
	AssigneeID       string    `json:"assignee_id"`
	Reporter         *User     `json:"reporter,omitempty"`
	Assignee         *User     `json:"assignee,omitempty"`
	Tags             []Tag     `json:"tags,omitempty"`
	Subtasks         []Subtask `json:"subtasks,omitempty"`
	Comments         []Comment `json:"comments,omitempty"`
	SubtaskTotal     int       `json:"subtask_total"`
	SubtaskCompleted int       `json:"subtask_completed"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CardFilter struct {
	Column   string
	Assignee string
	Reporter string
	Tag      string
	Priority int
}
