package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ticketing-system/backend/internal/auth"
	"ticketing-system/backend/internal/config"
	"ticketing-system/backend/internal/httpapi"
	"ticketing-system/backend/internal/migrate"
	"ticketing-system/backend/internal/store"
	"ticketing-system/backend/internal/webhook"
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

	migrationsDir := resolveMigrationsDir()
	if err := migrate.Apply(ctx, st.DB(), migrationsDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	authClient := auth.New(auth.Config{
		BaseURL:  cfg.KeycloakBaseURL,
		Realm:    cfg.KeycloakRealm,
		ClientID: cfg.KeycloakClientID,
		Username: cfg.KeycloakAdminUser,
		Password: cfg.KeycloakAdminPass,
	})

	dispatcher := webhook.New(st)

	handler := httpapi.NewHandler(st, authClient, dispatcher, httpapi.HandlerOptions{
		CookieName:     "ticketing_session",
		CookieSecure:   cfg.CookieSecure,
		AllowedOrigins: cfg.CORSAllowedOrigins,
	})
	router := httpapi.Router(handler)
	apiHandler := http.Handler(router)

	frontendDir := cfg.FrontendDir
	if frontendDir == "" {
		frontendDir = findFrontendDir([]string{"frontend/dist", "../frontend/dist"})
	}
	if frontendDir != "" {
		apiHandler = httpapi.WithFrontend(apiHandler, frontendDir, cfg.BasePath)
	}

	// Mount the API at the base path
	finalHandler := httpapi.WithBasePath(apiHandler, cfg.BasePath)

	addr := ":" + cfg.Port
	log.Printf("ticketing-system api listening on %s (base path: %s)", addr, cfg.BasePath)
	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func resolveMigrationsDir() string {
	candidates := []string{
		"backend/migrations",
		"migrations",
		filepath.Join("..", "migrations"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return "backend/migrations"
}

func findFrontendDir(candidates []string) string {
	for _, candidate := range candidates {
		indexPath := filepath.Join(candidate, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return candidate
		}
	}
	return ""
}
