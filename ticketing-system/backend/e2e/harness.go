//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"ticketing-system/backend/internal/auth"
	"ticketing-system/backend/internal/blob"
	"ticketing-system/backend/internal/httpapi"
	"ticketing-system/backend/internal/migrate"
	"ticketing-system/backend/internal/store"
	"ticketing-system/backend/internal/webhook"
)

const (
	e2eUserToken   = "e2e-static-token"
	e2eViewerToken = "e2e-viewer-token"
)

type Harness struct {
	t           *testing.T
	config      harnessConfig
	contract    FrontendContract
	seed        SeedData
	releaseSlot func()

	ctx    context.Context
	cancel context.CancelFunc

	server    *httptest.Server
	store     *store.Store
	container *postgres.PostgresContainer
	blobStore *blob.MemoryStore

	playwright     *playwright.Playwright
	browser        playwright.Browser
	page           playwright.Page
	webhookCapture *capturingWebhooks
}

type HarnessOption func(*Harness)

func WithWebhookCapture() HarnessOption {
	return func(h *Harness) {
		h.webhookCapture = &capturingWebhooks{}
	}
}

func WithViewerUser() HarnessOption {
	return func(h *Harness) {
		h.config.loginAsViewer = true
	}
}

type harnessConfig struct {
	testTimeout       time.Duration
	stepTimeout       time.Duration
	navigationTimeout time.Duration
	headless          bool
	artifactsDir      string
	contractFile      string
	loginAsViewer     bool
}

type FrontendContract struct {
	SchemaVersion int                              `json:"schemaVersion"`
	GeneratedAt   string                           `json:"generatedAt"`
	SourceFile    string                           `json:"sourceFile"`
	RouterFile    string                           `json:"routerFile"`
	Routes        map[string]FrontendContractRoute `json:"routes"`
	Selectors     map[string]string                `json:"selectors"`
	Flows         map[string]FrontendContractFlow  `json:"flows"`
}

type FrontendContractRoute struct {
	Name   string   `json:"name"`
	Path   string   `json:"path"`
	Params []string `json:"params"`
}

type FrontendContractFlow struct {
	Route           string   `json:"route"`
	AssertSelectors []string `json:"assertSelectors"`
}

type SeedData struct {
	ProjectID    string
	StoryID      string
	UserID       string
	ViewerUserID string
	BacklogID    string
	InProgressID string
	DoneID       string
}

type staticAuthEntry struct {
	user     auth.User
	token    string
	password string
}

type staticAuth struct {
	entries []staticAuthEntry
}

func (a staticAuth) Login(_ context.Context, username, password string) (auth.User, auth.TokenSet, error) {
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return auth.User{}, auth.TokenSet{}, errors.New("missing credentials")
	}
	for _, e := range a.entries {
		if (e.user.Email == username || e.user.Name == username) && e.password == password {
			return e.user, auth.TokenSet{
				AccessToken: e.token,
				ExpiresIn:   3600,
			}, nil
		}
	}
	return auth.User{}, auth.TokenSet{}, errors.New("invalid credentials")
}

func (a staticAuth) Verify(_ context.Context, token string) (auth.User, error) {
	for _, e := range a.entries {
		if e.token == token {
			return e.user, nil
		}
	}
	return auth.User{}, errors.New("invalid token")
}

func (a staticAuth) ListUsers(_ context.Context) ([]auth.User, error) {
	users := make([]auth.User, len(a.entries))
	for i, e := range a.entries {
		users[i] = e.user
	}
	return users, nil
}

type noopWebhooks struct{}

func (noopWebhooks) Dispatch(context.Context, uuid.UUID, string, any) {}

func (noopWebhooks) Test(context.Context, store.Webhook, string, any) (webhook.Result, error) {
	return webhook.Result{Delivered: true, StatusCode: http.StatusOK}, nil
}

// capturingWebhooks records every dispatched event for test assertions.
type capturingWebhooks struct {
	mu     sync.Mutex
	events []capturedEvent
}

type capturedEvent struct {
	ProjectID uuid.UUID
	Event     string
	Data      any
}

func (c *capturingWebhooks) Dispatch(_ context.Context, projectID uuid.UUID, event string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, capturedEvent{ProjectID: projectID, Event: event, Data: data})
}

func (c *capturingWebhooks) Test(_ context.Context, _ store.Webhook, _ string, _ any) (webhook.Result, error) {
	return webhook.Result{Delivered: true, StatusCode: http.StatusOK}, nil
}

func (c *capturingWebhooks) Events() []capturedEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]capturedEvent, len(c.events))
	copy(cp, c.events)
	return cp
}

func (c *capturingWebhooks) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = nil
}

func (c *capturingWebhooks) HasEvent(event string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.events {
		if e.Event == event {
			return true
		}
	}
	return false
}

func (c *capturingWebhooks) WaitForEvent(event string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if c.HasEvent(event) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

type failureArtifacts struct {
	URL        string
	Screenshot string
	HTML       string
}

var routeParamPattern = regexp.MustCompile(`:[a-zA-Z0-9_]+`)

var (
	harnessParallelOnce  sync.Once
	harnessParallelSlots chan struct{}
)

func NewHarness(t *testing.T, opts ...HarnessOption) *Harness {
	t.Helper()

	cfg := loadHarnessConfig()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.testTimeout)
	releaseSlot := acquireHarnessSlot()

	h := &Harness{
		t:           t,
		config:      cfg,
		ctx:         ctx,
		cancel:      cancel,
		releaseSlot: releaseSlot,
	}

	for _, opt := range opts {
		opt(h)
	}

	if err := h.start(); err != nil {
		releaseSlot()
		cancel()
		t.Fatalf("start e2e harness: %v", err)
	}

	return h
}

func (h *Harness) WebhookCapture() *capturingWebhooks {
	return h.webhookCapture
}

func (h *Harness) Close() {
	h.t.Helper()

	if h.page != nil {
		if err := h.page.Close(); err != nil {
			h.t.Logf("close page: %v", err)
		}
	}
	if h.browser != nil {
		if err := h.browser.Close(); err != nil {
			h.t.Logf("close browser: %v", err)
		}
	}
	if h.playwright != nil {
		if err := h.playwright.Stop(); err != nil {
			h.t.Logf("stop playwright: %v", err)
		}
	}
	if h.server != nil {
		h.server.Close()
	}
	if h.store != nil {
		h.store.Close()
	}
	if h.container != nil {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.container.Terminate(cleanupCtx); err != nil {
			h.t.Logf("terminate postgres container: %v", err)
		}
	}

	if h.cancel != nil {
		h.cancel()
	}
	if h.releaseSlot != nil {
		h.releaseSlot()
		h.releaseSlot = nil
	}
}

func (h *Harness) BaseURL() string {
	return h.server.URL
}

func (h *Harness) SeedData() SeedData {
	return h.seed
}

func (h *Harness) Store() *store.Store {
	return h.store
}

func (h *Harness) BlobStore() *blob.MemoryStore {
	return h.blobStore
}

func (h *Harness) Context() context.Context {
	return h.ctx
}

// LoginCredentials returns the (identifier, password) for the active user.
func (h *Harness) LoginCredentials() (string, string) {
	if h.config.loginAsViewer {
		return "NormalUser", "viewer123"
	}
	return "AdminUser", "admin123"
}

// APIRequest makes a direct HTTP request to the backend with the active user's auth cookie.
func (h *Harness) APIRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := h.resolveURL(path)
	req, err := http.NewRequestWithContext(h.ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	token := e2eUserToken
	if h.config.loginAsViewer {
		token = e2eViewerToken
	}
	req.AddCookie(&http.Cookie{Name: "ticketing_session", Value: token})
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: h.config.stepTimeout}
	return client.Do(req)
}

func (h *Harness) ExpectSelectorHidden(selector string) error {
	count, err := h.page.Locator(selector).Count()
	if err != nil {
		return fmt.Errorf("count selector %q: %w", selector, err)
	}
	if count > 0 {
		visible, err := h.page.Locator(selector).First().IsVisible()
		if err != nil {
			return fmt.Errorf("check visibility of %q: %w", selector, err)
		}
		if visible {
			return fmt.Errorf("selector %q is still visible", selector)
		}
	}
	return nil
}

func (h *Harness) ExpectSelectorHiddenKey(selectorKey string) error {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return err
	}
	return h.ExpectSelectorHidden(selector)
}

func (h *Harness) ElementCount(selector string) (int, error) {
	count, err := h.page.Locator(selector).Count()
	if err != nil {
		return 0, fmt.Errorf("count selector %q: %w", selector, err)
	}
	return count, nil
}

func (h *Harness) ExpectMinElements(selector string, min int) error {
	deadline := time.Now().Add(h.config.stepTimeout)
	for time.Now().Before(deadline) {
		count, err := h.ElementCount(selector)
		if err != nil {
			return err
		}
		if count >= min {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	count, _ := h.ElementCount(selector)
	return fmt.Errorf("expected at least %d elements matching %q, got %d", min, selector, count)
}

func (h *Harness) HealthCheck() error {
	req, err := http.NewRequestWithContext(h.ctx, http.MethodGet, h.BaseURL()+"/health", nil)
	if err != nil {
		return fmt.Errorf("build health request: %w", err)
	}
	client := &http.Client{Timeout: h.config.stepTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute health request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

func (h *Harness) GoTo(path string) error {
	if h.page == nil {
		return errors.New("browser page not initialized")
	}
	url := h.resolveURL(path)
	resp, err := h.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(durationMS(h.config.navigationTimeout)),
	})
	if err != nil {
		return fmt.Errorf("navigate to %s: %w", url, err)
	}
	if resp != nil && resp.Status() >= 400 {
		return fmt.Errorf("navigation returned status %d for %s", resp.Status(), url)
	}
	if err := h.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateLoad,
		Timeout: playwright.Float(durationMS(h.config.navigationTimeout)),
	}); err != nil {
		return fmt.Errorf("wait for load state on %s: %w", url, err)
	}
	return h.WaitVisible("body")
}

func (h *Harness) GoToRoute(routeKey string, params map[string]string) error {
	path, err := h.ResolveRoute(routeKey, params)
	if err != nil {
		return err
	}
	return h.GoTo(path)
}

func (h *Harness) ResolveRoute(routeKey string, params map[string]string) (string, error) {
	route, ok := h.contract.Routes[routeKey]
	if !ok {
		return "", fmt.Errorf("unknown contract route key %q", routeKey)
	}
	renderedPath, err := renderRoutePath(route.Path, params)
	if err != nil {
		return "", fmt.Errorf("resolve route key %q: %w", routeKey, err)
	}
	return renderedPath, nil
}

func (h *Harness) Selector(selectorKey string) (string, error) {
	selector, ok := h.contract.Selectors[selectorKey]
	if !ok {
		return "", fmt.Errorf("unknown contract selector key %q", selectorKey)
	}
	return selector, nil
}

func (h *Harness) WaitVisible(selector string) error {
	_, err := h.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	})
	if err != nil {
		return fmt.Errorf("wait for selector %q visible: %w", selector, err)
	}
	return nil
}

func (h *Harness) WaitVisibleKey(selectorKey string) error {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return err
	}
	return h.WaitVisible(selector)
}

func (h *Harness) WaitHidden(selector string) error {
	_, err := h.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	})
	if err != nil {
		return fmt.Errorf("wait for selector %q hidden: %w", selector, err)
	}
	return nil
}

func (h *Harness) Click(selector string) error {
	if err := h.WaitVisible(selector); err != nil {
		return err
	}
	if err := h.page.Locator(selector).Click(playwright.LocatorClickOptions{
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	}); err != nil {
		return fmt.Errorf("click selector %q: %w", selector, err)
	}
	return nil
}

func (h *Harness) ClickKey(selectorKey string) error {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return err
	}
	return h.Click(selector)
}

func (h *Harness) Fill(selector, value string) error {
	if err := h.WaitVisible(selector); err != nil {
		return err
	}
	if err := h.page.Locator(selector).Fill(value, playwright.LocatorFillOptions{
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	}); err != nil {
		return fmt.Errorf("fill selector %q: %w", selector, err)
	}
	return nil
}

func (h *Harness) FillKey(selectorKey, value string) error {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return err
	}
	return h.Fill(selector, value)
}

func (h *Harness) Press(selector, key string) error {
	if err := h.WaitVisible(selector); err != nil {
		return err
	}
	if err := h.page.Locator(selector).Press(key, playwright.LocatorPressOptions{
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	}); err != nil {
		return fmt.Errorf("press %q on selector %q: %w", key, selector, err)
	}
	return nil
}

func (h *Harness) PressKey(selectorKey, key string) error {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return err
	}
	return h.Press(selector, key)
}

func (h *Harness) SelectOptionByValue(selector, value string) error {
	if err := h.WaitVisible(selector); err != nil {
		return err
	}
	values := []string{value}
	if _, err := h.page.Locator(selector).SelectOption(playwright.SelectOptionValues{
		Values: &values,
	}, playwright.LocatorSelectOptionOptions{
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	}); err != nil {
		return fmt.Errorf("select option %q on selector %q: %w", value, selector, err)
	}
	return nil
}

func (h *Harness) SelectOptionByValueKey(selectorKey, value string) error {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return err
	}
	return h.SelectOptionByValue(selector, value)
}

func (h *Harness) ExpectTextVisible(text string) error {
	if err := h.page.GetByText(text).WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(durationMS(h.config.stepTimeout)),
	}); err != nil {
		return fmt.Errorf("wait for text %q visible: %w", text, err)
	}
	return nil
}

func (h *Harness) ExpectTextHidden(text string) error {
	deadline := time.Now().Add(h.config.stepTimeout)
	for time.Now().Before(deadline) {
		count, err := h.page.GetByText(text).Count()
		if err != nil {
			return fmt.Errorf("count text %q: %w", text, err)
		}
		if count == 0 {
			return nil
		}
		visible, err := h.page.GetByText(text).First().IsVisible()
		if err != nil {
			return fmt.Errorf("check visibility of text %q: %w", text, err)
		}
		if !visible {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("text %q is still visible", text)
}

func (h *Harness) IsButtonDisabledKey(selectorKey string) (bool, error) {
	selector, err := h.Selector(selectorKey)
	if err != nil {
		return false, err
	}
	if err := h.WaitVisible(selector); err != nil {
		return false, err
	}
	return h.page.Locator(selector).IsDisabled()
}

func (h *Harness) HandleNextDialog(accept bool) {
	h.page.OnDialog(func(dialog playwright.Dialog) {
		if accept {
			_ = dialog.Accept()
		} else {
			_ = dialog.Dismiss()
		}
	})
}

func (h *Harness) ExpectURLContains(fragment string) error {
	deadline := time.Now().Add(h.config.stepTimeout)
	for time.Now().Before(deadline) {
		if strings.Contains(h.page.URL(), fragment) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("url %q does not contain %q", h.page.URL(), fragment)
}

func (h *Harness) failStep(step string, err error) {
	h.t.Helper()
	artifacts := h.captureFailureArtifacts(step)

	var details strings.Builder
	details.WriteString(fmt.Sprintf("step failed: %s\n", step))
	details.WriteString(fmt.Sprintf("error: %v\n", err))
	if artifacts.URL != "" {
		details.WriteString(fmt.Sprintf("page url: %s\n", artifacts.URL))
	}
	if artifacts.Screenshot != "" {
		details.WriteString(fmt.Sprintf("screenshot: %s\n", artifacts.Screenshot))
	}
	if artifacts.HTML != "" {
		details.WriteString(fmt.Sprintf("html snapshot: %s\n", artifacts.HTML))
	}

	h.t.Fatal(details.String())
}

func (h *Harness) start() error {
	container, err := runPostgresContainer(h.ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ticketing"),
		postgres.WithUsername("ticketing"),
		postgres.WithPassword("ticketing"),
		postgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("start postgres container: %w", err)
	}
	h.container = container

	dbURL, err := container.ConnectionString(h.ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("postgres connection string: %w", err)
	}

	st, err := store.New(h.ctx, dbURL)
	if err != nil {
		return fmt.Errorf("connect store: %w", err)
	}
	h.store = st

	if err := migrate.Apply(h.ctx, st.DB(), migrationsDir(h.t)); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	testUserID := uuid.New()
	viewerUserID := uuid.New()
	authStub := staticAuth{
		entries: []staticAuthEntry{
			{
				user: auth.User{
					ID:    testUserID.String(),
					Email: "e2e@example.local",
					Name:  "AdminUser",
					Roles: []string{"admin"},
				},
				token:    e2eUserToken,
				password: "admin123",
			},
			{
				user: auth.User{
					ID:    viewerUserID.String(),
					Email: "viewer@example.local",
					Name:  "NormalUser",
					Roles: []string{"default-roles-ticketing"},
				},
				token:    e2eViewerToken,
				password: "viewer123",
			},
		},
	}

	seed, err := seedDefaultData(h.ctx, st, testUserID, viewerUserID)
	if err != nil {
		return fmt.Errorf("seed default data: %w", err)
	}
	h.seed = seed

	var webhookDispatcher httpapi.WebhookDispatcher = noopWebhooks{}
	if h.webhookCapture != nil {
		webhookDispatcher = h.webhookCapture
	}

	memBlob := blob.NewMemory()
	h.blobStore = memBlob

	api := httpapi.NewHandler(st, authStub, webhookDispatcher, httpapi.HandlerOptions{
		CookieName: "ticketing_session",
		BlobStore:  memBlob,
	})

	frontendDir := frontendDistDir(h.t)
	handler := httpapi.WithFrontend(httpapi.Router(api), frontendDir, "/")
	h.server = httptest.NewServer(handler)

	contract, err := loadFrontendContract(h.config.contractFile)
	if err != nil {
		return fmt.Errorf("load frontend contract: %w", err)
	}
	h.contract = contract

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("start playwright runtime: %w", err)
	}
	h.playwright = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(h.config.headless),
	})
	if err != nil {
		return fmt.Errorf("launch chromium: %w", err)
	}
	h.browser = browser

	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("create browser page: %w", err)
	}
	page.SetDefaultTimeout(durationMS(h.config.stepTimeout))
	page.SetDefaultNavigationTimeout(durationMS(h.config.navigationTimeout))
	h.page = page

	return nil
}

func (h *Harness) resolveURL(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "/" {
		return h.BaseURL()
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return h.BaseURL() + trimmed
}

func (h *Harness) captureFailureArtifacts(step string) failureArtifacts {
	result := failureArtifacts{}
	if h.page == nil {
		return result
	}

	result.URL = h.page.URL()
	baseDir := filepath.Join(h.config.artifactsDir, sanitizePath(h.t.Name()))
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		h.t.Logf("create artifacts directory: %v", err)
		return result
	}

	prefix := fmt.Sprintf("%s-%s", time.Now().UTC().Format("20060102-150405"), sanitizePath(step))

	screenshotPath := filepath.Join(baseDir, prefix+".png")
	if _, err := h.page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(screenshotPath),
		FullPage: playwright.Bool(true),
	}); err == nil {
		result.Screenshot = screenshotPath
	} else {
		h.t.Logf("capture screenshot: %v", err)
	}

	htmlPath := filepath.Join(baseDir, prefix+".html")
	if content, err := h.page.Content(); err == nil {
		if writeErr := os.WriteFile(htmlPath, []byte(content), 0o644); writeErr == nil {
			result.HTML = htmlPath
		} else {
			h.t.Logf("write html snapshot: %v", writeErr)
		}
	} else {
		h.t.Logf("capture html snapshot: %v", err)
	}

	return result
}

func loadHarnessConfig() harnessConfig {
	artifacts := os.Getenv("E2E_ARTIFACTS_DIR")
	if strings.TrimSpace(artifacts) == "" {
		artifacts = filepath.Join(e2eDir(), "artifacts")
	}

	return harnessConfig{
		testTimeout:       envDuration("E2E_TEST_TIMEOUT", 3*time.Minute),
		stepTimeout:       envDuration("E2E_STEP_TIMEOUT", 15*time.Second),
		navigationTimeout: envDuration("E2E_NAV_TIMEOUT", 20*time.Second),
		headless:          envBool("E2E_HEADLESS", true),
		artifactsDir:      artifacts,
		contractFile:      frontendContractFile(),
	}
}

func loadFrontendContract(file string) (FrontendContract, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return FrontendContract{}, fmt.Errorf("read %s: %w", file, err)
	}
	var contract FrontendContract
	if err := json.Unmarshal(content, &contract); err != nil {
		return FrontendContract{}, fmt.Errorf("decode %s: %w", file, err)
	}
	if len(contract.Routes) == 0 {
		return FrontendContract{}, fmt.Errorf("%s has no routes", file)
	}
	if len(contract.Selectors) == 0 {
		return FrontendContract{}, fmt.Errorf("%s has no selectors", file)
	}
	return contract, nil
}

func frontendContractFile() string {
	if custom := strings.TrimSpace(os.Getenv("E2E_FRONTEND_CONTRACT")); custom != "" {
		return custom
	}
	return filepath.Join(e2eDir(), "contracts", "frontend_contract.json")
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func acquireHarnessSlot() func() {
	harnessParallelOnce.Do(func() {
		maxParallel := envInt("E2E_MAX_PARALLEL", 2)
		if maxParallel < 1 {
			maxParallel = 1
		}
		harnessParallelSlots = make(chan struct{}, maxParallel)
	})
	harnessParallelSlots <- struct{}{}
	return func() {
		<-harnessParallelSlots
	}
}

func runPostgresContainer(ctx context.Context, image string, opts ...testcontainers.ContainerCustomizer) (container *postgres.PostgresContainer, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("testcontainers panic: %v", r)
		}
	}()
	return postgres.Run(ctx, image, opts...)
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	path := filepath.Join(e2eDir(), "..", "migrations")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("migrations directory not found at %s: %v", path, err)
	}
	return path
}

func frontendDistDir(t *testing.T) string {
	t.Helper()
	if custom := strings.TrimSpace(os.Getenv("E2E_FRONTEND_DIR")); custom != "" {
		if _, err := os.Stat(filepath.Join(custom, "index.html")); err == nil {
			return custom
		}
		t.Fatalf("E2E_FRONTEND_DIR does not contain index.html: %s", custom)
	}

	path := filepath.Join(e2eDir(), "..", "..", "frontend", "dist")
	if _, err := os.Stat(filepath.Join(path, "index.html")); err != nil {
		t.Skipf("frontend dist not found at %s (run `cd ticketing-system/frontend && npm run build`)", path)
	}
	return path
}

func e2eDir() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(currentFile)
}

func renderRoutePath(routePath string, params map[string]string) (string, error) {
	if params == nil {
		params = map[string]string{}
	}
	var missing []string
	rendered := routeParamPattern.ReplaceAllStringFunc(routePath, func(token string) string {
		name := strings.TrimPrefix(token, ":")
		value := strings.TrimSpace(params[name])
		if value == "" {
			missing = append(missing, name)
			return token
		}
		return value
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("missing route params: %s", strings.Join(missing, ", "))
	}
	return rendered, nil
}

func seedDefaultData(ctx context.Context, st *store.Store, userID, viewerUserID uuid.UUID) (SeedData, error) {
	project, err := st.CreateProject(ctx, store.ProjectCreateInput{
		Key:  "E2E1",
		Name: "E2E Project",
	})
	if err != nil {
		return SeedData{}, fmt.Errorf("create project: %w", err)
	}

	if err := st.UpsertUser(ctx, store.UserUpsertInput{
		ID:    userID,
		Name:  "AdminUser",
		Email: "e2e@example.local",
	}); err != nil {
		return SeedData{}, fmt.Errorf("upsert user: %w", err)
	}

	group, err := st.CreateGroup(ctx, store.GroupCreateInput{
		Name: "E2E Group",
	})
	if err != nil {
		return SeedData{}, fmt.Errorf("create group: %w", err)
	}

	if _, err := st.AddGroupMember(ctx, group.ID, userID); err != nil {
		return SeedData{}, fmt.Errorf("add group member: %w", err)
	}

	if _, err := st.AddProjectGroup(ctx, project.ID, group.ID, "admin"); err != nil {
		return SeedData{}, fmt.Errorf("add project group: %w", err)
	}

	// Seed viewer user + group with viewer role
	if err := st.UpsertUser(ctx, store.UserUpsertInput{
		ID:    viewerUserID,
		Name:  "NormalUser",
		Email: "viewer@example.local",
	}); err != nil {
		return SeedData{}, fmt.Errorf("upsert viewer user: %w", err)
	}

	viewerGroup, err := st.CreateGroup(ctx, store.GroupCreateInput{
		Name: "Viewer Group",
	})
	if err != nil {
		return SeedData{}, fmt.Errorf("create viewer group: %w", err)
	}

	if _, err := st.AddGroupMember(ctx, viewerGroup.ID, viewerUserID); err != nil {
		return SeedData{}, fmt.Errorf("add viewer group member: %w", err)
	}

	if _, err := st.AddProjectGroup(ctx, project.ID, viewerGroup.ID, "viewer"); err != nil {
		return SeedData{}, fmt.Errorf("add viewer project group: %w", err)
	}

	states, err := st.ReplaceWorkflowStates(ctx, project.ID, []store.WorkflowStateInput{
		{Name: "Backlog", Order: 1, IsDefault: true, IsClosed: false},
		{Name: "In Progress", Order: 2, IsDefault: false, IsClosed: false},
		{Name: "Done", Order: 3, IsDefault: false, IsClosed: true},
	})
	if err != nil {
		return SeedData{}, fmt.Errorf("replace workflow states: %w", err)
	}

	stateMap := map[string]string{}
	for _, s := range states {
		stateMap[s.Name] = s.ID.String()
	}

	story, err := st.CreateStory(ctx, project.ID, store.StoryCreateInput{
		Title: "E2E Story",
	})
	if err != nil {
		return SeedData{}, fmt.Errorf("create story: %w", err)
	}

	return SeedData{
		ProjectID:    project.ID.String(),
		StoryID:      story.ID.String(),
		UserID:       userID.String(),
		ViewerUserID: viewerUserID.String(),
		BacklogID:    stateMap["Backlog"],
		InProgressID: stateMap["In Progress"],
		DoneID:       stateMap["Done"],
	}, nil
}

func sanitizePath(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return "step"
	}
	var b strings.Builder
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		if r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('-')
	}
	return strings.Trim(b.String(), "-")
}

func durationMS(d time.Duration) float64 {
	return float64(d.Milliseconds())
}
