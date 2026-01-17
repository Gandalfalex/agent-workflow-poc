package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListStories(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	stories, err := h.store.ListStories(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "story_list_failed", "failed to list stories")
		return
	}

	items := make([]storyResponse, 0, len(stories))
	for _, story := range stories {
		items = append(items, mapStory(story))
	}
	writeJSON(w, http.StatusOK, storyListResponse{Items: items})
}

func (h *API) CreateStory(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	var req storyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "invalid_story", "title is required")
		return
	}

	story, err := h.store.CreateStory(r.Context(), projectID, store.StoryCreateInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "story_create_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, mapStory(story))
}

func (h *API) GetStory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	storyID := uuid.UUID(id)
	story, err := h.store.GetStory(r.Context(), storyID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "story not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "story_load_failed", "failed to load story")
		return
	}

	writeJSON(w, http.StatusOK, mapStory(story))
}

func (h *API) UpdateStory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	storyID := uuid.UUID(id)
	var req storyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}

	story, err := h.store.UpdateStory(r.Context(), storyID, store.StoryUpdateInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "story not found")
			return
		}
		writeError(w, http.StatusBadRequest, "story_update_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, mapStory(story))
}

func (h *API) DeleteStory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	storyID := uuid.UUID(id)
	if err := h.store.DeleteStory(r.Context(), storyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "story not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "story_delete_failed", "failed to delete story")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ListTicketComments(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID, err := parseOpenapiUUID(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ticket_id", "ticket id must be a UUID")
		return
	}

	comments, err := h.store.ListComments(r.Context(), ticketID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "comment_list_failed", "failed to list comments")
		return
	}

	items := make([]ticketCommentResponse, 0, len(comments))
	for _, comment := range comments {
		items = append(items, mapComment(comment))
	}
	writeJSON(w, http.StatusOK, ticketCommentListResponse{Items: items})
}

func (h *API) AddTicketComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID, err := parseOpenapiUUID(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_ticket_id", "ticket id must be a UUID")
		return
	}

	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	authorID, err := uuid.Parse(user.ID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "user id must be a UUID")
		return
	}

	var req ticketCommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
		return
	}
	if strings.TrimSpace(req.Message) == "" {
		writeError(w, http.StatusBadRequest, "invalid_comment", "message is required")
		return
	}

	comment, err := h.store.CreateComment(r.Context(), ticketID, store.CommentCreateInput{
		AuthorID:   authorID,
		AuthorName: user.Name,
		Message:    req.Message,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "comment_create_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, mapComment(comment))
}

func (h *API) DeleteTicketComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, commentId openapi_types.UUID) {
	commentID, err := parseOpenapiUUID(commentId)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_comment_id", "comment id must be a UUID")
		return
	}

	if err := h.store.DeleteComment(r.Context(), commentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "comment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "comment_delete_failed", "failed to delete comment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
