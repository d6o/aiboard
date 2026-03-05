package handler

import (
	"bytes"
	"net/http"

	"github.com/d6o/aiboard/internal/store"
)

type IdempotencyMiddleware struct {
	store *store.IdempotencyStore
	rw    responseWriter
}

func NewIdempotencyMiddleware(s *store.IdempotencyStore) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{store: s}
}

func (m *IdempotencyMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next(w, r)
			return
		}

		record, found, err := m.store.Find(key)
		if err != nil {
			m.rw.HandleError(w, err)
			return
		}
		if found {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(record.ResponseStatus)
			w.Write(record.ResponseBody)
			return
		}

		rec := &responseRecorder{ResponseWriter: w, body: &bytes.Buffer{}}
		next(rec, r)

		m.store.Save(key, rec.statusCode, rec.body.Bytes())
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
