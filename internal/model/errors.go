package model

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrDuplicateName      = errors.New("name already exists")
	ErrTagAlreadyAttached = errors.New("tag is already attached to this card")
	ErrUserInUse          = errors.New("user is referenced by existing cards")
	ErrChildrenNotDone    = errors.New("all child cards must be in done before this card can move to done")
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
