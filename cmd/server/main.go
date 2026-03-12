package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/d6o/aiboard/internal/config"
	"github.com/d6o/aiboard/internal/database"
	"github.com/d6o/aiboard/internal/server"
)

func main() {
	cfg := config.NewConfig()

	conn, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	defer conn.Close()

	migrator := database.NewMigrator(conn.DB)
	if err := migrator.Run(); err != nil {
		log.Fatal("failed to run migrations: ", err)
	}

	seeder := database.NewSeeder(conn.DB)
	if err := seeder.Run(); err != nil {
		log.Fatal("failed to seed database: ", err)
	}

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatal("failed to create upload directory: ", err)
	}

	srv := server.NewServer(conn.DB, cfg.UploadDir)
	srv.StartBackgroundTasks(context.Background())

	log.Println("AIBoard server starting on :" + cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, srv); err != nil {
		log.Fatal("server error: ", err)
	}
}
