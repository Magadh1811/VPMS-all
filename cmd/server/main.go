package main

import (
	"log"
	"os"

	"Backend-Go/internal/config"
	"Backend-Go/internal/db"
	"Backend-Go/internal/router"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("config error: ", err)
	}

	database, err := db.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("db error: ", err)
	}
	defer database.Close()

	r := router.Setup(database, cfg)

	// Determine port: cfg.Port -> $PORT -> 8080
	port := cfg.Port
	if port == "" {
		if p := os.Getenv("PORT"); p != "" {
			port = p
		} else {
			port = "8080"
		}
	}

	addr := "0.0.0.0:" + port // important for containerized hosts like Railway/Fly
	log.Printf("server starting on %s\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatal("server error: ", err)
	}
}
