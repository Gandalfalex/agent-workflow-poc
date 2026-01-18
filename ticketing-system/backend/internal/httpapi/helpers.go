package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// decodeJSON decodes JSON from the request body into the provided value.
// Returns false and writes an error response if decoding fails.
// Logs the error with request context for debugging.
func decodeJSON[T any](w http.ResponseWriter, r *http.Request, logCode string) (T, bool) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		logRequestError(r, logCode+"_invalid_json", err)
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return v, false
	}
	return v, true
}

// handleNotFoundError checks if the error is a "not found" error (pgx.ErrNoRows)
// and writes appropriate error response. Returns true if an error was handled.
// resourceType is used for the error message (e.g., "project", "ticket").
func handleNotFoundError(w http.ResponseWriter, r *http.Request, err error, resourceType, logCode string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		logRequestError(r, logCode+"_not_found", err)
		writeError(w, http.StatusNotFound, "not_found", resourceType+" not found")
		return true
	}
	return false
}

// handleDBError handles database errors by checking for "not found" first,
// then falling back to internal server error. Returns true if an error was handled.
// Use this for read operations where you expect either a result or not found.
func handleDBError(w http.ResponseWriter, r *http.Request, err error, resourceType, logCode string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		logRequestError(r, logCode+"_not_found", err)
		writeError(w, http.StatusNotFound, "not_found", resourceType+" not found")
		return true
	}
	logRequestError(r, logCode+"_failed", err)
	writeError(w, http.StatusInternalServerError, logCode+"_failed", "failed to load "+resourceType)
	return true
}

// handleDBErrorWithCode handles database errors with a custom error code suffix.
// Useful for operations like create/update where you want specific error codes.
// Returns true if an error was handled.
func handleDBErrorWithCode(w http.ResponseWriter, r *http.Request, err error, resourceType, logCode, errCode string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		logRequestError(r, logCode+"_not_found", err)
		writeError(w, http.StatusNotFound, "not_found", resourceType+" not found")
		return true
	}
	logRequestError(r, logCode+"_failed", err)
	writeError(w, http.StatusBadRequest, errCode, err.Error())
	return true
}

// handleDeleteError handles errors from delete operations.
// Returns not found for pgx.ErrNoRows, internal server error otherwise.
func handleDeleteError(w http.ResponseWriter, r *http.Request, err error, resourceType, logCode string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		logRequestError(r, logCode+"_not_found", err)
		writeError(w, http.StatusNotFound, "not_found", resourceType+" not found")
		return true
	}
	logRequestError(r, logCode+"_failed", err)
	writeError(w, http.StatusInternalServerError, logCode+"_failed", "failed to delete "+resourceType)
	return true
}

// handleListError handles errors from list operations.
// Always returns internal server error since lists don't have "not found" semantics.
func handleListError(w http.ResponseWriter, r *http.Request, err error, resourceType, logCode string) bool {
	if err == nil {
		return false
	}
	logRequestError(r, logCode+"_failed", err)
	writeError(w, http.StatusInternalServerError, logCode+"_failed", "failed to list "+resourceType)
	return true
}

// mapSlice transforms a slice of type T to a slice of type U using the provided mapper function.
// This eliminates repetitive slice mapping boilerplate throughout the codebase.
func mapSlice[T, U any](items []T, mapper func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = mapper(item)
	}
	return result
}
