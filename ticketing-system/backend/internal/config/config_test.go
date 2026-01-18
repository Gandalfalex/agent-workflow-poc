package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear all relevant env vars
	envVars := []string{
		"PORT", "DATABASE_URL", "KEYCLOAK_BASE_URL", "KEYCLOAK_REALM",
		"KEYCLOAK_CLIENT_ID", "COOKIE_SECURE", "CORS_ALLOWED_ORIGINS", "FRONTEND_DIR",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port '8080', got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://ticketing:ticketing@localhost:5432/ticketing?sslmode=disable" {
		t.Errorf("expected default database URL, got %q", cfg.DatabaseURL)
	}
	if cfg.KeycloakBaseURL != "http://localhost:8081" {
		t.Errorf("expected default keycloak base URL, got %q", cfg.KeycloakBaseURL)
	}
	if cfg.KeycloakRealm != "ticketing" {
		t.Errorf("expected default keycloak realm 'ticketing', got %q", cfg.KeycloakRealm)
	}
	if cfg.KeycloakClientID != "myclient" {
		t.Errorf("expected default keycloak client ID 'myclient', got %q", cfg.KeycloakClientID)
	}
	if cfg.CookieSecure != false {
		t.Errorf("expected default cookie secure false, got %v", cfg.CookieSecure)
	}
	if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:5173" {
		t.Errorf("expected default CORS origins, got %v", cfg.CORSAllowedOrigins)
	}
	if cfg.FrontendDir != "" {
		t.Errorf("expected empty frontend dir, got %q", cfg.FrontendDir)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	// Set custom values
	os.Setenv("PORT", "3000")
	os.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/mydb")
	os.Setenv("KEYCLOAK_BASE_URL", "https://auth.example.com")
	os.Setenv("KEYCLOAK_REALM", "myrealm")
	os.Setenv("KEYCLOAK_CLIENT_ID", "myapp")
	os.Setenv("COOKIE_SECURE", "true")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com,https://admin.example.com")
	os.Setenv("FRONTEND_DIR", "/var/www/html")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("KEYCLOAK_BASE_URL")
		os.Unsetenv("KEYCLOAK_REALM")
		os.Unsetenv("KEYCLOAK_CLIENT_ID")
		os.Unsetenv("COOKIE_SECURE")
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		os.Unsetenv("FRONTEND_DIR")
	}()

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("expected port '3000', got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://user:pass@db:5432/mydb" {
		t.Errorf("expected custom database URL, got %q", cfg.DatabaseURL)
	}
	if cfg.KeycloakBaseURL != "https://auth.example.com" {
		t.Errorf("expected custom keycloak base URL, got %q", cfg.KeycloakBaseURL)
	}
	if cfg.KeycloakRealm != "myrealm" {
		t.Errorf("expected realm 'myrealm', got %q", cfg.KeycloakRealm)
	}
	if cfg.KeycloakClientID != "myapp" {
		t.Errorf("expected client ID 'myapp', got %q", cfg.KeycloakClientID)
	}
	if cfg.CookieSecure != true {
		t.Errorf("expected cookie secure true, got %v", cfg.CookieSecure)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Errorf("expected 2 CORS origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.CORSAllowedOrigins[0] != "https://app.example.com" {
		t.Errorf("expected first origin 'https://app.example.com', got %q", cfg.CORSAllowedOrigins[0])
	}
	if cfg.CORSAllowedOrigins[1] != "https://admin.example.com" {
		t.Errorf("expected second origin 'https://admin.example.com', got %q", cfg.CORSAllowedOrigins[1])
	}
	if cfg.FrontendDir != "/var/www/html" {
		t.Errorf("expected frontend dir '/var/www/html', got %q", cfg.FrontendDir)
	}
}

func TestLoad_CookieSecure_OnlyTrueIsTrue(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"TRUE", false}, // only lowercase "true" works
		{"1", false},
		{"yes", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run("COOKIE_SECURE="+tt.input, func(t *testing.T) {
			if tt.input == "" {
				os.Unsetenv("COOKIE_SECURE")
			} else {
				os.Setenv("COOKIE_SECURE", tt.input)
			}
			defer os.Unsetenv("COOKIE_SECURE")

			cfg := Load()
			if cfg.CookieSecure != tt.expected {
				t.Errorf("COOKIE_SECURE=%q: expected %v, got %v", tt.input, tt.expected, cfg.CookieSecure)
			}
		})
	}
}

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single value",
			input:    "http://localhost:3000",
			expected: []string{"http://localhost:3000"},
		},
		{
			name:     "multiple values",
			input:    "http://a.com,http://b.com,http://c.com",
			expected: []string{"http://a.com", "http://b.com", "http://c.com"},
		},
		{
			name:     "values with spaces",
			input:    "  http://a.com  ,  http://b.com  ",
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "trailing comma",
			input:    "http://a.com,http://b.com,",
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "leading comma",
			input:    ",http://a.com,http://b.com",
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "multiple commas",
			input:    "http://a.com,,,,http://b.com",
			expected: []string{"http://a.com", "http://b.com"},
		},
		{
			name:     "only commas",
			input:    ",,,",
			expected: nil,
		},
		{
			name:     "whitespace only values",
			input:    "  ,  ,  ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCSV(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}
