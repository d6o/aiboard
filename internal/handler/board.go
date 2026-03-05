package handler

import "net/http"

type boardService interface {
	Reset() error
}

type BoardHandler struct {
	svc boardService
	rw  responseWriter
}

func NewBoardHandler(svc boardService) *BoardHandler {
	return &BoardHandler{svc: svc}
}

func (h *BoardHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Reset(); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "reset"})
}
