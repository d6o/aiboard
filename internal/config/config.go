package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
	UploadDir   string
}

func NewConfig() Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/aiboard?sslmode=disable"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}

	return Config{
		DatabaseURL: dbURL,
		Port:        port,
		UploadDir:   uploadDir,
	}
}
