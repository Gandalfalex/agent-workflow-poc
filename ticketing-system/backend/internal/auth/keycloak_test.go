package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNew(t *testing.T) {
	t.Run("default timeout", func(t *testing.T) {
		client := New(Config{
			BaseURL:  "http://localhost:8081",
			Realm:    "test",
			ClientID: "testclient",
		})

		if client.cfg.BaseURL != "http://localhost:8081" {
			t.Errorf("expected base URL, got %q", client.cfg.BaseURL)
		}
		if client.cfg.Realm != "test" {
			t.Errorf("expected realm 'test', got %q", client.cfg.Realm)
		}
		if client.cfg.ClientID != "testclient" {
			t.Errorf("expected client ID 'testclient', got %q", client.cfg.ClientID)
		}
		if client.httpClient.Timeout != 10*time.Second {
			t.Errorf("expected default timeout 10s, got %v", client.httpClient.Timeout)
		}
	})

	t.Run("custom timeout", func(t *testing.T) {
		client := New(Config{
			BaseURL:  "http://localhost:8081",
			Realm:    "test",
			ClientID: "testclient",
			Timeout:  5 * time.Second,
		})

		if client.httpClient.Timeout != 5*time.Second {
			t.Errorf("expected custom timeout 5s, got %v", client.httpClient.Timeout)
		}
	})
}

func TestTokenURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		realm    string
		expected string
	}{
		{
			name:     "basic",
			baseURL:  "http://localhost:8081",
			realm:    "myrealm",
			expected: "http://localhost:8081/realms/myrealm/protocol/openid-connect/token",
		},
		{
			name:     "with trailing slash",
			baseURL:  "http://localhost:8081/",
			realm:    "myrealm",
			expected: "http://localhost:8081/realms/myrealm/protocol/openid-connect/token",
		},
		{
			name:     "https",
			baseURL:  "https://auth.example.com",
			realm:    "prod",
			expected: "https://auth.example.com/realms/prod/protocol/openid-connect/token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(Config{BaseURL: tt.baseURL, Realm: tt.realm})
			result := client.tokenURL()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestJwksURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		realm    string
		expected string
	}{
		{
			name:     "basic",
			baseURL:  "http://localhost:8081",
			realm:    "myrealm",
			expected: "http://localhost:8081/realms/myrealm/protocol/openid-connect/certs",
		},
		{
			name:     "with trailing slash",
			baseURL:  "http://localhost:8081/",
			realm:    "myrealm",
			expected: "http://localhost:8081/realms/myrealm/protocol/openid-connect/certs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(Config{BaseURL: tt.baseURL, Realm: tt.realm})
			result := client.jwksURL()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestUserFromClaims(t *testing.T) {
	t.Run("full claims", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":                "user-123",
			"email":              "test@example.com",
			"preferred_username": "testuser",
			"name":               "Test User",
			"realm_access": map[string]any{
				"roles": []any{"admin", "user"},
			},
		}

		user := userFromClaims(claims)

		if user.ID != "user-123" {
			t.Errorf("expected ID 'user-123', got %q", user.ID)
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got %q", user.Email)
		}
		if user.Name != "Test User" {
			t.Errorf("expected name 'Test User', got %q", user.Name)
		}
		if len(user.Roles) != 2 {
			t.Errorf("expected 2 roles, got %d", len(user.Roles))
		}
	})

	t.Run("preferred username fallback", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":                "user-123",
			"preferred_username": "testuser",
		}

		user := userFromClaims(claims)

		if user.Name != "testuser" {
			t.Errorf("expected name to fall back to preferred_username, got %q", user.Name)
		}
	})

	t.Run("empty name uses preferred_username", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub":                "user-123",
			"preferred_username": "testuser",
			"name":               "",
		}

		user := userFromClaims(claims)

		if user.Name != "testuser" {
			t.Errorf("expected name to use preferred_username when name is empty, got %q", user.Name)
		}
	})

	t.Run("missing claims", func(t *testing.T) {
		claims := jwt.MapClaims{}

		user := userFromClaims(claims)

		if user.ID != "" {
			t.Errorf("expected empty ID, got %q", user.ID)
		}
		if user.Email != "" {
			t.Errorf("expected empty email, got %q", user.Email)
		}
		if user.Name != "" {
			t.Errorf("expected empty name, got %q", user.Name)
		}
		if len(user.Roles) != 0 {
			t.Errorf("expected no roles, got %v", user.Roles)
		}
	})
}

func TestRolesFromClaims(t *testing.T) {
	t.Run("valid roles", func(t *testing.T) {
		claims := jwt.MapClaims{
			"realm_access": map[string]any{
				"roles": []any{"admin", "user", "moderator"},
			},
		}

		roles := rolesFromClaims(claims)

		if len(roles) != 3 {
			t.Fatalf("expected 3 roles, got %d", len(roles))
		}
		if roles[0] != "admin" || roles[1] != "user" || roles[2] != "moderator" {
			t.Errorf("unexpected roles: %v", roles)
		}
	})

	t.Run("missing realm_access", func(t *testing.T) {
		claims := jwt.MapClaims{}

		roles := rolesFromClaims(claims)

		if roles != nil {
			t.Errorf("expected nil roles, got %v", roles)
		}
	})

	t.Run("missing roles array", func(t *testing.T) {
		claims := jwt.MapClaims{
			"realm_access": map[string]any{},
		}

		roles := rolesFromClaims(claims)

		if roles != nil {
			t.Errorf("expected nil roles, got %v", roles)
		}
	})

	t.Run("invalid realm_access type", func(t *testing.T) {
		claims := jwt.MapClaims{
			"realm_access": "invalid",
		}

		roles := rolesFromClaims(claims)

		if roles != nil {
			t.Errorf("expected nil roles, got %v", roles)
		}
	})

	t.Run("mixed role types filtered", func(t *testing.T) {
		claims := jwt.MapClaims{
			"realm_access": map[string]any{
				"roles": []any{"admin", 123, "user", nil},
			},
		}

		roles := rolesFromClaims(claims)

		if len(roles) != 2 {
			t.Fatalf("expected 2 string roles, got %d: %v", len(roles), roles)
		}
		if roles[0] != "admin" || roles[1] != "user" {
			t.Errorf("unexpected roles: %v", roles)
		}
	})
}

func TestVerify_MissingToken(t *testing.T) {
	client := New(Config{
		BaseURL:  "http://localhost:8081",
		Realm:    "test",
		ClientID: "testclient",
	})

	_, err := client.Verify(context.Background(), "")
	if err == nil {
		t.Error("expected error for missing token")
	}
	if err.Error() != "missing token" {
		t.Errorf("expected 'missing token' error, got %q", err.Error())
	}
}

func TestPasswordGrant(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("expected form content type, got %s", r.Header.Get("Content-Type"))
			}

			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}

			if r.Form.Get("grant_type") != "password" {
				t.Errorf("expected grant_type 'password', got %q", r.Form.Get("grant_type"))
			}
			if r.Form.Get("username") != "testuser" {
				t.Errorf("expected username 'testuser', got %q", r.Form.Get("username"))
			}
			if r.Form.Get("password") != "testpass" {
				t.Errorf("expected password 'testpass', got %q", r.Form.Get("password"))
			}
			if r.Form.Get("client_id") != "testclient" {
				t.Errorf("expected client_id 'testclient', got %q", r.Form.Get("client_id"))
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_in":    3600,
			})
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:  server.URL,
			Realm:    "test",
			ClientID: "testclient",
		})

		token, err := client.passwordGrant(context.Background(), "testuser", "testpass")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if token.AccessToken != "test-access-token" {
			t.Errorf("expected access token, got %q", token.AccessToken)
		}
		if token.RefreshToken != "test-refresh-token" {
			t.Errorf("expected refresh token, got %q", token.RefreshToken)
		}
		if token.ExpiresIn != 3600 {
			t.Errorf("expected expires_in 3600, got %d", token.ExpiresIn)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Invalid user credentials"))
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:  server.URL,
			Realm:    "test",
			ClientID: "testclient",
		})

		_, err := client.passwordGrant(context.Background(), "baduser", "badpass")
		if err == nil {
			t.Error("expected error for invalid credentials")
		}
		if err.Error() != "login failed: Invalid user credentials" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing access token in response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"refresh_token": "test-refresh-token",
			})
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:  server.URL,
			Realm:    "test",
			ClientID: "testclient",
		})

		_, err := client.passwordGrant(context.Background(), "testuser", "testpass")
		if err == nil {
			t.Error("expected error for missing access token")
		}
		if err.Error() != "missing access token" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:  server.URL,
			Realm:    "test",
			ClientID: "testclient",
		})

		_, err := client.passwordGrant(context.Background(), "testuser", "testpass")
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestUserIDFromLocationHeader(t *testing.T) {
	tests := []struct {
		name     string
		location string
		want     string
	}{
		{name: "full URL", location: "http://localhost:8081/admin/realms/test/users/abc-123", want: "abc-123"},
		{name: "trailing slash", location: "http://localhost:8081/admin/realms/test/users/abc-123/", want: "abc-123"},
		{name: "empty", location: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := userIDFromLocationHeader(tt.location)
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/realms/test/protocol/openid-connect/token":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"access_token": "admin-token",
					"expires_in":   3600,
				})
				return
			case r.URL.Path == "/admin/realms/test/users" && r.Method == http.MethodPost:
				w.Header().Set("Location", "/admin/realms/test/users/44444444-4444-4444-4444-444444444444")
				w.WriteHeader(http.StatusCreated)
				return
			default:
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:  server.URL,
			Realm:    "test",
			ClientID: "testclient",
			Username: "admin",
			Password: "admin123",
		})

		created, err := client.CreateUser(context.Background(), UserCreateInput{
			Username:  "new.user",
			Email:     "new.user@example.com",
			FirstName: "New",
			LastName:  "User",
			Password:  "supersecret123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created.ID != "44444444-4444-4444-4444-444444444444" {
			t.Fatalf("unexpected id: %q", created.ID)
		}
		if created.Name != "New User" {
			t.Fatalf("unexpected name: %q", created.Name)
		}
		if created.Email != "new.user@example.com" {
			t.Fatalf("unexpected email: %q", created.Email)
		}
	})

	t.Run("conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/realms/test/protocol/openid-connect/token":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"access_token": "admin-token",
					"expires_in":   3600,
				})
				return
			case r.URL.Path == "/admin/realms/test/users" && r.Method == http.MethodPost:
				http.Error(w, "already exists", http.StatusConflict)
				return
			default:
				t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
			}
		}))
		defer server.Close()

		client := New(Config{
			BaseURL:  server.URL,
			Realm:    "test",
			ClientID: "testclient",
			Username: "admin",
			Password: "admin123",
		})

		_, err := client.CreateUser(context.Background(), UserCreateInput{
			Username: "new.user",
			Email:    "new.user@example.com",
			Password: "supersecret123",
		})
		if err == nil {
			t.Fatal("expected conflict error")
		}
	})
}
