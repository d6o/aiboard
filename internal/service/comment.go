package service

import (
	"regexp"
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type commentStore interface {
	FindByCardID(cardID string) ([]model.Comment, error)
	Create(cardID, userID, content string) (model.Comment, error)
	Delete(id string) error
	FindByID(id string) (model.Comment, error)
}

type commentUserFinder interface {
	FindAll() ([]model.User, error)
}

type commentNotifier interface {
	Create(userID, message, cardID string) (model.Notification, error)
}

type commentCardFinder interface {
	FindByID(id string) (model.Card, error)
}

type commentActivityLogger interface {
	Create(action, resourceType, resourceID, userID, details, cardID string) (model.ActivityEntry, error)
}

type CommentService struct {
	comments commentStore
	users    commentUserFinder
	notifs   commentNotifier
	cards    commentCardFinder
	activity commentActivityLogger
}

func NewCommentService(comments commentStore, users commentUserFinder, notifs commentNotifier, cards commentCardFinder, activity commentActivityLogger) *CommentService {
	return &CommentService{
		comments: comments,
		users:    users,
		notifs:   notifs,
		cards:    cards,
		activity: activity,
	}
}

func (s *CommentService) List(cardID string) ([]model.Comment, error) {
	return s.comments.FindByCardID(cardID)
}

func (s *CommentService) Create(cardID, userID, content, actingUserID string) (model.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return model.Comment{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "content", Message: "content is required"}},
		}
	}

	if _, err := s.cards.FindByID(cardID); err != nil {
		return model.Comment{}, err
	}

	comment, err := s.comments.Create(cardID, userID, content)
	if err != nil {
		return comment, err
	}

	s.activity.Create("created", "comment", comment.ID, actingUserID, "comment added", cardID)

	s.processMentions(content, userID, cardID)

	return comment, nil
}

func (s *CommentService) Delete(id, actingUserID string) error {
	comment, err := s.comments.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.comments.Delete(id); err != nil {
		return err
	}

	s.activity.Create("deleted", "comment", id, actingUserID, "comment deleted", comment.CardID)
	return nil
}

func (s *CommentService) processMentions(content, authorID, cardID string) {
	mentionRe := regexp.MustCompile(`@(\w+)`)
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return
	}

	users, err := s.users.FindAll()
	if err != nil {
		return
	}

	userMap := make(map[string]model.User, len(users))
	for _, u := range users {
		userMap[strings.ToLower(u.Name)] = u
	}

	notified := make(map[string]bool)
	card, _ := s.cards.FindByID(cardID)
	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	for _, match := range matches {
		name := strings.ToLower(match[1])
		u, ok := userMap[name]
		if !ok || u.ID == authorID || notified[u.ID] {
			continue
		}
		notified[u.ID] = true
		msg := "You were mentioned in a comment on card \"" + card.Title + "\": " + preview
		s.notifs.Create(u.ID, msg, cardID)
	}
}
