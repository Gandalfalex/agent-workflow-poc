package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDecodeJSON(t *testing.T) {
	type testPayload struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	t.Run("valid json", func(t *testing.T) {
		body := `{"name":"test","value":42}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rec := httptest.NewRecorder()

		result, ok := decodeJSON[testPayload](rec, req, "test")

		if !ok {
			t.Fatal("expected decoding to succeed")
		}
		if result.Name != "test" {
			t.Errorf("expected name 'test', got %q", result.Name)
		}
		if result.Value != 42 {
			t.Errorf("expected value 42, got %d", result.Value)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		body := `{"name": invalid}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		rec := httptest.NewRecorder()

		_, ok := decodeJSON[testPayload](rec, req, "test")

		if ok {
			t.Fatal("expected decoding to fail")
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var errResp errorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "invalid_json" {
			t.Errorf("expected error code 'invalid_json', got %q", errResp.Error)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		rec := httptest.NewRecorder()

		_, ok := decodeJSON[testPayload](rec, req, "test")

		if ok {
			t.Fatal("expected decoding to fail for empty body")
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})
}

func TestHandleNotFoundError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleNotFoundError(rec, req, nil, "project", "project_load")

		if handled {
			t.Error("expected nil error to not be handled")
		}
	})

	t.Run("not found error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleNotFoundError(rec, req, pgx.ErrNoRows, "project", "project_load")

		if !handled {
			t.Error("expected not found error to be handled")
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}

		var errResp errorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "not_found" {
			t.Errorf("expected error code 'not_found', got %q", errResp.Error)
		}
		if errResp.Message != "project not found" {
			t.Errorf("expected message 'project not found', got %q", errResp.Message)
		}
	})

	t.Run("other error not handled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleNotFoundError(rec, req, errors.New("some error"), "project", "project_load")

		if handled {
			t.Error("expected other errors to not be handled by handleNotFoundError")
		}
	})
}

func TestHandleDBError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDBError(rec, req, nil, "project", "project_load")

		if handled {
			t.Error("expected nil error to not be handled")
		}
	})

	t.Run("not found error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDBError(rec, req, pgx.ErrNoRows, "project", "project_load")

		if !handled {
			t.Error("expected not found error to be handled")
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("other error returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDBError(rec, req, errors.New("db connection failed"), "project", "project_load")

		if !handled {
			t.Error("expected error to be handled")
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var errResp errorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "project_load_failed" {
			t.Errorf("expected error code 'project_load_failed', got %q", errResp.Error)
		}
	})
}

func TestHandleDBErrorWithCode(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDBErrorWithCode(rec, req, nil, "project", "project_create", "project_create_failed")

		if handled {
			t.Error("expected nil error to not be handled")
		}
	})

	t.Run("not found error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDBErrorWithCode(rec, req, pgx.ErrNoRows, "project", "project_update", "project_update_failed")

		if !handled {
			t.Error("expected not found error to be handled")
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("other error returns 400 with custom code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDBErrorWithCode(rec, req, errors.New("key already exists"), "project", "project_create", "project_create_failed")

		if !handled {
			t.Error("expected error to be handled")
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		var errResp errorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "project_create_failed" {
			t.Errorf("expected error code 'project_create_failed', got %q", errResp.Error)
		}
		if errResp.Message != "key already exists" {
			t.Errorf("expected original error message, got %q", errResp.Message)
		}
	})
}

func TestHandleDeleteError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDeleteError(rec, req, nil, "project", "project_delete")

		if handled {
			t.Error("expected nil error to not be handled")
		}
	})

	t.Run("not found error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDeleteError(rec, req, pgx.ErrNoRows, "project", "project_delete")

		if !handled {
			t.Error("expected not found error to be handled")
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("other error returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleDeleteError(rec, req, errors.New("foreign key constraint"), "project", "project_delete")

		if !handled {
			t.Error("expected error to be handled")
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var errResp errorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "project_delete_failed" {
			t.Errorf("expected error code 'project_delete_failed', got %q", errResp.Error)
		}
	})
}

func TestHandleListError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleListError(rec, req, nil, "projects", "project_list")

		if handled {
			t.Error("expected nil error to not be handled")
		}
	})

	t.Run("error returns 500", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		handled := handleListError(rec, req, errors.New("db timeout"), "projects", "project_list")

		if !handled {
			t.Error("expected error to be handled")
		}
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}

		var errResp errorResponse
		if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}
		if errResp.Error != "project_list_failed" {
			t.Errorf("expected error code 'project_list_failed', got %q", errResp.Error)
		}
	})
}

func TestMapSlice(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		result := mapSlice([]int{}, func(i int) string { return "" })
		if len(result) != 0 {
			t.Errorf("expected empty slice, got length %d", len(result))
		}
	})

	t.Run("int to string", func(t *testing.T) {
		input := []int{1, 2, 3}
		result := mapSlice(input, func(i int) string {
			return string(rune('a' + i - 1))
		})

		if len(result) != 3 {
			t.Fatalf("expected 3 items, got %d", len(result))
		}
		expected := []string{"a", "b", "c"}
		for i, v := range result {
			if v != expected[i] {
				t.Errorf("at index %d: expected %q, got %q", i, expected[i], v)
			}
		}
	})

	t.Run("struct transformation", func(t *testing.T) {
		type input struct {
			Name string
			Age  int
		}
		type output struct {
			DisplayName string
		}

		items := []input{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}

		result := mapSlice(items, func(i input) output {
			return output{DisplayName: i.Name}
		})

		if len(result) != 2 {
			t.Fatalf("expected 2 items, got %d", len(result))
		}
		if result[0].DisplayName != "Alice" {
			t.Errorf("expected 'Alice', got %q", result[0].DisplayName)
		}
		if result[1].DisplayName != "Bob" {
			t.Errorf("expected 'Bob', got %q", result[1].DisplayName)
		}
	})
}
