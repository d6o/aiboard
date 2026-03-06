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

func NewServer(db *sql.DB, uploadDir string) *Server {
	userStore := store.NewUserStore(db)
	cardStore := store.NewCardStore(db)
	tagStore := store.NewTagStore(db)
	commentStore := store.NewCommentStore(db)
	notifStore := store.NewNotificationStore(db)
	activityStore := store.NewActivityStore(db)
	fileStore := store.NewFileStore(db)
	boardStore := store.NewBoardStore(db)
	idempotencyStore := store.NewIdempotencyStore(db)

	userSvc := service.NewUserService(userStore)
	cardSvc := service.NewCardService(cardStore, tagStore, commentStore, activityStore, notifStore)
	tagSvc := service.NewTagService(tagStore, activityStore)
	commentSvc := service.NewCommentService(commentStore, userStore, notifStore, cardStore, activityStore)
	notifSvc := service.NewNotificationService(notifStore)
	activitySvc := service.NewActivityService(activityStore)
	fileSvc := service.NewFileService(fileStore, cardStore, activityStore, commentStore, uploadDir)
	boardSvc := service.NewBoardService(boardStore)

	userH := handler.NewUserHandler(userSvc)
	cardH := handler.NewCardHandler(cardSvc)
	tagH := handler.NewTagHandler(tagSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	notifH := handler.NewNotificationHandler(notifSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	fileH := handler.NewFileHandler(fileSvc)
	boardH := handler.NewBoardHandler(boardSvc)
	idempotency := handler.NewIdempotencyMiddleware(idempotencyStore)

	mux := http.NewServeMux()

	// Users
	mux.HandleFunc("GET /api/users", userH.List)
	mux.HandleFunc("GET /api/users/{id}", userH.Get)
	mux.HandleFunc("POST /api/users", idempotency.Wrap(userH.Create))
	mux.HandleFunc("DELETE /api/users/{id}", userH.Delete)

	// Cards
	mux.HandleFunc("GET /api/cards", cardH.List)
	mux.HandleFunc("GET /api/cards/{id}", cardH.Get)
	mux.HandleFunc("POST /api/cards", idempotency.Wrap(cardH.Create))
	mux.HandleFunc("PUT /api/cards/{id}", cardH.Update)
	mux.HandleFunc("DELETE /api/cards/{id}", cardH.Delete)
	mux.HandleFunc("PATCH /api/cards/{id}/move", cardH.Move)

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

	// Files
	mux.HandleFunc("GET /api/cards/{cardID}/files", fileH.List)
	mux.HandleFunc("POST /api/cards/{cardID}/files", fileH.Upload)
	mux.HandleFunc("GET /api/files/{id}", fileH.Get)
	mux.HandleFunc("GET /api/files/{id}/raw", fileH.Raw)
	mux.HandleFunc("DELETE /api/files/{id}", fileH.Delete)

	// Notifications
	mux.HandleFunc("GET /api/users/{userID}/notifications", notifH.List)
	mux.HandleFunc("PATCH /api/users/{userID}/notifications/{id}/read", notifH.MarkRead)
	mux.HandleFunc("PATCH /api/users/{userID}/notifications/read-all", notifH.MarkAllRead)

	// Activity log
	mux.HandleFunc("GET /api/activity", activityH.List)

	// Board management
	mux.HandleFunc("POST /api/board/reset", boardH.Reset)

	// Static files
	mux.Handle("GET /", http.FileServer(http.Dir("static")))

	return &Server{mux: mux}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
