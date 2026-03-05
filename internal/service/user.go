package service

import (
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type userStore interface {
	FindAll() ([]model.User, error)
	FindByID(id string) (model.User, error)
	FindByName(name string) (model.User, error)
	Create(name, avatarColor string) (model.User, error)
	Delete(id string) error
}

type UserService struct {
	users userStore
}

func NewUserService(users userStore) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List() ([]model.User, error) {
	return s.users.FindAll()
}

func (s *UserService) Get(id string) (model.User, error) {
	return s.users.FindByID(id)
}

func (s *UserService) GetByName(name string) (model.User, error) {
	return s.users.FindByName(name)
}

func (s *UserService) Create(name, avatarColor string) (model.User, error) {
	var fieldErrors []model.FieldError

	name = strings.TrimSpace(name)
	if name == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "name", Message: "name is required"})
	}

	avatarColor = strings.TrimSpace(avatarColor)
	if avatarColor == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "avatar_color", Message: "avatar_color is required"})
	}

	if len(fieldErrors) > 0 {
		return model.User{}, &model.ValidationError{Fields: fieldErrors}
	}

	return s.users.Create(name, avatarColor)
}

func (s *UserService) Delete(id string) error {
	return s.users.Delete(id)
}
