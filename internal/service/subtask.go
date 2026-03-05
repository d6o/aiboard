package service

import (
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type subtaskStore interface {
	FindByCardID(cardID string) ([]model.Subtask, error)
	FindByID(id string) (model.Subtask, error)
	Create(cardID, title string, sortOrder int) (model.Subtask, error)
	Update(id, title string, completed bool) (model.Subtask, error)
	Delete(id string) error
	CountByCardID(cardID string) (int, error)
	CountIncompleteByCardID(cardID string) (int, error)
	HasDuplicateTitle(cardID, title, excludeID string) (bool, error)
	Reorder(cardID string, ids []string) error
	NextSortOrder(cardID string) (int, error)
}

type subtaskCardFinder interface {
	FindByID(id string) (model.Card, error)
}

type subtaskActivityLogger interface {
	Create(action, resourceType, resourceID, userID, details, cardID string) (model.ActivityEntry, error)
}

type subtaskNotifier interface {
	Create(userID, message, cardID string) (model.Notification, error)
}

type SubtaskService struct {
	subtasks subtaskStore
	cards    subtaskCardFinder
	activity subtaskActivityLogger
	notifs   subtaskNotifier
}

func NewSubtaskService(subtasks subtaskStore, cards subtaskCardFinder, activity subtaskActivityLogger, notifs subtaskNotifier) *SubtaskService {
	return &SubtaskService{
		subtasks: subtasks,
		cards:    cards,
		activity: activity,
		notifs:   notifs,
	}
}

func (s *SubtaskService) List(cardID string) ([]model.Subtask, error) {
	return s.subtasks.FindByCardID(cardID)
}

func (s *SubtaskService) Create(cardID, title, actingUserID string) (model.Subtask, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Subtask{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "title", Message: "title is required"}},
		}
	}

	if _, err := s.cards.FindByID(cardID); err != nil {
		return model.Subtask{}, err
	}

	count, err := s.subtasks.CountByCardID(cardID)
	if err != nil {
		return model.Subtask{}, err
	}
	if count >= 20 {
		return model.Subtask{}, model.ErrSubtaskLimit
	}

	dup, err := s.subtasks.HasDuplicateTitle(cardID, title, "")
	if err != nil {
		return model.Subtask{}, err
	}
	if dup {
		return model.Subtask{}, model.ErrDuplicateSubtaskName
	}

	sortOrder, err := s.subtasks.NextSortOrder(cardID)
	if err != nil {
		return model.Subtask{}, err
	}

	st, err := s.subtasks.Create(cardID, title, sortOrder)
	if err != nil {
		return st, err
	}

	s.activity.Create("created", "subtask", st.ID, actingUserID, "subtask created: "+title, cardID)
	return st, nil
}

func (s *SubtaskService) Update(id, title string, completed bool, actingUserID string) (model.Subtask, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return model.Subtask{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "title", Message: "title is required"}},
		}
	}

	existing, err := s.subtasks.FindByID(id)
	if err != nil {
		return model.Subtask{}, err
	}

	dup, err := s.subtasks.HasDuplicateTitle(existing.CardID, title, id)
	if err != nil {
		return model.Subtask{}, err
	}
	if dup {
		return model.Subtask{}, model.ErrDuplicateSubtaskName
	}

	st, err := s.subtasks.Update(id, title, completed)
	if err != nil {
		return st, err
	}

	s.activity.Create("updated", "subtask", st.ID, actingUserID, "subtask updated: "+title, st.CardID)

	if completed && !existing.Completed {
		incomplete, err := s.subtasks.CountIncompleteByCardID(st.CardID)
		if err != nil {
			return st, nil
		}
		if incomplete == 0 {
			total, err := s.subtasks.CountByCardID(st.CardID)
			if err != nil {
				return st, nil
			}
			if total > 0 {
				card, err := s.cards.FindByID(st.CardID)
				if err != nil {
					return st, nil
				}
				s.notifs.Create(card.AssigneeID, "All subtasks completed on card \""+card.Title+"\"", card.ID)
			}
		}
	}

	return st, nil
}

func (s *SubtaskService) Delete(id, actingUserID string) error {
	st, err := s.subtasks.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.subtasks.Delete(id); err != nil {
		return err
	}

	s.activity.Create("deleted", "subtask", id, actingUserID, "subtask deleted: "+st.Title, st.CardID)
	return nil
}

func (s *SubtaskService) Reorder(cardID string, ids []string, actingUserID string) error {
	if _, err := s.cards.FindByID(cardID); err != nil {
		return err
	}

	if err := s.subtasks.Reorder(cardID, ids); err != nil {
		return err
	}

	s.activity.Create("reordered", "subtask", cardID, actingUserID, "subtasks reordered", cardID)
	return nil
}
