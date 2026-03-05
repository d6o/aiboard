package service

import "github.com/d6o/aiboard/internal/model"

type notificationStore interface {
	FindByUserID(userID string, unreadOnly bool) ([]model.Notification, error)
	Create(userID, message, cardID string) (model.Notification, error)
	MarkRead(id string) error
	MarkAllRead(userID string) error
}

type NotificationService struct {
	notifs notificationStore
}

func NewNotificationService(notifs notificationStore) *NotificationService {
	return &NotificationService{notifs: notifs}
}

func (s *NotificationService) List(userID string, unreadOnly bool) ([]model.Notification, error) {
	return s.notifs.FindByUserID(userID, unreadOnly)
}

func (s *NotificationService) MarkRead(id string) error {
	return s.notifs.MarkRead(id)
}

func (s *NotificationService) MarkAllRead(userID string) error {
	return s.notifs.MarkAllRead(userID)
}
