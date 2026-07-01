package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"frpg-backend/internal/service"
)

func main() {
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
