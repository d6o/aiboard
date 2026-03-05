package handler

import (
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type activityService interface {
	List(filter model.ActivityFilter) ([]model.ActivityEntry, error)
}

type ActivityHandler struct {
	svc activityService
	rw  responseWriter
}

func NewActivityHandler(svc activityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.ActivityFilter{
		CardID: q.Get("card_id"),
		UserID: q.Get("user_id"),
		Action: q.Get("action"),
	}

	entries, err := h.svc.List(filter)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if entries == nil {
		entries = []model.ActivityEntry{}
	}
	h.rw.JSON(w, http.StatusOK, entries)
}
