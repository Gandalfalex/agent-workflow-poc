package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"ticketing-system/backend/internal/auth"
	"ticketing-system/backend/internal/store"
	"ticketing-system/backend/internal/webhook"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Store interface {
	Ping(ctx context.Context) error
	ListProjects(ctx context.Context) ([]store.Project, error)
	ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]store.Project, error)
	ListProjectIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetProject(ctx context.Context, id uuid.UUID) (store.Project, error)
	CreateProject(ctx context.Context, input store.ProjectCreateInput) (store.Project, error)
	UpdateProject(ctx context.Context, id uuid.UUID, input store.ProjectUpdateInput) (store.Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	ListProjectGroups(ctx context.Context, projectID uuid.UUID) ([]store.ProjectGroup, error)
	AddProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID, role string) (store.ProjectGroup, error)
	UpdateProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID, role string) (store.ProjectGroup, error)
	DeleteProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID) error
	ListGroups(ctx context.Context) ([]store.Group, error)
	GetGroup(ctx context.Context, id uuid.UUID) (store.Group, error)
	CreateGroup(ctx context.Context, input store.GroupCreateInput) (store.Group, error)
	UpdateGroup(ctx context.Context, id uuid.UUID, input store.GroupUpdateInput) (store.Group, error)
	DeleteGroup(ctx context.Context, id uuid.UUID) error
	ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]store.GroupMember, error)
	AddGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (store.GroupMember, error)
	DeleteGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error
	UpsertUser(ctx context.Context, input store.UserUpsertInput) error
	ListUsers(ctx context.Context, query string) ([]store.UserSummary, error)
	ListStories(ctx context.Context, projectID uuid.UUID) ([]store.Story, error)
	GetStory(ctx context.Context, id uuid.UUID) (store.Story, error)
	CreateStory(ctx context.Context, projectID uuid.UUID, input store.StoryCreateInput) (store.Story, error)
	UpdateStory(ctx context.Context, id uuid.UUID, input store.StoryUpdateInput) (store.Story, error)
	DeleteStory(ctx context.Context, id uuid.UUID) error
	ListComments(ctx context.Context, ticketID uuid.UUID) ([]store.Comment, error)
	CreateComment(ctx context.Context, ticketID uuid.UUID, input store.CommentCreateInput) (store.Comment, error)
	DeleteComment(ctx context.Context, id uuid.UUID) error
	ListWorkflowStates(ctx context.Context, projectID uuid.UUID) ([]store.WorkflowState, error)
	ReplaceWorkflowStates(ctx context.Context, projectID uuid.UUID, inputs []store.WorkflowStateInput) ([]store.WorkflowState, error)
	ListTickets(ctx context.Context, filter store.TicketFilter) ([]store.Ticket, int, error)
	ListTicketsForBoard(ctx context.Context, projectID uuid.UUID) ([]store.Ticket, error)
	GetTicket(ctx context.Context, id uuid.UUID) (store.Ticket, error)
	CreateTicket(ctx context.Context, projectID uuid.UUID, input store.TicketCreateInput) (store.Ticket, error)
	UpdateTicket(ctx context.Context, id uuid.UUID, input store.TicketUpdateInput) (store.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
	ListWebhooks(ctx context.Context, projectID uuid.UUID) ([]store.Webhook, error)
	GetWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) (store.Webhook, error)
	CreateWebhook(ctx context.Context, projectID uuid.UUID, input store.WebhookCreateInput) (store.Webhook, error)
	UpdateWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID, input store.WebhookUpdateInput) (store.Webhook, error)
	DeleteWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) error
}

type Authenticator interface {
	Login(ctx context.Context, username, password string) (auth.User, auth.TokenSet, error)
	Verify(ctx context.Context, token string) (auth.User, error)
	ListUsers(ctx context.Context) ([]auth.User, error)
}

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, projectID uuid.UUID, event string, data any)
	Test(ctx context.Context, hook store.Webhook, event string, data any) (webhook.Result, error)
}

type API struct {
	Unimplemented
	store                     Store
	auth                      Authenticator
	webhooks                  WebhookDispatcher
	cookieName                string
	cookieTTL                 time.Duration
	cookieSecure              bool
	allowedOrigins            []string
	defaultProjectID          openapi_types.UUID
	defaultProjectKey         ProjectKey
	defaultProjectName        string
	defaultProjectDescription *string
	defaultProjectCreatedAt   time.Time
	defaultProjectUpdatedAt   time.Time
}

func NewHandler(st Store, authClient Authenticator, webhookDispatcher WebhookDispatcher, opts HandlerOptions) *API {
	cookieName := opts.CookieName
	if cookieName == "" {
		cookieName = "ticketing_session"
	}

	ttl := opts.CookieTTL
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	defaultProjectID := uuid.Nil
	if opts.DefaultProjectID != "" {
		if parsed, err := uuid.Parse(opts.DefaultProjectID); err == nil {
			defaultProjectID = parsed
		}
	}

	defaultProjectKey := opts.DefaultProjectKey
	if defaultProjectKey == "" {
		defaultProjectKey = "DEMO"
	}

	defaultProjectName := opts.DefaultProjectName
	if defaultProjectName == "" {
		defaultProjectName = "Default Project"
	}

	now := time.Now()

	return &API{
		store:                     st,
		auth:                      authClient,
		webhooks:                  webhookDispatcher,
		cookieName:                cookieName,
		cookieTTL:                 ttl,
		cookieSecure:              opts.CookieSecure,
		allowedOrigins:            opts.AllowedOrigins,
		defaultProjectID:          defaultProjectID,
		defaultProjectKey:         ProjectKey(defaultProjectKey),
		defaultProjectName:        defaultProjectName,
		defaultProjectDescription: opts.DefaultProjectDescription,
		defaultProjectCreatedAt:   now,
		defaultProjectUpdatedAt:   now,
	}
}

type HandlerOptions struct {
	CookieName                string
	CookieSecure              bool
	CookieTTL                 time.Duration
	AllowedOrigins            []string
	DefaultProjectID          string
	DefaultProjectKey         string
	DefaultProjectName        string
	DefaultProjectDescription *string
}

func (h *API) projectFor(projectID openapi_types.UUID) Project {
	if projectID == uuid.Nil {
		projectID = h.defaultProjectID
	}
	return Project{
		Id:          projectID,
		Key:         h.defaultProjectKey,
		Name:        h.defaultProjectName,
		Description: h.defaultProjectDescription,
		CreatedAt:   h.defaultProjectCreatedAt,
		UpdatedAt:   h.defaultProjectUpdatedAt,
	}
}

func parseOpenapiUUID(value openapi_types.UUID) (uuid.UUID, error) {
	return uuid.UUID(value), nil
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if !isAdmin(r.Context()) {
		writeError(w, http.StatusForbidden, "forbidden", "admin access required")
		return false
	}
	return true
}

func isAdmin(ctx context.Context) bool {
	user, ok := authUser(ctx)
	if !ok {
		return false
	}
	for _, role := range user.Roles {
		if role == "admin" {
			return true
		}
	}
	return false
}

func toWebhookEventStrings(events []WebhookEvent) []string {
	out := make([]string, 0, len(events))
	for _, event := range events {
		out = append(out, string(event))
	}
	return out
}

func (h *API) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "db_unavailable", "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (h *API) Login(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[loginRequest](w, r, "login")
	if !ok {
		return
	}
	identifier := ""
	if req.Identifier != nil {
		identifier = strings.TrimSpace(*req.Identifier)
	} else if req.Email != nil {
		identifier = strings.TrimSpace(string(*req.Email))
	} else if req.Username != nil {
		identifier = strings.TrimSpace(*req.Username)
	}
	if identifier == "" || req.Password == "" {
		logRequestError(r, "login_missing_credentials", nil)
		writeError(w, http.StatusBadRequest, "invalid_credentials", "identifier and password required")
		return
	}

	user, tokenSet, err := h.auth.Login(r.Context(), identifier, req.Password)
	if err != nil {
		logRequestError(r, "login_failed identifier="+identifier, err)
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		return
	}
	h.syncUser(r, user)

	maxAge := int(h.cookieTTL.Seconds())
	if tokenSet.ExpiresIn > 0 {
		maxAge = tokenSet.ExpiresIn
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    tokenSet.AccessToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, loginResponse{User: mapUser(user)})
}

func (h *API) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *API) ListUsers(w http.ResponseWriter, r *http.Request, params ListUsersParams) {
	query := strings.TrimSpace(derefString(params.Q))
	users, err := h.store.ListUsers(r.Context(), query)
	if handleListError(w, r, err, "users", "user_list") {
		return
	}

	items := mapSlice(users, func(user store.UserSummary) userSummary {
		var email *openapi_types.Email
		if strings.TrimSpace(user.Email) != "" {
			parsed := openapi_types.Email(user.Email)
			email = &parsed
		}
		return userSummary{
			Id:    toOpenapiUUID(user.ID),
			Name:  user.Name,
			Email: email,
		}
	})
	writeJSON(w, http.StatusOK, UserListResponse{Items: items})
}

func (h *API) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.listProjectsForCurrentUser(r.Context())
	if handleListError(w, r, err, "projects", "project_list") {
		return
	}

	items := mapSlice(projects, mapProject)
	writeJSON(w, http.StatusOK, projectListResponse{Items: items})
}

func (h *API) CreateProject(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	req, ok := decodeJSON[projectCreateRequest](w, r, "project_create")
	if !ok {
		return
	}
	if req.Key == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_project", "key and name are required")
		return
	}

	project, err := h.store.CreateProject(r.Context(), store.ProjectCreateInput{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	})
	if handleDBErrorWithCode(w, r, err, "project", "project_create", "project_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapProject(project))
}

func (h *API) GetProject(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectID) {
		return
	}
	project, err := h.store.GetProject(r.Context(), projectID)
	if handleDBError(w, r, err, "project", "project_load") {
		return
	}

	writeJSON(w, http.StatusOK, mapProject(project))
}

func (h *API) UpdateProject(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	projectID := uuid.UUID(projectId)
	req, ok := decodeJSON[projectUpdateRequest](w, r, "project_update")
	if !ok {
		return
	}

	project, err := h.store.UpdateProject(r.Context(), projectID, store.ProjectUpdateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if handleDBErrorWithCode(w, r, err, "project", "project_update", "project_update_failed") {
		return
	}

	writeJSON(w, http.StatusOK, mapProject(project))
}

func (h *API) DeleteProject(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	projectID := uuid.UUID(projectId)
	if err := h.store.DeleteProject(r.Context(), projectID); handleDeleteError(w, r, err, "project", "project_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ListProjectGroups(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	items, err := h.store.ListProjectGroups(r.Context(), projectID)
	if handleListError(w, r, err, "project groups", "project_group_list") {
		return
	}
	writeJSON(w, http.StatusOK, projectGroupListResponse{Items: mapSlice(items, mapProjectGroup)})
}

func (h *API) AddProjectGroup(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	projectID := uuid.UUID(projectId)
	req, ok := decodeJSON[projectGroupCreateRequest](w, r, "project_group_create")
	if !ok {
		return
	}

	groupID := uuid.UUID(req.GroupId)
	item, err := h.store.AddProjectGroup(r.Context(), projectID, groupID, string(req.Role))
	if handleDBErrorWithCode(w, r, err, "project group", "project_group_create", "project_group_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapProjectGroup(item))
}

func (h *API) UpdateProjectGroup(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, groupId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	projectID := uuid.UUID(projectId)
	groupUUID := uuid.UUID(groupId)
	req, ok := decodeJSON[projectGroupUpdateRequest](w, r, "project_group_update")
	if !ok {
		return
	}

	item, err := h.store.UpdateProjectGroup(r.Context(), projectID, groupUUID, string(req.Role))
	if handleDBErrorWithCode(w, r, err, "project group", "project_group_update", "project_group_update_failed") {
		return
	}

	writeJSON(w, http.StatusOK, mapProjectGroup(item))
}

func (h *API) DeleteProjectGroup(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, groupId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	projectID := uuid.UUID(projectId)
	groupUUID := uuid.UUID(groupId)
	if err := h.store.DeleteProjectGroup(r.Context(), projectID, groupUUID); handleDeleteError(w, r, err, "project group", "project_group_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.store.ListGroups(r.Context())
	if handleListError(w, r, err, "groups", "group_list") {
		return
	}
	writeJSON(w, http.StatusOK, groupListResponse{Items: mapSlice(groups, mapGroup)})
}

func (h *API) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	req, ok := decodeJSON[groupCreateRequest](w, r, "group_create")
	if !ok {
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_group", "name is required")
		return
	}

	group, err := h.store.CreateGroup(r.Context(), store.GroupCreateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if handleDBErrorWithCode(w, r, err, "group", "group_create", "group_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapGroup(group))
}

func (h *API) GetGroup(w http.ResponseWriter, r *http.Request, groupId openapi_types.UUID) {
	groupID := uuid.UUID(groupId)
	group, err := h.store.GetGroup(r.Context(), groupID)
	if handleDBError(w, r, err, "group", "group_load") {
		return
	}
	writeJSON(w, http.StatusOK, mapGroup(group))
}

func (h *API) UpdateGroup(w http.ResponseWriter, r *http.Request, groupId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	groupID := uuid.UUID(groupId)
	req, ok := decodeJSON[groupUpdateRequest](w, r, "group_update")
	if !ok {
		return
	}

	group, err := h.store.UpdateGroup(r.Context(), groupID, store.GroupUpdateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if handleDBErrorWithCode(w, r, err, "group", "group_update", "group_update_failed") {
		return
	}

	writeJSON(w, http.StatusOK, mapGroup(group))
}

func (h *API) DeleteGroup(w http.ResponseWriter, r *http.Request, groupId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	groupID := uuid.UUID(groupId)
	if err := h.store.DeleteGroup(r.Context(), groupID); handleDeleteError(w, r, err, "group", "group_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ListGroupMembers(w http.ResponseWriter, r *http.Request, groupId openapi_types.UUID) {
	groupID := uuid.UUID(groupId)
	members, err := h.store.ListGroupMembers(r.Context(), groupID)
	if handleListError(w, r, err, "group members", "group_member_list") {
		return
	}
	writeJSON(w, http.StatusOK, groupMemberListResponse{Items: mapSlice(members, mapGroupMember)})
}

func (h *API) AddGroupMember(w http.ResponseWriter, r *http.Request, groupId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	groupID := uuid.UUID(groupId)
	req, ok := decodeJSON[groupMemberCreateRequest](w, r, "group_member_create")
	if !ok {
		return
	}

	userID := uuid.UUID(req.UserId)
	member, err := h.store.AddGroupMember(r.Context(), groupID, userID)
	if err != nil {
		// Check for user not found error
		if err.Error() == "user not found" {
			logRequestError(r, "group_member_create_user_not_found", err)
			writeError(w, http.StatusNotFound, "user_not_found", "user not found")
			return
		}
		// Handle other database errors
		if handleDBErrorWithCode(w, r, err, "group member", "group_member_create", "group_member_create_failed") {
			return
		}
	}

	writeJSON(w, http.StatusCreated, mapGroupMember(member))
}

func (h *API) DeleteGroupMember(w http.ResponseWriter, r *http.Request, groupId openapi_types.UUID, userId openapi_types.UUID) {
	if !requireAdmin(w, r) {
		return
	}
	groupID := uuid.UUID(groupId)
	userUUID := uuid.UUID(userId)
	if err := h.store.DeleteGroupMember(r.Context(), groupID, userUUID); handleDeleteError(w, r, err, "group member", "group_member_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) GetBoard(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	project, err := h.store.GetProject(r.Context(), projectUUID)
	if handleDBError(w, r, err, "project", "project_load") {
		return
	}
	states, err := h.store.ListWorkflowStates(r.Context(), projectUUID)
	if handleListError(w, r, err, "workflow states", "workflow_load") {
		return
	}
	tickets, err := h.store.ListTicketsForBoard(r.Context(), projectUUID)
	if handleListError(w, r, err, "tickets", "ticket_load") {
		return
	}

	writeJSON(w, http.StatusOK, boardResponse{
		Project: mapProject(project),
		States:  mapWorkflowStates(states, projectId),
		Tickets: mapSlice(tickets, mapTicket),
	})
}

func (h *API) ListTickets(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params ListTicketsParams) {
	filter := store.TicketFilter{
		ProjectID: uuid.UUID(projectId),
		Query:     derefString(params.Q),
		Limit:     derefInt(params.Limit, 50),
		Offset:    derefInt(params.Offset, 0),
	}

	if !h.requireProjectAccess(w, r, filter.ProjectID) {
		return
	}

	if params.StateId != nil {
		id, err := parseOpenapiUUID(*params.StateId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_state_id", "stateId must be a UUID")
			return
		}
		filter.StateID = &id
	}

	if params.AssigneeId != nil {
		id, err := parseOpenapiUUID(*params.AssigneeId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_assignee_id", "assigneeId must be a UUID")
			return
		}
		filter.AssigneeID = &id
	}

	tickets, total, err := h.store.ListTickets(r.Context(), filter)
	if handleListError(w, r, err, "tickets", "ticket_list") {
		return
	}

	writeJSON(w, http.StatusOK, ticketListResponse{Items: mapSlice(tickets, mapTicket), Total: total})
}

func (h *API) CreateTicket(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	req, ok := decodeJSON[ticketCreateRequest](w, r, "ticket_create")
	if !ok {
		return
	}
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "invalid_ticket", "title is required")
		return
	}

	var stateID *uuid.UUID
	if req.StateId != nil {
		id, err := parseOpenapiUUID(*req.StateId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_state_id", "stateId must be a UUID")
			return
		}
		stateID = &id
	}

	var assigneeID *uuid.UUID
	if req.AssigneeId != nil {
		id, err := parseOpenapiUUID(*req.AssigneeId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_assignee_id", "assigneeId must be a UUID")
			return
		}
		assigneeID = &id
	}

	priority := ""
	if req.Priority != nil {
		priority = string(*req.Priority)
	}

	ticketType := ""
	if req.Type != nil {
		ticketType = string(*req.Type)
	}

	var storyID *uuid.UUID
	if req.StoryId != nil {
		id, err := parseOpenapiUUID(*req.StoryId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_story_id", "storyId must be a UUID")
			return
		}
		storyID = &id
	}

	ticket, err := h.store.CreateTicket(r.Context(), projectUUID, store.TicketCreateInput{
		Title:       req.Title,
		Description: derefString(req.Description),
		Type:        ticketType,
		StoryID:     storyID,
		StateID:     stateID,
		AssigneeID:  assigneeID,
		Priority:    priority,
	})
	if handleDBErrorWithCode(w, r, err, "ticket", "ticket_create", "ticket_create_failed") {
		return
	}

	response := mapTicket(ticket)
	if h.webhooks != nil {
		h.webhooks.Dispatch(r.Context(), projectUUID, "ticket.created", map[string]any{
			"ticket": response,
		})
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *API) GetTicket(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}

	writeJSON(w, http.StatusOK, mapTicket(ticket))
}

func (h *API) UpdateTicket(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	req, ok := decodeJSON[ticketUpdateRequest](w, r, "ticket_update")
	if !ok {
		return
	}

	var previous *store.Ticket
	if h.webhooks != nil {
		if current, err := h.store.GetTicket(r.Context(), ticketID); err == nil {
			previous = &current
		}
	}

	input := store.TicketUpdateInput{
		Title:       req.Title,
		Description: req.Description,
	}
	if req.Position != nil {
		position := float64(*req.Position)
		input.Position = &position
	}
	if req.StateId != nil {
		stateID, err := parseOpenapiUUID(*req.StateId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_state_id", "stateId must be a UUID")
			return
		}
		input.StateID = &stateID
	}
	if req.AssigneeId != nil {
		assigneeID, err := parseOpenapiUUID(*req.AssigneeId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_assignee_id", "assigneeId must be a UUID")
			return
		}
		input.AssigneeID = &assigneeID
	}
	if req.Priority != nil {
		priority := string(*req.Priority)
		input.Priority = &priority
	}
	if req.Type != nil {
		ticketType := string(*req.Type)
		input.Type = &ticketType
	}
	if req.StoryId != nil {
		storyID, err := parseOpenapiUUID(*req.StoryId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_story_id", "storyId must be a UUID")
			return
		}
		input.StoryID = &storyID
	}

	ticket, err := h.store.UpdateTicket(r.Context(), ticketID, input)
	if handleDBErrorWithCode(w, r, err, "ticket", "ticket_update", "ticket_update_failed") {
		return
	}

	response := mapTicket(ticket)
	projectUUID := uuid.UUID(response.ProjectId)
	if h.webhooks != nil {
		h.webhooks.Dispatch(r.Context(), projectUUID, "ticket.updated", map[string]any{
			"ticket": response,
		})
		if previous != nil && previous.StateID != ticket.StateID {
			h.webhooks.Dispatch(r.Context(), projectUUID, "ticket.state_changed", map[string]any{
				"ticket":      response,
				"fromStateId": previous.StateID.String(),
				"toStateId":   ticket.StateID.String(),
			})
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *API) DeleteTicket(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)

	var deletedTicket *store.Ticket
	if ticket, err := h.store.GetTicket(r.Context(), ticketID); err == nil {
		deletedTicket = &ticket
	}

	if err := h.store.DeleteTicket(r.Context(), ticketID); handleDeleteError(w, r, err, "ticket", "ticket_delete") {
		return
	}

	if deletedTicket != nil && h.webhooks != nil {
		projectUUID := uuid.UUID(deletedTicket.ProjectID)
		h.webhooks.Dispatch(r.Context(), projectUUID, "ticket.deleted", map[string]any{
			"ticket": mapTicket(*deletedTicket),
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *API) GetWorkflow(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	states, err := h.store.ListWorkflowStates(r.Context(), projectUUID)
	if handleListError(w, r, err, "workflow states", "workflow_load") {
		return
	}
	writeJSON(w, http.StatusOK, workflowResponse{States: mapWorkflowStates(states, projectId)})
}

func (h *API) UpdateWorkflow(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	req, ok := decodeJSON[workflowUpdateRequest](w, r, "workflow_update")
	if !ok {
		return
	}
	if len(req.States) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_workflow", "states are required")
		return
	}

	inputs := make([]store.WorkflowStateInput, 0, len(req.States))
	for _, state := range req.States {
		id, err := parseOptionalUUID(state.Id)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_state_id", "state id must be a UUID")
			return
		}
		inputs = append(inputs, store.WorkflowStateInput{
			ID:        id,
			Name:      state.Name,
			Order:     state.Order,
			IsDefault: state.IsDefault,
			IsClosed:  state.IsClosed,
		})
	}

	states, err := h.store.ReplaceWorkflowStates(r.Context(), projectUUID, inputs)
	if handleListError(w, r, err, "workflow states", "workflow_update") {
		return
	}

	writeJSON(w, http.StatusOK, workflowResponse{States: mapWorkflowStates(states, projectId)})
}

func (h *API) ListWebhooks(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	items, err := h.store.ListWebhooks(r.Context(), projectUUID)
	if handleListError(w, r, err, "webhooks", "webhook_list") {
		return
	}

	writeJSON(w, http.StatusOK, webhookListResponse{Items: mapSlice(items, func(hook store.Webhook) webhookResponse {
		return mapWebhook(hook, projectId)
	})})
}

func (h *API) CreateWebhook(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	req, ok := decodeJSON[webhookCreateRequest](w, r, "webhook_create")
	if !ok {
		return
	}
	if req.Url == "" || len(req.Events) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_webhook", "url and events are required")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	hook, err := h.store.CreateWebhook(r.Context(), projectUUID, store.WebhookCreateInput{
		URL:     req.Url,
		Events:  toWebhookEventStrings(req.Events),
		Enabled: enabled,
		Secret:  req.Secret,
	})
	if handleDBErrorWithCode(w, r, err, "webhook", "webhook_create", "webhook_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, mapWebhook(hook, projectId))
}

func (h *API) GetWebhook(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, id openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	webhookID := uuid.UUID(id)
	hook, err := h.store.GetWebhook(r.Context(), projectUUID, webhookID)
	if handleDBError(w, r, err, "webhook", "webhook_load") {
		return
	}

	writeJSON(w, http.StatusOK, mapWebhook(hook, projectId))
}

func (h *API) UpdateWebhook(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, id openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	webhookID := uuid.UUID(id)
	req, ok := decodeJSON[webhookUpdateRequest](w, r, "webhook_update")
	if !ok {
		return
	}

	var events *[]string
	if req.Events != nil {
		values := toWebhookEventStrings(*req.Events)
		events = &values
	}

	hook, err := h.store.UpdateWebhook(r.Context(), projectUUID, webhookID, store.WebhookUpdateInput{
		URL:     req.Url,
		Events:  events,
		Enabled: req.Enabled,
		Secret:  req.Secret,
	})
	if handleDBErrorWithCode(w, r, err, "webhook", "webhook_update", "webhook_update_failed") {
		return
	}

	writeJSON(w, http.StatusOK, mapWebhook(hook, projectId))
}

func (h *API) DeleteWebhook(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, id openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	webhookID := uuid.UUID(id)
	if err := h.store.DeleteWebhook(r.Context(), projectUUID, webhookID); handleDeleteError(w, r, err, "webhook", "webhook_delete") {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *API) TestWebhook(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, id openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	webhookID := uuid.UUID(id)
	hook, err := h.store.GetWebhook(r.Context(), projectUUID, webhookID)
	if handleDBError(w, r, err, "webhook", "webhook_load") {
		return
	}

	var req webhookTestRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logRequestError(r, "webhook_test_invalid_json", err)
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid json body")
			return
		}
	}

	event := "ticket.updated"
	if req.Event != nil {
		event = string(*req.Event)
	}

	data := map[string]any{}
	if req.Payload != nil {
		data = *req.Payload
	}
	if req.TicketId != nil {
		data["ticketId"] = req.TicketId.String()
	}

	if h.webhooks == nil {
		writeError(w, http.StatusServiceUnavailable, "webhooks_unavailable", "webhook dispatcher unavailable")
		return
	}

	result, _ := h.webhooks.Test(r.Context(), hook, event, data)
	var statusCode *int
	var responseBody *string
	if result.StatusCode != 0 {
		statusCode = &result.StatusCode
	}
	if result.ResponseBody != "" {
		responseBody = &result.ResponseBody
	}
	writeJSON(w, http.StatusOK, webhookTestResponse{
		Delivered:    result.Delivered,
		StatusCode:   statusCode,
		ResponseBody: responseBody,
	})
}

func mapUser(user auth.User) userResponse {
	parsedID, err := uuid.Parse(user.ID)
	if err != nil {
		parsedID = uuid.Nil
	}
	return userResponse{
		Id:        parsedID,
		Email:     openapi_types.Email(user.Email),
		Name:      user.Name,
		CreatedAt: time.Now(),
	}
}

func (h *API) currentUserID(ctx context.Context) (uuid.UUID, error) {
	user, ok := authUser(ctx)
	if !ok {
		return uuid.Nil, errors.New("missing user")
	}
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (h *API) listProjectsForCurrentUser(ctx context.Context) ([]store.Project, error) {
	if isAdmin(ctx) {
		return h.store.ListProjects(ctx)
	}
	userID, err := h.currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	return h.store.ListProjectsForUser(ctx, userID)
}

func (h *API) projectIDsForCurrentUser(ctx context.Context) (map[uuid.UUID]struct{}, error) {
	if isAdmin(ctx) {
		projects, err := h.store.ListProjects(ctx)
		if err != nil {
			return nil, err
		}
		resp := make(map[uuid.UUID]struct{}, len(projects))
		for _, project := range projects {
			resp[project.ID] = struct{}{}
		}
		return resp, nil
	}
	userID, err := h.currentUserID(ctx)
	if err != nil {
		return nil, err
	}
	ids, err := h.store.ListProjectIDsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		resp[id] = struct{}{}
	}
	return resp, nil
}

func (h *API) hasProjectAccess(ctx context.Context, projectID uuid.UUID) (bool, error) {
	if isAdmin(ctx) {
		return true, nil
	}
	projectIDs, err := h.projectIDsForCurrentUser(ctx)
	if err != nil {
		return false, err
	}
	_, ok := projectIDs[projectID]
	return ok, nil
}

func (h *API) requireProjectAccess(w http.ResponseWriter, r *http.Request, projectID uuid.UUID) bool {
	ok, err := h.hasProjectAccess(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "project_access_failed", "unable to verify project access")
		return false
	}
	if !ok {
		writeError(w, http.StatusForbidden, "project_access_denied", "project access denied")
		return false
	}
	return true
}

func (h *API) syncUser(r *http.Request, user auth.User) {
	id, err := uuid.Parse(user.ID)
	if err != nil {
		logRequestError(r, "sync_user_invalid_id", err)
		return
	}
	email := strings.TrimSpace(user.Email)
	name := strings.TrimSpace(user.Name)
	if email == "" {
		email = user.ID
	}
	if name == "" {
		name = email
	}
	if err := h.store.UpsertUser(r.Context(), store.UserUpsertInput{
		ID:    id,
		Name:  name,
		Email: email,
	}); err != nil {
		logRequestError(r, "sync_user_failed", err)
	}
}

// SyncUsers syncs all users from Keycloak to the local database
// POST /admin/sync-users
func (h *API) SyncUsers(w http.ResponseWriter, r *http.Request) {
	// Get all users from Keycloak
	keycloakUsers, err := h.auth.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sync_failed", "failed to fetch users from Keycloak: "+err.Error())
		return
	}

	synced := 0

	// Upsert each user into the database
	for _, user := range keycloakUsers {
		id, err := uuid.Parse(user.ID)
		if err != nil {
			continue
		}

		email := strings.TrimSpace(user.Email)
		name := strings.TrimSpace(user.Name)
		if email == "" {
			email = user.ID + "@local"
		}
		if name == "" {
			name = email
		}

		if err := h.store.UpsertUser(r.Context(), store.UserUpsertInput{
			ID:    id,
			Name:  name,
			Email: email,
		}); err == nil {
			synced++
		}
	}

	writeJSON(w, http.StatusOK, SyncUsersResponse{
		Synced: synced,
		Total:  len(keycloakUsers),
	})
}
