package config

import (
	"os"
	"strings"
)

type Config struct {
	Port               string
	DatabaseURL        string
	KeycloakBaseURL    string
	KeycloakRealm      string
	KeycloakClientID   string
	KeycloakAdminUser  string
	KeycloakAdminPass  string
	CookieSecure       bool
	CORSAllowedOrigins []string
	FrontendDir        string
	BasePath           string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://ticketing:ticketing@localhost:5432/ticketing?sslmode=disable"
	}

	keycloakBase := os.Getenv("KEYCLOAK_BASE_URL")
	if keycloakBase == "" {
		keycloakBase = "http://localhost:8081"
	}

	keycloakRealm := os.Getenv("KEYCLOAK_REALM")
	if keycloakRealm == "" {
		keycloakRealm = "ticketing"
	}

	keycloakClient := os.Getenv("KEYCLOAK_CLIENT_ID")
	if keycloakClient == "" {
		keycloakClient = "myclient"
	}

	keycloakAdminUser := os.Getenv("KEYCLOAK_ADMIN_USER")
	if keycloakAdminUser == "" {
		keycloakAdminUser = "admin"
	}

	keycloakAdminPass := os.Getenv("KEYCLOAK_ADMIN_PASSWORD")
	if keycloakAdminPass == "" {
		keycloakAdminPass = "admin"
	}

	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	allowedOrigins := parseCSV(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:5173"}
	}
	frontendDir := os.Getenv("FRONTEND_DIR")

	basePath := os.Getenv("BASE_PATH")
	if basePath == "" {
		basePath = "/"
	}
	// Ensure base path starts with / and doesn't end with /
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	if basePath != "/" && strings.HasSuffix(basePath, "/") {
		basePath = strings.TrimSuffix(basePath, "/")
	}

	return Config{
		Port:               port,
		DatabaseURL:        dbURL,
		KeycloakBaseURL:    keycloakBase,
		KeycloakRealm:      keycloakRealm,
		KeycloakClientID:   keycloakClient,
		KeycloakAdminUser:  keycloakAdminUser,
		KeycloakAdminPass:  keycloakAdminPass,
		CookieSecure:       cookieSecure,
		CORSAllowedOrigins: allowedOrigins,
		FrontendDir:        frontendDir,
		BasePath:           basePath,
	}
}

func parseCSV(input string) []string {
	if input == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(input, ",") {
		value := strings.TrimSpace(part)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
