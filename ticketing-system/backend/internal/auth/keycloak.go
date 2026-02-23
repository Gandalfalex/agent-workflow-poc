package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	BaseURL  string
	Realm    string
	ClientID string
	Username string // Admin username for Keycloak admin API
	Password string // Admin password for Keycloak admin API
	Timeout  time.Duration
}

type TokenSet struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type User struct {
	ID    string
	Email string
	Name  string
	Roles []string
}

type UserCreateInput struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	Password  string
}

type Client struct {
	cfg        Config
	httpClient *http.Client

	jwksOnce sync.Once
	jwks     *keyfunc.JWKS
	jwksErr  error
}

func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Login(ctx context.Context, username, password string) (User, TokenSet, error) {
	token, err := c.passwordGrant(ctx, username, password)
	if err != nil {
		return User{}, TokenSet{}, err
	}

	user, err := c.Verify(ctx, token.AccessToken)
	if err != nil {
		return User{}, TokenSet{}, err
	}

	return user, token, nil
}

func (c *Client) CreateUser(ctx context.Context, input UserCreateInput) (User, error) {
	token, err := c.adminAccessToken(ctx)
	if err != nil {
		return User{}, fmt.Errorf("failed to get admin token: %w", err)
	}

	usersURL := strings.TrimRight(c.cfg.BaseURL, "/") + "/admin/realms/" + c.cfg.Realm + "/users"
	payload := map[string]any{
		"username":      strings.TrimSpace(input.Username),
		"email":         strings.TrimSpace(input.Email),
		"enabled":       true,
		"emailVerified": true,
		"firstName":     strings.TrimSpace(input.FirstName),
		"lastName":      strings.TrimSpace(input.LastName),
		"credentials": []map[string]any{
			{
				"type":      "password",
				"value":     input.Password,
				"temporary": false,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return User{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, usersURL, strings.NewReader(string(body)))
	if err != nil {
		return User{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		// continue
	case http.StatusConflict:
		raw, _ := io.ReadAll(resp.Body)
		return User{}, fmt.Errorf("user already exists: %s", strings.TrimSpace(string(raw)))
	default:
		raw, _ := io.ReadAll(resp.Body)
		return User{}, fmt.Errorf("keycloak create user error: %d - %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	id := userIDFromLocationHeader(resp.Header.Get("Location"))
	if id == "" {
		foundID, findErr := c.findUserIDByUsername(ctx, token, input.Username)
		if findErr != nil {
			return User{}, findErr
		}
		id = foundID
	}

	name := strings.TrimSpace(input.FirstName + " " + input.LastName)
	if name == "" {
		name = strings.TrimSpace(input.Username)
	}
	email := strings.TrimSpace(input.Email)

	return User{
		ID:    id,
		Name:  name,
		Email: email,
	}, nil
}

func (c *Client) Verify(ctx context.Context, tokenString string) (User, error) {
	if tokenString == "" {
		return User{}, errors.New("missing token")
	}
	if err := c.initJWKS(); err != nil {
		return User{}, err
	}

	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	token, err := parser.ParseWithClaims(tokenString, claims, c.jwks.Keyfunc)
	if err != nil {
		return User{}, err
	}
	if !token.Valid {
		return User{}, errors.New("invalid token")
	}

	issuer := fmt.Sprintf("%s/realms/%s", strings.TrimRight(c.cfg.BaseURL, "/"), c.cfg.Realm)
	if iss, ok := claims["iss"].(string); ok && iss != issuer {
		return User{}, errors.New("invalid issuer")
	}

	return userFromClaims(claims), nil
}

func (c *Client) passwordGrant(ctx context.Context, username, password string) (TokenSet, error) {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", c.cfg.ClientID)
	form.Set("username", username)
	form.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return TokenSet{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return TokenSet{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenSet{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return TokenSet{}, fmt.Errorf("login failed: %s", strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return TokenSet{}, err
	}

	if payload.AccessToken == "" {
		return TokenSet{}, errors.New("missing access token")
	}

	return TokenSet{
		AccessToken:  payload.AccessToken,
		RefreshToken: payload.RefreshToken,
		ExpiresIn:    payload.ExpiresIn,
	}, nil
}

// ListUsers retrieves all users from Keycloak admin API
func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	token, err := c.adminAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	// Construct admin API URL
	usersURL := strings.TrimRight(c.cfg.BaseURL, "/") + "/admin/realms/" + c.cfg.Realm + "/users"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usersURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("keycloak admin API error: %d - %s", resp.StatusCode, string(body))
	}

	var keycloakUsers []struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Email         string `json:"email"`
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		Enabled       bool   `json:"enabled"`
		EmailVerified bool   `json:"emailVerified"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&keycloakUsers); err != nil {
		return nil, err
	}

	users := make([]User, 0, len(keycloakUsers))
	for _, ku := range keycloakUsers {
		if !ku.Enabled {
			continue // Skip disabled users
		}

		name := strings.TrimSpace(ku.FirstName + " " + ku.LastName)
		if name == "" {
			name = ku.Username
		}

		email := ku.Email
		if email == "" {
			email = ku.Username + "@local"
		}

		users = append(users, User{
			ID:    ku.ID,
			Name:  name,
			Email: email,
		})
	}

	return users, nil
}

func (c *Client) adminAccessToken(ctx context.Context) (string, error) {
	tokenSet, err := c.passwordGrant(ctx, c.cfg.Username, c.cfg.Password)
	if err != nil {
		return "", err
	}
	return tokenSet.AccessToken, nil
}

func userIDFromLocationHeader(location string) string {
	trimmed := strings.TrimSpace(location)
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(strings.TrimRight(trimmed, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func (c *Client) findUserIDByUsername(ctx context.Context, token, username string) (string, error) {
	lookupURL := strings.TrimRight(c.cfg.BaseURL, "/") + "/admin/realms/" + c.cfg.Realm + "/users?exact=true&max=1&username=" + url.QueryEscape(strings.TrimSpace(username))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("keycloak user lookup error: %d - %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var users []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", err
	}
	if len(users) == 0 || strings.TrimSpace(users[0].ID) == "" {
		return "", errors.New("created user id not returned by identity provider")
	}
	return strings.TrimSpace(users[0].ID), nil
}

func (c *Client) initJWKS() error {
	c.jwksOnce.Do(func() {
		jwks, err := keyfunc.Get(c.jwksURL(), keyfunc.Options{
			RefreshInterval:   time.Hour,
			RefreshRateLimit:  time.Minute,
			RefreshTimeout:    10 * time.Second,
			RefreshUnknownKID: true,
		})
		c.jwks = jwks
		c.jwksErr = err
	})
	return c.jwksErr
}

func (c *Client) tokenURL() string {
	base := strings.TrimRight(c.cfg.BaseURL, "/")
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", base, c.cfg.Realm)
}

func (c *Client) jwksURL() string {
	base := strings.TrimRight(c.cfg.BaseURL, "/")
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", base, c.cfg.Realm)
}

func userFromClaims(claims jwt.MapClaims) User {
	user := User{}
	if sub, ok := claims["sub"].(string); ok {
		user.ID = sub
	}
	if email, ok := claims["email"].(string); ok {
		user.Email = email
	}
	name := ""
	if preferred, ok := claims["preferred_username"].(string); ok {
		name = preferred
	}
	if full, ok := claims["name"].(string); ok && full != "" {
		name = full
	}
	user.Name = name
	user.Roles = rolesFromClaims(claims)
	return user
}

func rolesFromClaims(claims jwt.MapClaims) []string {
	realmAccess, ok := claims["realm_access"].(map[string]any)
	if !ok {
		return nil
	}
	rawRoles, ok := realmAccess["roles"].([]any)
	if !ok {
		return nil
	}

	roles := make([]string, 0, len(rawRoles))
	for _, role := range rawRoles {
		if value, ok := role.(string); ok {
			roles = append(roles, value)
		}
	}
	return roles
}
