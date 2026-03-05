package handler

import (
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type userService interface {
	List() ([]model.User, error)
	Get(id string) (model.User, error)
	Create(name, avatarColor string) (model.User, error)
	Delete(id string) error
}

type UserHandler struct {
	svc userService
	rw  responseWriter
}

func NewUserHandler(svc userService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List()
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if users == nil {
		users = []model.User{}
	}
	h.rw.JSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.svc.Get(id)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		AvatarColor string `json:"avatar_color"`
	}
	if err := decodeJSON(r, &req); err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body", nil)
		return
	}
	user, err := h.svc.Create(req.Name, req.AvatarColor)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.PathValue("id")); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
