package main

import (
	"log"

	"github.com/sajadHazrati2000/kei/backend/internal/config"
	"github.com/sajadHazrati2000/kei/backend/internal/db"
	"github.com/sajadHazrati2000/kei/backend/internal/server"
)

func main() {
	cfg := config.Load()

	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db.Connect: %v", err)
	}
	defer pool.Close()

	srv := server.New(cfg, pool)
	log.Printf("kei listening on :%s", cfg.Port)
	if err := srv.Run(); err != nil {
		log.Fatalf("server.Run: %v", err)
	}
}
