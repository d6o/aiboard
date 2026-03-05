package handler

import (
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type tagService interface {
	List() ([]model.Tag, error)
	Create(name, color, actingUserID string) (model.Tag, error)
	Delete(id, actingUserID string) error
	AttachToCard(cardID, tagID, actingUserID string) error
	DetachFromCard(cardID, tagID, actingUserID string) error
}

type TagHandler struct {
	svc tagService
	rw  responseWriter
}

func NewTagHandler(svc tagService) *TagHandler {
	return &TagHandler{svc: svc}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.svc.List()
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if tags == nil {
		tags = []model.Tag{}
	}
	h.rw.JSON(w, http.StatusOK, tags)
}

func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name   string `json:"name"`
		Color  string `json:"color"`
		UserID string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	tag, err := h.svc.Create(req.Name, req.Color, req.UserID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, tag)
}

func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := h.svc.Delete(r.PathValue("id"), userID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *TagHandler) AttachToCard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TagID  string `json:"tag_id"`
		UserID string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := h.svc.AttachToCard(r.PathValue("cardID"), req.TagID, req.UserID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "attached"})
}

func (h *TagHandler) DetachFromCard(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := h.svc.DetachFromCard(r.PathValue("cardID"), r.PathValue("tagID"), userID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "detached"})
}
