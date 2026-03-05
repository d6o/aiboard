package handler

import (
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type notificationService interface {
	List(userID string, unreadOnly bool) ([]model.Notification, error)
	MarkRead(id string) error
	MarkAllRead(userID string) error
}

type NotificationHandler struct {
	svc notificationService
	rw  responseWriter
}

func NewNotificationHandler(svc notificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	unreadOnly := r.URL.Query().Get("unread") == "true"
	notifs, err := h.svc.List(r.PathValue("userID"), unreadOnly)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if notifs == nil {
		notifs = []model.Notification{}
	}
	h.rw.JSON(w, http.StatusOK, notifs)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.MarkRead(r.PathValue("id")); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (h *NotificationHandler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.MarkAllRead(r.PathValue("userID")); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "all_read"})
}
