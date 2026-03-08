package handler

import (
	"net/http"
	"strconv"

	"github.com/d6o/aiboard/internal/model"
)

type messageService interface {
	List(limit int, before string) ([]model.Message, error)
	Create(userID, content string) (model.Message, error)
	Delete(id string) error
	UnreadCount(userID string) (int, error)
	MarkRead(userID string) error
}

type MessageHandler struct {
	svc messageService
	rw  responseWriter
}

func NewMessageHandler(svc messageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	before := q.Get("before")

	messages, err := h.svc.List(limit, before)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if messages == nil {
		messages = []model.Message{}
	}
	h.rw.JSON(w, http.StatusOK, messages)
}

func (h *MessageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
		UserID  string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	msg, err := h.svc.Create(req.UserID, req.Content)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, msg)
}

func (h *MessageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.PathValue("id")); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *MessageHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	count, err := h.svc.UnreadCount(userID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

func (h *MessageHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := h.svc.MarkRead(req.UserID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
