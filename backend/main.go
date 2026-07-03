package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"frpg-backend/internal/service"
)

func main() {
	// Load backend/.env in local dev. In production the file is absent and this
	// is a no-op — config comes from the platform's environment. Load never
	// overrides variables already set in the environment.
	if err := godotenv.Load(); err == nil {
		log.Println("loaded .env")
	}

	server := service.NewServer(context.Background())

	port := envOr("PORT", "8080")
	log.Printf("backend listening on :%s", port)
	if err := http.ListenAndServe(":"+port, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
