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
