package handler

import (
	"net/http"
	"strconv"

	"github.com/d6o/aiboard/internal/model"
)

type standupService interface {
	GetConfig() (model.StandupConfig, error)
	UpdateConfig(intervalHours int, enabled bool) (model.StandupConfig, error)
	List(limit int) ([]model.Standup, error)
	Get(id string) (model.Standup, error)
	PostEntry(standupID, userID, content string) (model.StandupEntry, error)
}

type StandupHandler struct {
	svc standupService
	rw  responseWriter
}

func NewStandupHandler(svc standupService) *StandupHandler {
	return &StandupHandler{svc: svc}
}

func (h *StandupHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, cfg)
}

func (h *StandupHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IntervalHours int  `json:"interval_hours"`
		Enabled       bool `json:"enabled"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	cfg, err := h.svc.UpdateConfig(req.IntervalHours, req.Enabled)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, cfg)
}

func (h *StandupHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	standups, err := h.svc.List(limit)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if standups == nil {
		standups = []model.Standup{}
	}
	h.rw.JSON(w, http.StatusOK, standups)
}

func (h *StandupHandler) Get(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.Get(r.PathValue("id"))
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, st)
}

func (h *StandupHandler) PostEntry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
		UserID  string `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}

	entry, err := h.svc.PostEntry(r.PathValue("id"), req.UserID, req.Content)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, entry)
}
