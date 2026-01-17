package httpapi

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	value := chi.URLParam(r, name)
	return uuid.Parse(value)
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

func parseOptionalUUID(value *openapi_types.UUID) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	id := uuid.UUID(*value)
	return &id, nil
}
