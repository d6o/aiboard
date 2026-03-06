package handler

import (
	"io"
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type fileService interface {
	List(cardID string) ([]model.File, error)
	Get(id string) (model.File, error)
	Upload(cardID, filename, contentType string, size int64, body io.Reader, actingUserID string) (model.File, error)
	Delete(id, actingUserID string) error
	FilePath(id string) (string, error)
}

type FileHandler struct {
	svc fileService
	rw  responseWriter
}

func NewFileHandler(svc fileService) *FileHandler {
	return &FileHandler{svc: svc}
}

func (h *FileHandler) List(w http.ResponseWriter, r *http.Request) {
	files, err := h.svc.List(r.PathValue("cardID"))
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	if files == nil {
		files = []model.File{}
	}
	h.rw.JSON(w, http.StatusOK, files)
}

func (h *FileHandler) Get(w http.ResponseWriter, r *http.Request) {
	f, err := h.svc.Get(r.PathValue("id"))
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, f)
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	const maxUpload = 10 << 20 // 10 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	file, header, err := r.FormFile("file")
	if err != nil {
		h.rw.Error(w, http.StatusBadRequest, "INVALID_FILE", "file is required (multipart field 'file')", nil)
		return
	}
	defer file.Close()

	userID := r.FormValue("user_id")
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	f, err := h.svc.Upload(r.PathValue("cardID"), header.Filename, contentType, header.Size, file, userID)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusCreated, f)
}

func (h *FileHandler) Raw(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f, err := h.svc.Get(id)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}

	diskPath, err := h.svc.FilePath(id)
	if err != nil {
		h.rw.HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+f.Filename+"\"")
	http.ServeFile(w, r, diskPath)
}

func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if err := h.svc.Delete(r.PathValue("id"), userID); err != nil {
		h.rw.HandleError(w, err)
		return
	}
	h.rw.JSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
