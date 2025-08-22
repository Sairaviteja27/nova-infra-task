package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sairaviteja27/nova-infra-task/server"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env found or failed to load; relying on process env")
	}

	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	}

	srv := server.NewServer(addr)
	log.Printf("Starting server on %s", addr)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
