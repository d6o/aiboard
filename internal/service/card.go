package service

import (
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type cardStore interface {
	FindAll(filter model.CardFilter) ([]model.Card, error)
	FindByID(id string) (model.Card, error)
	Create(title, description string, priority int, col model.Column, sortOrder int, reporterID, assigneeID string) (model.Card, error)
	Update(id, title, description string, priority int, col model.Column, sortOrder int, reporterID, assigneeID string) (model.Card, error)
	UpdateColumn(id string, col model.Column) (model.Card, error)
	Delete(id string) error
	NextSortOrder(col model.Column) (int, error)
}

type cardTagFinder interface {
	FindByCardID(cardID string) ([]model.Tag, error)
}

type cardSubtaskFinder interface {
	FindByCardID(cardID string) ([]model.Subtask, error)
}

type cardCommentFinder interface {
	FindByCardID(cardID string) ([]model.Comment, error)
}

type cardActivityLogger interface {
	Create(action, resourceType, resourceID, userID, details, cardID string) (model.ActivityEntry, error)
}

type cardNotifier interface {
	Create(userID, message, cardID string) (model.Notification, error)
}

type CardService struct {
	cards    cardStore
	tags     cardTagFinder
	subtasks cardSubtaskFinder
	comments cardCommentFinder
	activity cardActivityLogger
	notifs   cardNotifier
}

func NewCardService(cards cardStore, tags cardTagFinder, subtasks cardSubtaskFinder, comments cardCommentFinder, activity cardActivityLogger, notifs cardNotifier) *CardService {
	return &CardService{
		cards:    cards,
		tags:     tags,
		subtasks: subtasks,
		comments: comments,
		activity: activity,
		notifs:   notifs,
	}
}

func (s *CardService) List(filter model.CardFilter) ([]model.Card, error) {
	cards, err := s.cards.FindAll(filter)
	if err != nil {
		return nil, err
	}

	for i := range cards {
		tags, err := s.tags.FindByCardID(cards[i].ID)
		if err != nil {
			return nil, err
		}
		cards[i].Tags = tags
	}

	return cards, nil
}

func (s *CardService) Get(id string) (model.Card, error) {
	card, err := s.cards.FindByID(id)
	if err != nil {
		return card, err
	}

	tags, err := s.tags.FindByCardID(id)
	if err != nil {
		return card, err
	}
	card.Tags = tags

	subtasks, err := s.subtasks.FindByCardID(id)
	if err != nil {
		return card, err
	}
	card.Subtasks = subtasks

	comments, err := s.comments.FindByCardID(id)
	if err != nil {
		return card, err
	}
	card.Comments = comments

	return card, nil
}

func (s *CardService) Create(title, description string, priority int, col model.Column, reporterID, assigneeID, actingUserID string) (model.Card, error) {
	if err := validateCard(title, priority, col, reporterID, assigneeID); err != nil {
		return model.Card{}, err
	}

	sortOrder, err := s.cards.NextSortOrder(col)
	if err != nil {
		return model.Card{}, err
	}

	card, err := s.cards.Create(title, description, priority, col, sortOrder, reporterID, assigneeID)
	if err != nil {
		return card, err
	}

	s.activity.Create("created", "card", card.ID, actingUserID, "card created", card.ID)

	return s.cards.FindByID(card.ID)
}

func (s *CardService) Update(id, title, description string, priority int, col model.Column, sortOrder int, reporterID, assigneeID, actingUserID string) (model.Card, error) {
	if err := validateCard(title, priority, col, reporterID, assigneeID); err != nil {
		return model.Card{}, err
	}

	oldCard, err := s.cards.FindByID(id)
	if err != nil {
		return model.Card{}, err
	}

	card, err := s.cards.Update(id, title, description, priority, col, sortOrder, reporterID, assigneeID)
	if err != nil {
		return card, err
	}

	s.activity.Create("updated", "card", card.ID, actingUserID, "card updated", card.ID)

	if oldCard.Column != model.ColumnDone && col == model.ColumnDone {
		s.notifs.Create(card.ReporterID, "Card \""+card.Title+"\" was moved to Done", card.ID)
	}

	return s.cards.FindByID(card.ID)
}

func (s *CardService) Move(id string, col model.Column, actingUserID string) (model.Card, error) {
	if !col.Valid() {
		return model.Card{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "column", Message: "column must be one of: todo, doing, done"}},
		}
	}

	oldCard, err := s.cards.FindByID(id)
	if err != nil {
		return model.Card{}, err
	}

	card, err := s.cards.UpdateColumn(id, col)
	if err != nil {
		return card, err
	}

	details := "moved from " + string(oldCard.Column) + " to " + string(col)
	s.activity.Create("moved", "card", card.ID, actingUserID, details, card.ID)

	if oldCard.Column != model.ColumnDone && col == model.ColumnDone {
		s.notifs.Create(card.ReporterID, "Card \""+card.Title+"\" was moved to Done", card.ID)
	}

	return s.cards.FindByID(card.ID)
}

func (s *CardService) Delete(id, actingUserID string) error {
	_, err := s.cards.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.cards.Delete(id); err != nil {
		return err
	}

	s.activity.Create("deleted", "card", id, actingUserID, "card deleted", "")
	return nil
}

func validateCard(title string, priority int, col model.Column, reporterID, assigneeID string) error {
	var fieldErrors []model.FieldError

	if strings.TrimSpace(title) == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "title", Message: "title is required"})
	}
	if priority < 1 || priority > 5 {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "priority", Message: "priority must be between 1 and 5"})
	}
	if !col.Valid() {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "column", Message: "column must be one of: todo, doing, done"})
	}
	if strings.TrimSpace(reporterID) == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "reporter_id", Message: "reporter_id is required"})
	}
	if strings.TrimSpace(assigneeID) == "" {
		fieldErrors = append(fieldErrors, model.FieldError{Field: "assignee_id", Message: "assignee_id is required"})
	}

	if len(fieldErrors) > 0 {
		return &model.ValidationError{Fields: fieldErrors}
	}
	return nil
}
