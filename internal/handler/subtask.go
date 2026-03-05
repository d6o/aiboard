package handler

import (
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type subtaskService interface {
	List(cardID string) ([]model.Subtask, error)
	Create(cardID, title, actingUserID string) (model.Subtask, error)
	Update(id, title string, completed bool, actingUserID string) (model.Subtask, error)
	Delete(id, actingUserID string) error
	Reorder(cardID string, ids []string, actingUserID string) error
}

type SubtaskHandler struct {
	svc subtaskService
	rw  responseWriter
}

func NewSubtaskHandler(svc subtaskService) *SubtaskHandler {
	return &SubtaskHandler{svc: svc}
}

func (h *SubtaskHandler) List(w http.ResponseWriter, r *http.Request) {
	subtasks, err := h.svc.List(r.PathValue("cardID"))
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if subtasks == nil {
		subtasks = []model.Subtask{}
	}
	h.rw.JSON(w, http.StatusOK, subtasks)
}

func (h *SubtaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title  string `json:"title"`
		UserID string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	st, err := h.svc.Create(r.PathValue("cardID"), req.Title, req.UserID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, st)
}

func (h *SubtaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
		UserID    string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	st, err := h.svc.Update(r.PathValue("id"), req.Title, req.Completed, req.UserID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, st)
}

func (h *SubtaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := h.svc.Delete(r.PathValue("id"), userID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *SubtaskHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs    []string `json:"ids"`
		UserID string   `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	if err := h.svc.Reorder(r.PathValue("cardID"), req.IDs, req.UserID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "reordered"})
}
