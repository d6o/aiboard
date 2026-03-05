package server

import (
	"database/sql"
	"net/http"

	"github.com/d6o/aiboard/internal/handler"
	"github.com/d6o/aiboard/internal/service"
	"github.com/d6o/aiboard/internal/store"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer(db *sql.DB) *Server {
	userStore := store.NewUserStore(db)
	cardStore := store.NewCardStore(db)
	subtaskStore := store.NewSubtaskStore(db)
	tagStore := store.NewTagStore(db)
	commentStore := store.NewCommentStore(db)
	notifStore := store.NewNotificationStore(db)
	activityStore := store.NewActivityStore(db)
	idempotencyStore := store.NewIdempotencyStore(db)

	userSvc := service.NewUserService(userStore)
	cardSvc := service.NewCardService(cardStore, tagStore, subtaskStore, commentStore, activityStore, notifStore)
	subtaskSvc := service.NewSubtaskService(subtaskStore, cardStore, activityStore, notifStore)
	tagSvc := service.NewTagService(tagStore, activityStore)
	commentSvc := service.NewCommentService(commentStore, userStore, notifStore, cardStore, activityStore)
	notifSvc := service.NewNotificationService(notifStore)
	activitySvc := service.NewActivityService(activityStore)

	userH := handler.NewUserHandler(userSvc)
	cardH := handler.NewCardHandler(cardSvc)
	subtaskH := handler.NewSubtaskHandler(subtaskSvc)
	tagH := handler.NewTagHandler(tagSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	notifH := handler.NewNotificationHandler(notifSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	idempotency := handler.NewIdempotencyMiddleware(idempotencyStore)

	mux := http.NewServeMux()

	// Users
	mux.HandleFunc("GET /api/users", userH.List)
	mux.HandleFunc("GET /api/users/{id}", userH.Get)
	mux.HandleFunc("POST /api/users", idempotency.Wrap(userH.Create))

	// Cards
	mux.HandleFunc("GET /api/cards", cardH.List)
	mux.HandleFunc("GET /api/cards/{id}", cardH.Get)
	mux.HandleFunc("POST /api/cards", idempotency.Wrap(cardH.Create))
	mux.HandleFunc("PUT /api/cards/{id}", cardH.Update)
	mux.HandleFunc("DELETE /api/cards/{id}", cardH.Delete)
	mux.HandleFunc("PATCH /api/cards/{id}/move", cardH.Move)

	// Subtasks
	mux.HandleFunc("GET /api/cards/{cardID}/subtasks", subtaskH.List)
	mux.HandleFunc("POST /api/cards/{cardID}/subtasks", idempotency.Wrap(subtaskH.Create))
	mux.HandleFunc("PUT /api/cards/{cardID}/subtasks/{id}", subtaskH.Update)
	mux.HandleFunc("DELETE /api/cards/{cardID}/subtasks/{id}", subtaskH.Delete)
	mux.HandleFunc("PATCH /api/cards/{cardID}/subtasks/reorder", subtaskH.Reorder)

	// Tags
	mux.HandleFunc("GET /api/tags", tagH.List)
	mux.HandleFunc("POST /api/tags", idempotency.Wrap(tagH.Create))
	mux.HandleFunc("DELETE /api/tags/{id}", tagH.Delete)
	mux.HandleFunc("POST /api/cards/{cardID}/tags", tagH.AttachToCard)
	mux.HandleFunc("DELETE /api/cards/{cardID}/tags/{tagID}", tagH.DetachFromCard)

	// Comments
	mux.HandleFunc("GET /api/cards/{cardID}/comments", commentH.List)
	mux.HandleFunc("POST /api/cards/{cardID}/comments", idempotency.Wrap(commentH.Create))
	mux.HandleFunc("DELETE /api/cards/{cardID}/comments/{id}", commentH.Delete)

	// Notifications
	mux.HandleFunc("GET /api/users/{userID}/notifications", notifH.List)
	mux.HandleFunc("PATCH /api/users/{userID}/notifications/{id}/read", notifH.MarkRead)
	mux.HandleFunc("PATCH /api/users/{userID}/notifications/read-all", notifH.MarkAllRead)

	// Activity log
	mux.HandleFunc("GET /api/activity", activityH.List)

	// Static files
	mux.Handle("GET /", http.FileServer(http.Dir("static")))

	return &Server{mux: mux}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
