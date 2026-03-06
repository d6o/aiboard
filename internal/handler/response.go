package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/d6o/aiboard/internal/model"
)

type apiError struct {
	Code    string             `json:"code"`
	Message string             `json:"message"`
	Fields  []model.FieldError `json:"fields,omitempty"`
}

type apiResponse struct {
	Data  any       `json:"data,omitempty"`
	Error *apiError `json:"error,omitempty"`
}

type responseWriter struct{}

func (responseWriter) JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{Data: data})
}

func (responseWriter) Error(w http.ResponseWriter, status int, code, message string, fields []model.FieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(apiResponse{
		Error: &apiError{Code: code, Message: message, Fields: fields},
	})
}

func (rw responseWriter) HandleError(w http.ResponseWriter, err error) {
	var ve *model.ValidationError
	if errors.As(err, &ve) {
		rw.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", ve.Error(), ve.Fields)
		return
	}

	if errors.Is(err, model.ErrNotFound) {
		rw.Error(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
		return
	}

	if errors.Is(err, model.ErrDuplicateName) {
		rw.Error(w, http.StatusConflict, "DUPLICATE_NAME", "name already exists", nil)
		return
	}

	if errors.Is(err, model.ErrTagAlreadyAttached) {
		rw.Error(w, http.StatusConflict, "TAG_ALREADY_ATTACHED", "tag is already attached to this card", nil)
		return
	}

	if errors.Is(err, model.ErrUserInUse) {
		rw.Error(w, http.StatusConflict, "USER_IN_USE", "user is referenced by existing cards and cannot be deleted", nil)
		return
	}

	if errors.Is(err, model.ErrChildrenNotDone) {
		rw.Error(w, http.StatusConflict, "CHILDREN_NOT_DONE", "all child cards must be in done before this card can move to done", nil)
		return
	}

	rw.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an internal error occurred", nil)
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}
