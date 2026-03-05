package service

import (
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type tagStore interface {
	FindAll() ([]model.Tag, error)
	FindByID(id string) (model.Tag, error)
	Create(name, color string) (model.Tag, error)
	Delete(id string) error
	AttachToCard(cardID, tagID string) error
	DetachFromCard(cardID, tagID string) error
}

type tagActivityLogger interface {
	Create(action, resourceType, resourceID, userID, details, cardID string) (model.ActivityEntry, error)
}

type TagService struct {
	tags     tagStore
	activity tagActivityLogger
}

func NewTagService(tags tagStore, activity tagActivityLogger) *TagService {
	return &TagService{tags: tags, activity: activity}
}

func (s *TagService) List() ([]model.Tag, error) {
	return s.tags.FindAll()
}

func (s *TagService) Create(name, color, actingUserID string) (model.Tag, error) {
	var fieldErrors []model.FieldError

	name = strings.TrimSpace(name)
	if name == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "name", Message: "name is required"})
	}

	color = strings.TrimSpace(color)
	if color == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "color", Message: "color is required"})
	}

	if len(fieldErrors) > 0 {
		return model.Tag{}, &model.ValidationError{Fields: fieldErrors}
	}

	tag, err := s.tags.Create(name, color)
	if err != nil {
		return tag, err
	}

	s.activity.Create("created", "tag", tag.ID, actingUserID, "tag created: "+name, "")
	return tag, nil
}

func (s *TagService) Delete(id, actingUserID string) error {
	tag, err := s.tags.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.tags.Delete(id); err != nil {
		return err
	}

	s.activity.Create("deleted", "tag", id, actingUserID, "tag deleted: "+tag.Name, "")
	return nil
}

func (s *TagService) AttachToCard(cardID, tagID, actingUserID string) error {
	if err := s.tags.AttachToCard(cardID, tagID); err != nil {
		return err
	}
	s.activity.Create("attached", "tag", tagID, actingUserID, "tag attached to card", cardID)
	return nil
}

func (s *TagService) DetachFromCard(cardID, tagID, actingUserID string) error {
	if err := s.tags.DetachFromCard(cardID, tagID); err != nil {
		return err
	}
	s.activity.Create("detached", "tag", tagID, actingUserID, "tag detached from card", cardID)
	return nil
}
