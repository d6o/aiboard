package model

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrDuplicateSubtaskName = errors.New("duplicate subtask title within card")
	ErrSubtaskLimit         = errors.New("card cannot have more than 20 subtasks")
	ErrDuplicateName        = errors.New("name already exists")
	ErrTagAlreadyAttached   = errors.New("tag is already attached to this card")
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}
