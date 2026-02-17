package httpapi

import (
	"net/http"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListStories(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectID) {
		return
	}
	stories, err := h.store.ListStories(r.Context(), projectID)
	if handleListError(w, r, err, "stories", "story_list") {
		return
	}

	writeJSON(w, http.StatusOK, storyListResponse{Items: mapSlice(stories, mapStory)})
}

func (h *API) CreateStory(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleContributor) {
		return
	}
	req, ok := decodeJSON[storyCreateRequest](w, r, "story_create")
	if !ok {
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
	if handleDBErrorWithCode(w, r, err, "story", "story_create", "story_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapStory(story))
}

func (h *API) GetStory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	storyID := uuid.UUID(id)
	story, err := h.store.GetStory(r.Context(), storyID)
	if handleDBError(w, r, err, "story", "story_load") {
		return
	}

	// Check project access after fetching - return same error to avoid information disclosure
	if !h.requireProjectAccess(w, r, story.ProjectID) {
		return
	}

	writeJSON(w, http.StatusOK, mapStory(story))
}

func (h *API) UpdateStory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	storyID := uuid.UUID(id)

	existing, err := h.store.GetStory(r.Context(), storyID)
	if handleDBError(w, r, err, "story", "story_load") {
		return
	}
	if !h.requireProjectRole(w, r, existing.ProjectID, roleContributor) {
		return
	}

	req, ok := decodeJSON[storyUpdateRequest](w, r, "story_update")
	if !ok {
		return
	}

	story, err := h.store.UpdateStory(r.Context(), storyID, store.StoryUpdateInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if handleDBErrorWithCode(w, r, err, "story", "story_update", "story_update_failed") {
		return
	}

	writeJSON(w, http.StatusOK, mapStory(story))
}

func (h *API) DeleteStory(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	storyID := uuid.UUID(id)
	story, err := h.store.GetStory(r.Context(), storyID)
	if handleDBError(w, r, err, "story", "story_load") {
		return
	}

	if !h.requireProjectRole(w, r, story.ProjectID, roleContributor) {
		return
	}

	if err := h.store.DeleteStory(r.Context(), storyID); handleDeleteError(w, r, err, "story", "story_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ListTicketComments(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	comments, err := h.store.ListComments(r.Context(), ticketID)
	if handleListError(w, r, err, "comments", "comment_list") {
		return
	}

	writeJSON(w, http.StatusOK, ticketCommentListResponse{Items: mapSlice(comments, mapComment)})
}

func (h *API) AddTicketComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}

	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	authorID, err := uuid.Parse(user.ID)
	if err != nil {
		logRequestError(r, "comment_create_invalid_user_id", err)
		writeError(w, http.StatusBadRequest, "invalid_user_id", "user id must be a UUID")
		return
	}

	req, ok := decodeJSON[ticketCommentCreateRequest](w, r, "comment_create")
	if !ok {
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
	if handleDBErrorWithCode(w, r, err, "comment", "comment_create", "comment_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapComment(comment))
}

func (h *API) DeleteTicketComment(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, commentId openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}

	commentID := uuid.UUID(commentId)
	if err := h.store.DeleteComment(r.Context(), commentID); handleDeleteError(w, r, err, "comment", "comment_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
