package handler

import (
	"net/http"
	"strconv"

	"github.com/d6o/aiboard/internal/model"
)

type cardService interface {
	List(filter model.CardFilter) ([]model.Card, error)
	Get(id string) (model.Card, error)
	Create(title, description string, priority int, col model.Column, reporterID, assigneeID, parentID, actingUserID string) (model.Card, error)
	Update(id, title, description string, priority int, col model.Column, sortOrder int, reporterID, assigneeID, actingUserID string) (model.Card, error)
	Move(id string, col model.Column, actingUserID string) (model.Card, error)
	Delete(id, actingUserID string) error
}

type CardHandler struct {
	svc cardService
	rw  responseWriter
}

func NewCardHandler(svc cardService) *CardHandler {
	return &CardHandler{svc: svc}
}

func (h *CardHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	priority, _ := strconv.Atoi(q.Get("priority"))
	filter := model.CardFilter{
		Column:   q.Get("column"),
		Assignee: q.Get("assignee_id"),
		Reporter: q.Get("reporter_id"),
		Tag:      q.Get("tag_id"),
		Priority: priority,
		ParentID: q.Get("parent_id"),
	}

	cards, err := h.svc.List(filter)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if cards == nil {
		cards = []model.Card{}
	}
	h.rw.JSON(w, http.StatusOK, cards)
}

func (h *CardHandler) Get(w http.ResponseWriter, r *http.Request) {
	card, err := h.svc.Get(r.PathValue("id"))
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, card)
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string       `json:"title"`
		Description string       `json:"description"`
		Priority    int          `json:"priority"`
		Column      model.Column `json:"column"`
		ReporterID  string       `json:"reporter_id"`
		AssigneeID  string       `json:"assignee_id"`
		ParentID    string       `json:"parent_id"`
		UserID      string       `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	actingUser := req.UserID
	if actingUser == "" {
		actingUser = req.ReporterID
	}

	card, err := h.svc.Create(req.Title, req.Description, req.Priority, req.Column, req.ReporterID, req.AssigneeID, req.ParentID, actingUser)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, card)
}

func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string       `json:"title"`
		Description string       `json:"description"`
		Priority    int          `json:"priority"`
		Column      model.Column `json:"column"`
		SortOrder   int          `json:"sort_order"`
		ReporterID  string       `json:"reporter_id"`
		AssigneeID  string       `json:"assignee_id"`
		UserID      string       `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	card, err := h.svc.Update(r.PathValue("id"), req.Title, req.Description, req.Priority, req.Column, req.SortOrder, req.ReporterID, req.AssigneeID, req.UserID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, card)
}

func (h *CardHandler) Move(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Column model.Column `json:"column"`
		UserID string       `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	card, err := h.svc.Move(r.PathValue("id"), req.Column, req.UserID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, card)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := h.svc.Delete(r.PathValue("id"), userID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
