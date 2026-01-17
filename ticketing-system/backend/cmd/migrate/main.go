package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"ticketing-system/backend/internal/config"
	"ticketing-system/backend/internal/migrate"
	"ticketing-system/backend/internal/store"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer st.Close()

	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	if err := migrate.Apply(ctx, st.DB(), migrationsDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	log.Print("migrations applied")
}

func resolveMigrationsDir() (string, error) {
	candidates := []string{
		"backend/migrations",
		"migrations",
		filepath.Join("..", "migrations"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", os.ErrNotExist
}
