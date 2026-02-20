package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWithFrontend_HTMLNavigationServesIndexForSPAPaths(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html><body>spa</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	apiCalled := false
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		http.NotFound(w, r)
	})

	h := WithFrontend(api, dir, "")

	req := httptest.NewRequest(http.MethodGet, "/projects/abc/settings", nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if apiCalled {
		t.Fatal("expected SPA navigation not to hit API handler")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "spa") {
		t.Fatalf("expected index.html body, got %q", rec.Body.String())
	}
}

func TestWithFrontend_APIRequestStillHitsAPI(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html><body>spa</body></html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	apiCalled := false
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusNoContent)
	})

	h := WithFrontend(api, dir, "")

	req := httptest.NewRequest(http.MethodGet, "/projects/abc/board", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !apiCalled {
		t.Fatal("expected API request to hit API handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 from API, got %d", rec.Code)
	}
}
