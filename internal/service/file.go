package service

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/d6o/aiboard/internal/model"
)

type fileStore interface {
	FindByCardID(cardID string) ([]model.File, error)
	FindByID(id string) (model.File, error)
	Create(cardID, filename, contentType string, size int64) (model.File, error)
	Delete(id string) error
}

type fileCardFinder interface {
	FindByID(id string) (model.Card, error)
}

type fileActivityLogger interface {
	Create(action, resourceType, resourceID, userID, details, cardID string) (model.ActivityEntry, error)
}

type fileCommentCreator interface {
	Create(cardID, userID, content string) (model.Comment, error)
}

type FileService struct {
	files    fileStore
	cards    fileCardFinder
	activity fileActivityLogger
	comments fileCommentCreator
	dir      string
}

func NewFileService(files fileStore, cards fileCardFinder, activity fileActivityLogger, comments fileCommentCreator, uploadDir string) *FileService {
	return &FileService{
		files:    files,
		cards:    cards,
		activity: activity,
		comments: comments,
		dir:      uploadDir,
	}
}

func (s *FileService) List(cardID string) ([]model.File, error) {
	files, err := s.files.FindByCardID(cardID)
	if err != nil {
		return nil, err
	}
	for i := range files {
		files[i].RawURL = "/api/files/" + files[i].ID + "/raw"
	}
	return files, nil
}

func (s *FileService) Get(id string) (model.File, error) {
	f, err := s.files.FindByID(id)
	if err != nil {
		return f, err
	}
	f.RawURL = "/api/files/" + f.ID + "/raw"
	return f, nil
}

func (s *FileService) Upload(cardID, filename, contentType string, size int64, body io.Reader, actingUserID string) (model.File, error) {
	filename = sanitizeFilename(filename)
	if strings.TrimSpace(filename) == "" {
		return model.File{}, &model.ValidationError{
			Fields: []model.FieldError{{Field: "file", Message: "filename is required"}},
		}
	}

	if _, err := s.cards.FindByID(cardID); err != nil {
		return model.File{}, err
	}

	f, err := s.files.Create(cardID, filename, contentType, size)
	if err != nil {
		return f, err
	}

	diskPath := filepath.Join(s.dir, f.ID)
	dst, err := os.Create(diskPath)
	if err != nil {
		s.files.Delete(f.ID)
		return model.File{}, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, body); err != nil {
		dst.Close()
		os.Remove(diskPath)
		s.files.Delete(f.ID)
		return model.File{}, err
	}

	f.RawURL = "/api/files/" + f.ID + "/raw"
	s.activity.Create("uploaded", "file", f.ID, actingUserID, "file uploaded: "+filename, cardID)
	s.comments.Create(cardID, actingUserID, "Uploaded file "+filename)
	return f, nil
}

func (s *FileService) Delete(id, actingUserID string) error {
	f, err := s.files.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.files.Delete(id); err != nil {
		return err
	}

	os.Remove(filepath.Join(s.dir, id))
	s.activity.Create("deleted", "file", id, actingUserID, "file deleted: "+f.Filename, f.CardID)
	return nil
}

func (s *FileService) FilePath(id string) (string, error) {
	if _, err := s.files.FindByID(id); err != nil {
		return "", err
	}
	return filepath.Join(s.dir, id), nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == '\x00' {
			return '_'
		}
		return r
	}, name)
	return name
}
