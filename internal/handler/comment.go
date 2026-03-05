package handler

import (
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type commentService interface {
	List(cardID string) ([]model.Comment, error)
	Create(cardID, userID, content, actingUserID string) (model.Comment, error)
	Delete(id, actingUserID string) error
}

type CommentHandler struct {
	svc commentService
	rw  responseWriter
}

func NewCommentHandler(svc commentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	comments, err := h.svc.List(r.PathValue("cardID"))
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if comments == nil {
		comments = []model.Comment{}
	}
	h.rw.JSON(w, http.StatusOK, comments)
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
		UserID  string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	comment, err := h.svc.Create(r.PathValue("cardID"), req.UserID, req.Content, req.UserID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := h.svc.Delete(r.PathValue("id"), userID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
