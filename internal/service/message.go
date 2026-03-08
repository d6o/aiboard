package service

import (
	"regexp"
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type messageStore interface {
	FindAll(limit int, before string) ([]model.Message, error)
	FindByID(id string) (model.Message, error)
	Create(userID, content string) (model.Message, error)
	Delete(id string) error
	UnreadCount(userID string) (int, error)
	MarkRead(userID string) error
}

type messageUserFinder interface {
	FindAll() ([]model.User, error)
}

type messageNotifier interface {
	Create(userID, message, cardID string) (model.Notification, error)
}

type MessageService struct {
	messages messageStore
	users    messageUserFinder
	notifs   messageNotifier
}

func NewMessageService(messages messageStore, users messageUserFinder, notifs messageNotifier) *MessageService {
	return &MessageService{
		messages: messages,
		users:    users,
		notifs:   notifs,
	}
}

func (s *MessageService) List(limit int, before string) ([]model.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.messages.FindAll(limit, before)
}

func (s *MessageService) Create(userID, content string) (model.Message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return model.Message{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "content", Message: "content is required"}},
		}
	}

	msg, err := s.messages.Create(userID, content)
	if err != nil {
		return msg, err
	}

	s.processMentions(content, userID)

	return msg, nil
}

func (s *MessageService) Delete(id string) error {
	if _, err := s.messages.FindByID(id); err != nil {
		return err
	}
	return s.messages.Delete(id)
}

func (s *MessageService) UnreadCount(userID string) (int, error) {
	return s.messages.UnreadCount(userID)
}

func (s *MessageService) MarkRead(userID string) error {
	return s.messages.MarkRead(userID)
}

func (s *MessageService) processMentions(content, authorID string) {
	mentionRe := regexp.MustCompile(`@(\w+)`)
	matches := mentionRe.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return
	}

	allUsers, err := s.users.FindAll()
	if err != nil {
		return
	}

	userMap := make(map[string]model.User, len(allUsers))
	for _, u := range allUsers {
		userMap[strings.ToLower(u.Name)] = u
	}

	// Find author name for the notification message
	authorName := "Someone"
	for _, u := range allUsers {
		if u.ID == authorID {
			authorName = u.Name
			break
		}
	}

	preview := content
	if len(preview) > 100 {
		preview = preview[:100] + "..."
	}

	notified := make(map[string]bool)
	for _, match := range matches {
		name := strings.ToLower(match[1])
		u, ok := userMap[name]
		if !ok || u.ID == authorID || notified[u.ID] {
			continue
		}
		notified[u.ID] = true
		msg := authorName + " mentioned you in chat: " + preview
		s.notifs.Create(u.ID, msg, "")
	}
}
