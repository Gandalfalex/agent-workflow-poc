package httpapi

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"ticketing-system/backend/internal/auth"
	"ticketing-system/backend/internal/blob"
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
	ListTicketDependencies(ctx context.Context, projectID, ticketID uuid.UUID) ([]store.TicketDependency, error)
	GetTicketDependencyForTicket(ctx context.Context, dependencyID, projectID, ticketID uuid.UUID) (store.TicketDependency, error)
	CreateTicketDependency(ctx context.Context, projectID uuid.UUID, input store.TicketDependencyCreateInput) (store.TicketDependency, error)
	DeleteTicketDependency(ctx context.Context, dependencyID, projectID, ticketID uuid.UUID) error
	GetTicketDependencyGraph(ctx context.Context, projectID uuid.UUID, rootTicketID *uuid.UUID, depth int) (store.TicketDependencyGraph, error)
	CreateTicket(ctx context.Context, projectID uuid.UUID, input store.TicketCreateInput) (store.Ticket, error)
	UpdateTicket(ctx context.Context, id uuid.UUID, input store.TicketUpdateInput) (store.Ticket, error)
	DeleteTicket(ctx context.Context, id uuid.UUID) error
	GetProjectStats(ctx context.Context, projectID uuid.UUID) (store.ProjectStats, error)
	ListSprints(ctx context.Context, projectID uuid.UUID) ([]store.Sprint, error)
	CreateSprint(ctx context.Context, projectID uuid.UUID, input store.SprintCreateInput) (store.Sprint, error)
	GetSprint(ctx context.Context, projectID, sprintID uuid.UUID) (store.Sprint, error)
	AddSprintTickets(ctx context.Context, projectID, sprintID uuid.UUID, ticketIDs []uuid.UUID) (store.Sprint, error)
	RemoveSprintTickets(ctx context.Context, projectID, sprintID uuid.UUID, ticketIDs []uuid.UUID) (store.Sprint, error)
	ListCapacitySettings(ctx context.Context, projectID uuid.UUID) ([]store.CapacitySetting, error)
	ReplaceCapacitySettings(ctx context.Context, projectID uuid.UUID, inputs []store.CapacitySettingInput) ([]store.CapacitySetting, error)
	GetSprintForecastSummary(ctx context.Context, projectID uuid.UUID, sprintID *uuid.UUID, iterations int) (store.SprintForecastSummary, error)
	GetAiTriageSettings(ctx context.Context, projectID uuid.UUID) (store.AiTriageSettings, error)
	UpdateAiTriageSettings(ctx context.Context, projectID uuid.UUID, enabled bool) (store.AiTriageSettings, error)
	CreateAiTriageSuggestion(ctx context.Context, projectID uuid.UUID, input store.AiTriageSuggestionCreateInput) (store.AiTriageSuggestion, error)
	GetAiTriageSuggestion(ctx context.Context, projectID, suggestionID uuid.UUID) (store.AiTriageSuggestion, error)
	CreateAiTriageSuggestionDecision(ctx context.Context, projectID, suggestionID uuid.UUID, input store.AiTriageSuggestionDecisionCreateInput) (store.AiTriageSuggestionDecision, error)
	GetProjectReportingSummary(ctx context.Context, projectID uuid.UUID, from, to time.Time) (store.ProjectReportingSummary, error)
	ListWebhooks(ctx context.Context, projectID uuid.UUID) ([]store.Webhook, error)
	GetWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) (store.Webhook, error)
	CreateWebhook(ctx context.Context, projectID uuid.UUID, input store.WebhookCreateInput) (store.Webhook, error)
	UpdateWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID, input store.WebhookUpdateInput) (store.Webhook, error)
	DeleteWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) error
	ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID) ([]store.WebhookDelivery, error)
	ListAttachments(ctx context.Context, ticketID uuid.UUID) ([]store.Attachment, error)
	GetAttachment(ctx context.Context, id uuid.UUID) (store.Attachment, error)
	CreateAttachment(ctx context.Context, ticketID uuid.UUID, input store.AttachmentCreateInput) (store.Attachment, error)
	DeleteAttachment(ctx context.Context, id uuid.UUID) error
	GetProjectRoleForUser(ctx context.Context, projectID, userID uuid.UUID) (string, error)
	ListActivities(ctx context.Context, ticketID uuid.UUID) ([]store.Activity, error)
	ListProjectActivities(ctx context.Context, projectID uuid.UUID, limit int) ([]store.ProjectActivity, error)
	CreateActivity(ctx context.Context, ticketID uuid.UUID, input store.ActivityCreateInput) error
	CreateTicketWebhookEvent(ctx context.Context, input store.TicketWebhookEventCreateInput) error
	ListTicketWebhookEvents(ctx context.Context, ticketID uuid.UUID) ([]store.TicketWebhookEvent, error)
	CreateNotification(ctx context.Context, input store.NotificationCreateInput) (store.Notification, error)
	ListNotifications(ctx context.Context, filter store.NotificationFilter) ([]store.Notification, error)
	CountUnreadNotifications(ctx context.Context, projectID, userID uuid.UUID) (int, error)
	MarkNotificationRead(ctx context.Context, id, projectID, userID uuid.UUID) (store.Notification, error)
	MarkAllNotificationsRead(ctx context.Context, projectID, userID uuid.UUID) (int, error)
	GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (store.NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, input store.NotificationPreferencesUpdateInput) (store.NotificationPreferences, error)
	ListBoardFilterPresets(ctx context.Context, projectID, ownerID uuid.UUID) ([]store.BoardFilterPreset, error)
	CreateBoardFilterPreset(ctx context.Context, projectID, ownerID uuid.UUID, input store.BoardFilterPresetCreateInput) (store.BoardFilterPreset, error)
	UpdateBoardFilterPreset(ctx context.Context, projectID, ownerID, presetID uuid.UUID, input store.BoardFilterPresetUpdateInput) (store.BoardFilterPreset, error)
	DeleteBoardFilterPreset(ctx context.Context, projectID, ownerID, presetID uuid.UUID) error
	GetSharedBoardFilterPreset(ctx context.Context, projectID uuid.UUID, token string) (store.BoardFilterPreset, error)
	ListTimeEntries(ctx context.Context, ticketID uuid.UUID) ([]store.TimeEntry, int, error)
	CreateTimeEntry(ctx context.Context, ticketID uuid.UUID, input store.TimeEntryCreateInput) (store.TimeEntry, error)
	DeleteTimeEntry(ctx context.Context, id uuid.UUID) error
}

type Authenticator interface {
	Login(ctx context.Context, username, password string) (auth.User, auth.TokenSet, error)
	Verify(ctx context.Context, token string) (auth.User, error)
	ListUsers(ctx context.Context) ([]auth.User, error)
	CreateUser(ctx context.Context, input auth.UserCreateInput) (auth.User, error)
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
	live                      *projectLiveHub
	blob                      blob.ObjectStore
	maxUploadSize             int64
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

	maxUpload := opts.MaxUploadSize
	if maxUpload <= 0 {
		maxUpload = 10 << 20 // 10 MB
	}

	now := time.Now()

	return &API{
		store:                     st,
		auth:                      authClient,
		webhooks:                  webhookDispatcher,
		live:                      newProjectLiveHub(),
		blob:                      opts.BlobStore,
		maxUploadSize:             maxUpload,
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
	BlobStore                 blob.ObjectStore
	MaxUploadSize             int64
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

const (
	roleViewer      = "viewer"
	roleContributor = "contributor"
	roleAdmin       = "admin"
)

func roleRank(role string) int {
	switch role {
	case roleAdmin:
		return 3
	case roleContributor:
		return 2
	case roleViewer:
		return 1
	default:
		return 0
	}
}

func (h *API) requireProjectRole(w http.ResponseWriter, r *http.Request, projectID uuid.UUID, minRole string) bool {
	if isAdmin(r.Context()) {
		return true
	}
	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return false
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "role_check_failed", "invalid user id")
		return false
	}
	role, err := h.store.GetProjectRoleForUser(r.Context(), projectID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "role_check_failed", "unable to verify role")
		return false
	}
	if roleRank(role) < roleRank(minRole) {
		writeError(w, http.StatusForbidden, "insufficient_role", "requires "+minRole+" role or higher")
		return false
	}
	return true
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
		Name:                      req.Name,
		Description:               req.Description,
		DefaultSprintDurationDays: req.DefaultSprintDurationDays,
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
	if !h.requireProjectAccess(w, r, projectID) {
		return
	}
	items, err := h.store.ListProjectGroups(r.Context(), projectID)
	if handleListError(w, r, err, "project groups", "project_group_list") {
		return
	}
	writeJSON(w, http.StatusOK, projectGroupListResponse{Items: mapSlice(items, mapProjectGroup)})
}

func (h *API) AddProjectGroup(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleAdmin) {
		return
	}
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
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleAdmin) {
		return
	}
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
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleAdmin) {
		return
	}
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
	if params.Blocked != nil {
		filter.Blocked = params.Blocked
	}

	tickets, total, err := h.store.ListTickets(r.Context(), filter)
	if handleListError(w, r, err, "tickets", "ticket_list") {
		return
	}

	writeJSON(w, http.StatusOK, ticketListResponse{Items: mapSlice(tickets, mapTicket), Total: total})
}

func (h *API) CreateTicket(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectUUID, roleContributor) {
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

	storyID, err := parseOpenapiUUID(req.StoryId)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_story_id", "storyId is required and must be a UUID")
		return
	}

	ticket, err := h.store.CreateTicket(r.Context(), projectUUID, store.TicketCreateInput{
		Title:               req.Title,
		Description:         derefString(req.Description),
		Type:                ticketType,
		StoryID:             storyID,
		StateID:             stateID,
		AssigneeID:          assigneeID,
		Priority:            priority,
		IncidentEnabled:     req.IncidentEnabled != nil && *req.IncidentEnabled,
		IncidentSeverity:    mapStringPtr(req.IncidentSeverity, func(v TicketIncidentSeverity) string { return string(v) }),
		IncidentImpact:      req.IncidentImpact,
		IncidentCommanderID: parseOpenapiUUIDPtr(req.IncidentCommanderId),
		StoryPoints:         req.StoryPoints,
		TimeEstimate:        req.TimeEstimate,
	})
	if handleDBErrorWithCode(w, r, err, "ticket", "ticket_create", "ticket_create_failed") {
		return
	}

	response := mapTicket(ticket)
	if actor, ok := authUser(r.Context()); ok {
		if actorID, err := uuid.Parse(actor.ID); err == nil {
			h.notifyAssignment(r, store.Ticket{}, ticket, actorID, actor.Name)
		}
	}
	h.dispatchTicketWebhook(r.Context(), projectUUID, ticket.ID, "ticket.created", map[string]any{"ticket": response})
	h.publishProjectLiveEvent(projectUUID, projectEventBoardRefresh, map[string]any{
		"reason": "ticket.created",
	})
	h.publishProjectLiveEvent(projectUUID, projectEventActivityChanged, map[string]any{
		"reason": "ticket.created",
	})

	writeJSON(w, http.StatusCreated, response)
}

func (h *API) BulkTicketOperation(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	req, ok := decodeJSON[BulkTicketOperationRequest](w, r, "ticket_bulk")
	if !ok {
		return
	}
	if len(req.TicketIds) == 0 {
		writeError(w, http.StatusBadRequest, "invalid_bulk_request", "ticketIds must not be empty")
		return
	}

	switch req.Action {
	case BulkTicketActionMoveState:
		if req.StateId == nil {
			writeError(w, http.StatusBadRequest, "invalid_bulk_request", "stateId is required for move_state")
			return
		}
	case BulkTicketActionAssign:
		if req.AssigneeId == nil {
			writeError(w, http.StatusBadRequest, "invalid_bulk_request", "assigneeId is required for assign")
			return
		}
	case BulkTicketActionSetPriority:
		if req.Priority == nil {
			writeError(w, http.StatusBadRequest, "invalid_bulk_request", "priority is required for set_priority")
			return
		}
	case BulkTicketActionDelete:
		// no-op
	default:
		writeError(w, http.StatusBadRequest, "invalid_bulk_request", "unsupported action")
		return
	}

	results := make([]BulkTicketOperationResult, 0, len(req.TicketIds))
	successCount := 0
	errorCount := 0

	var actorID *uuid.UUID
	var actorName string
	if actor, ok := authUser(r.Context()); ok {
		actorName = actor.Name
		if parsed, err := uuid.Parse(actor.ID); err == nil {
			actorID = &parsed
		}
	}

	for _, ticketIDRaw := range req.TicketIds {
		ticketID := uuid.UUID(ticketIDRaw)
		result := BulkTicketOperationResult{TicketId: ticketIDRaw}

		ticket, err := h.store.GetTicket(r.Context(), ticketID)
		if err != nil {
			errorCount++
			code := "not_found"
			msg := "ticket not found"
			result.Success = false
			result.ErrorCode = &code
			result.Message = &msg
			results = append(results, result)
			continue
		}
		if ticket.ProjectID != projectUUID {
			errorCount++
			code := "project_mismatch"
			msg := "ticket does not belong to project"
			result.Success = false
			result.ErrorCode = &code
			result.Message = &msg
			results = append(results, result)
			continue
		}
		if !h.hasBulkOperationRole(r.Context(), ticket.ProjectID) {
			errorCount++
			code := "insufficient_role"
			msg := "requires contributor role or higher"
			result.Success = false
			result.ErrorCode = &code
			result.Message = &msg
			results = append(results, result)
			continue
		}

		switch req.Action {
		case BulkTicketActionMoveState:
			stateID := uuid.UUID(*req.StateId)
			updated, err := h.store.UpdateTicket(r.Context(), ticketID, store.TicketUpdateInput{
				StateID: &stateID,
			})
			if err != nil {
				errorCount++
				code := "ticket_update_failed"
				msg := err.Error()
				result.Success = false
				result.ErrorCode = &code
				result.Message = &msg
				results = append(results, result)
				continue
			}
			if actorID != nil {
				h.recordTicketActivities(r.Context(), ticket, updated, *actorID, actorName)
			}
			mapped := mapTicket(updated)
			result.Success = true
			result.Ticket = &mapped
			results = append(results, result)
			successCount++
		case BulkTicketActionAssign:
			assigneeID := uuid.UUID(*req.AssigneeId)
			updated, err := h.store.UpdateTicket(r.Context(), ticketID, store.TicketUpdateInput{
				AssigneeID: &assigneeID,
			})
			if err != nil {
				errorCount++
				code := "ticket_update_failed"
				msg := err.Error()
				result.Success = false
				result.ErrorCode = &code
				result.Message = &msg
				results = append(results, result)
				continue
			}
			if actorID != nil {
				h.recordTicketActivities(r.Context(), ticket, updated, *actorID, actorName)
			}
			mapped := mapTicket(updated)
			result.Success = true
			result.Ticket = &mapped
			results = append(results, result)
			successCount++
		case BulkTicketActionSetPriority:
			priority := string(*req.Priority)
			updated, err := h.store.UpdateTicket(r.Context(), ticketID, store.TicketUpdateInput{
				Priority: &priority,
			})
			if err != nil {
				errorCount++
				code := "ticket_update_failed"
				msg := err.Error()
				result.Success = false
				result.ErrorCode = &code
				result.Message = &msg
				results = append(results, result)
				continue
			}
			if actorID != nil {
				h.recordTicketActivities(r.Context(), ticket, updated, *actorID, actorName)
			}
			mapped := mapTicket(updated)
			result.Success = true
			result.Ticket = &mapped
			results = append(results, result)
			successCount++
		case BulkTicketActionDelete:
			if err := h.store.DeleteTicket(r.Context(), ticketID); err != nil {
				errorCount++
				code := "ticket_delete_failed"
				msg := err.Error()
				result.Success = false
				result.ErrorCode = &code
				result.Message = &msg
				results = append(results, result)
				continue
			}
			if h.webhooks != nil {
				h.webhooks.Dispatch(r.Context(), ticket.ProjectID, "ticket.deleted", map[string]any{
					"ticket": mapTicket(ticket),
				})
			}
			result.Success = true
			results = append(results, result)
			successCount++
		}
	}

	if successCount > 0 {
		h.publishProjectLiveEvent(projectUUID, projectEventBoardRefresh, map[string]any{
			"reason": "tickets.bulk",
			"action": string(req.Action),
		})
		h.publishProjectLiveEvent(projectUUID, projectEventActivityChanged, map[string]any{
			"reason": "tickets.bulk",
			"action": string(req.Action),
		})
	}

	writeJSON(w, http.StatusOK, BulkTicketOperationResponse{
		Action:       req.Action,
		Total:        len(req.TicketIds),
		SuccessCount: successCount,
		ErrorCount:   errorCount,
		Results:      results,
	})
}

func (h *API) hasBulkOperationRole(ctx context.Context, projectID uuid.UUID) bool {
	if isAdmin(ctx) {
		return true
	}
	user, ok := authUser(ctx)
	if !ok {
		return false
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		return false
	}
	role, err := h.store.GetProjectRoleForUser(ctx, projectID, userID)
	if err != nil {
		return false
	}
	return roleRank(role) >= roleRank(roleContributor)
}

func (h *API) GetTicket(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectAccess(w, r, ticket.ProjectID) {
		return
	}

	writeJSON(w, http.StatusOK, mapTicket(ticket))
}

func (h *API) UpdateTicket(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)

	current, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, current.ProjectID, roleContributor) {
		return
	}

	req, ok := decodeJSON[ticketUpdateRequest](w, r, "ticket_update")
	if !ok {
		return
	}

	var previous *store.Ticket
	if h.webhooks != nil {
		previous = &current
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
	if req.IncidentEnabled != nil {
		input.IncidentEnabled = req.IncidentEnabled
	}
	if req.IncidentSeverity != nil {
		incidentSeverity := string(*req.IncidentSeverity)
		input.IncidentSeverity = &incidentSeverity
	}
	if req.IncidentImpact != nil {
		input.IncidentImpact = req.IncidentImpact
	}
	if req.IncidentCommanderId != nil {
		incidentCommanderID, err := parseOpenapiUUID(*req.IncidentCommanderId)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_incident_commander_id", "incidentCommanderId must be a UUID")
			return
		}
		input.IncidentCommanderID = &incidentCommanderID
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
	if req.StoryPoints != nil {
		input.StoryPoints = req.StoryPoints
	}
	if req.TimeEstimate != nil {
		input.TimeEstimate = req.TimeEstimate
	}

	ticket, err := h.store.UpdateTicket(r.Context(), ticketID, input)
	if handleDBErrorWithCode(w, r, err, "ticket", "ticket_update", "ticket_update_failed") {
		return
	}

	var actorID uuid.UUID
	var actorName string
	actorResolved := false
	if actor, ok := authUser(r.Context()); ok {
		if parsedActorID, err := uuid.Parse(actor.ID); err == nil {
			actorID = parsedActorID
			actorName = actor.Name
			actorResolved = true
			h.recordTicketActivities(r.Context(), current, ticket, actorID, actor.Name)
			h.notifyAssignment(r, current, ticket, actorID, actor.Name)
			h.notifyAssigneeTicketUpdate(r, current, ticket, actorID, actor.Name)
		}
	}
	if actorResolved && req.Description != nil && current.Description != ticket.Description {
		h.notifyMentions(r, ticket.ProjectID, ticket, actorID, actorName, ticket.Description)
	}

	response := mapTicket(ticket)
	projectUUID := uuid.UUID(response.ProjectId)
	h.dispatchTicketWebhook(r.Context(), projectUUID, ticket.ID, "ticket.updated", map[string]any{"ticket": response})
	if previous != nil && previous.StateID != ticket.StateID {
		h.dispatchTicketWebhook(r.Context(), projectUUID, ticket.ID, "ticket.state_changed", map[string]any{
			"ticket":      response,
			"fromStateId": previous.StateID.String(),
			"toStateId":   ticket.StateID.String(),
		})
	}
	h.publishProjectLiveEvent(projectUUID, projectEventBoardRefresh, map[string]any{
		"reason": "ticket.updated",
		"id":     ticket.ID.String(),
	})
	h.publishProjectLiveEvent(projectUUID, projectEventActivityChanged, map[string]any{
		"reason": "ticket.updated",
		"id":     ticket.ID.String(),
	})

	writeJSON(w, http.StatusOK, response)
}

func (h *API) DeleteTicket(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)

	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}

	deletedTicket := &ticket

	if err := h.store.DeleteTicket(r.Context(), ticketID); handleDeleteError(w, r, err, "ticket", "ticket_delete") {
		return
	}

	if deletedTicket != nil {
		projectUUID := deletedTicket.ProjectID
		h.dispatchTicketWebhook(r.Context(), projectUUID, deletedTicket.ID, "ticket.deleted", map[string]any{
			"ticket": mapTicket(*deletedTicket),
		})
		h.publishProjectLiveEvent(projectUUID, projectEventBoardRefresh, map[string]any{
			"reason": "ticket.deleted",
			"id":     deletedTicket.ID.String(),
		})
		h.publishProjectLiveEvent(projectUUID, projectEventActivityChanged, map[string]any{
			"reason": "ticket.deleted",
			"id":     deletedTicket.ID.String(),
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
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
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
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
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
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
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
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
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
	if !h.requireProjectRole(w, r, projectUUID, roleAdmin) {
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

func (h *API) ListWebhookDeliveries(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, id openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	webhookID := uuid.UUID(id)
	// Verify webhook belongs to project
	if _, err := h.store.GetWebhook(r.Context(), projectUUID, webhookID); handleDBError(w, r, err, "webhook", "webhook_load") {
		return
	}
	deliveries, err := h.store.ListWebhookDeliveries(r.Context(), webhookID)
	if handleListError(w, r, err, "webhook deliveries", "webhook_delivery_list") {
		return
	}
	writeJSON(w, http.StatusOK, webhookDeliveryListResponse{Items: mapSlice(deliveries, mapWebhookDelivery)})
}

func (h *API) GetProjectStats(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	stats, err := h.store.GetProjectStats(r.Context(), projectUUID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "stats_error", "Failed to load project statistics")
		return
	}
	writeJSON(w, http.StatusOK, mapProjectStats(stats))
}

func (h *API) GetProjectReportingSummary(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params GetProjectReportingSummaryParams) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	from, to, ok := parseReportingRange(params.From, params.To)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_date_range", "`to` must be on or after `from`")
		return
	}

	report, err := h.store.GetProjectReportingSummary(r.Context(), projectUUID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reporting_error", "Failed to load project reporting summary")
		return
	}
	writeJSON(w, http.StatusOK, mapProjectReportingSummary(report))
}

func (h *API) ExportProjectReportingSnapshot(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params ExportProjectReportingSnapshotParams) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	from, to, ok := parseReportingRange(params.From, params.To)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_date_range", "`to` must be on or after `from`")
		return
	}

	report, err := h.store.GetProjectReportingSummary(r.Context(), projectUUID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "reporting_error", "Failed to load project reporting summary")
		return
	}

	format := "json"
	if params.Format != nil {
		format = strings.ToLower(string(*params.Format))
	}

	if format == "csv" {
		content, err := renderProjectReportingCSV(report)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "reporting_export_error", "Failed to render reporting export")
			return
		}
		filename := fmt.Sprintf("project-reporting-%s-to-%s.csv", report.From.Format("2006-01-02"), report.To.Format("2006-01-02"))
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
		return
	}

	writeJSON(w, http.StatusOK, ProjectReportingExportJson{
		GeneratedAt: time.Now().UTC(),
		Summary:     mapProjectReportingSummary(report),
	})
}

func parseReportingRange(fromParam, toParam *openapi_types.Date) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	to := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	from := to.AddDate(0, 0, -13)

	if fromParam != nil {
		parsed := fromParam.Time.UTC()
		from = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	}
	if toParam != nil {
		parsed := toParam.Time.UTC()
		to = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	}

	return from, to, !to.Before(from)
}

func renderProjectReportingCSV(report store.ProjectReportingSummary) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	if err := writer.Write([]string{"section", "date", "label", "value"}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"meta", report.From.Format("2006-01-02"), "from", report.From.Format("2006-01-02")}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"meta", report.To.Format("2006-01-02"), "to", report.To.Format("2006-01-02")}); err != nil {
		return nil, err
	}
	if err := writer.Write([]string{"meta", "", "average_cycle_time_hours", fmt.Sprintf("%.2f", report.AverageCycleTimeHours)}); err != nil {
		return nil, err
	}
	for _, point := range report.ThroughputByDay {
		if err := writer.Write([]string{"throughput_by_day", point.Date.Format("2006-01-02"), "throughput", fmt.Sprintf("%d", point.Value)}); err != nil {
			return nil, err
		}
	}
	for _, point := range report.OpenByState {
		for _, count := range point.Counts {
			if err := writer.Write([]string{"open_by_state", point.Date.Format("2006-01-02"), count.Label, fmt.Sprintf("%d", count.Value)}); err != nil {
				return nil, err
			}
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *API) GetMyProjectRole(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}

	// System admins get admin role on all projects
	if isAdmin(r.Context()) {
		writeJSON(w, http.StatusOK, map[string]string{"role": roleAdmin})
		return
	}

	userID, err := uuid.Parse(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid_user_id", "invalid user id")
		return
	}
	role, err := h.store.GetProjectRoleForUser(r.Context(), projectUUID, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "role_load_failed", "unable to load role")
		return
	}
	if role == "" {
		writeError(w, http.StatusForbidden, "no_access", "no access to this project")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"role": role})
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
	if !requireAdmin(w, r) {
		return
	}

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

// CreateAdminUser creates a user in the identity provider and upserts it in app storage.
// POST /admin/users
func (h *API) CreateAdminUser(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	req, ok := decodeJSON[AdminUserCreateRequest](w, r, "admin_user_create")
	if !ok {
		return
	}
	email := strings.TrimSpace(string(req.Email))
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" || email == "" {
		writeError(w, http.StatusBadRequest, "invalid_user_create", "username, email, and password are required")
		return
	}

	created, err := h.auth.CreateUser(r.Context(), auth.UserCreateInput{
		Username:  req.Username,
		Email:     email,
		FirstName: derefString(req.FirstName),
		LastName:  derefString(req.LastName),
		Password:  req.Password,
	})
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "already exists") || strings.Contains(msg, "409") {
			writeError(w, http.StatusConflict, "user_exists", "user already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "user_create_failed", "failed to create user")
		return
	}

	userID, parseErr := uuid.Parse(created.ID)
	if parseErr != nil {
		writeError(w, http.StatusInternalServerError, "user_create_failed", "identity provider returned invalid user id")
		return
	}

	createdEmail := strings.TrimSpace(created.Email)
	name := strings.TrimSpace(created.Name)
	if createdEmail == "" {
		createdEmail = created.ID + "@local"
	}
	if name == "" {
		name = createdEmail
	}

	if err := h.store.UpsertUser(r.Context(), store.UserUpsertInput{
		ID:    userID,
		Name:  name,
		Email: createdEmail,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "user_create_failed", "failed to persist user")
		return
	}

	var emailValue openapi_types.Email
	emailValue = openapi_types.Email(createdEmail)
	writeJSON(w, http.StatusCreated, UserSummary{
		Id:    toOpenapiUUID(userID),
		Name:  name,
		Email: &emailValue,
	})
}

func (h *API) ListProjectActivities(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params ListProjectActivitiesParams) {
	projectID := uuid.UUID(projectId)
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	activities, err := h.store.ListProjectActivities(r.Context(), projectID, limit)
	if handleListError(w, r, err, "activities", "project_activity_list") {
		return
	}
	writeJSON(w, http.StatusOK, projectActivityListResponse{Items: mapSlice(activities, mapProjectActivity)})
}

func (h *API) ListTicketActivities(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	activities, err := h.store.ListActivities(r.Context(), ticketID)
	if handleListError(w, r, err, "activities", "activity_list") {
		return
	}
	writeJSON(w, http.StatusOK, ticketActivityListResponse{Items: mapSlice(activities, mapActivity)})
}

func (h *API) ListTicketIncidentTimeline(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectAccess(w, r, ticket.ProjectID) {
		return
	}

	activities, _ := h.store.ListActivities(r.Context(), ticketID)
	comments, _ := h.store.ListComments(r.Context(), ticketID)
	webhookEvents, _ := h.store.ListTicketWebhookEvents(r.Context(), ticketID)

	items := make([]IncidentTimelineItem, 0, len(activities)+len(comments)+len(webhookEvents))
	for _, activity := range activities {
		title := fmt.Sprintf("%s by %s", activity.Action, activity.ActorName)
		body := ""
		if activity.OldValue != nil || activity.NewValue != nil {
			body = strings.TrimSpace(fmt.Sprintf("%s -> %s", derefString(activity.OldValue), derefString(activity.NewValue)))
		}
		items = append(items, IncidentTimelineItem{
			Id:        activity.ID.String(),
			TicketId:  toOpenapiUUID(ticketID),
			Type:      IncidentTimelineItemTypeActivity,
			Title:     title,
			Body:      nullableString(body),
			CreatedAt: activity.CreatedAt,
		})
	}
	for _, comment := range comments {
		items = append(items, IncidentTimelineItem{
			Id:        comment.ID.String(),
			TicketId:  toOpenapiUUID(ticketID),
			Type:      IncidentTimelineItemTypeComment,
			Title:     fmt.Sprintf("comment by %s", comment.AuthorName),
			Body:      &comment.Message,
			CreatedAt: comment.CreatedAt,
		})
	}
	for _, evt := range webhookEvents {
		body := fmt.Sprintf("event=%s delivered=%t", evt.Event, evt.Delivered)
		items = append(items, IncidentTimelineItem{
			Id:        evt.ID.String(),
			TicketId:  toOpenapiUUID(ticketID),
			Type:      IncidentTimelineItemTypeWebhook,
			Title:     "webhook dispatch",
			Body:      &body,
			CreatedAt: evt.CreatedAt,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.Before(items[j].CreatedAt)
	})

	writeJSON(w, http.StatusOK, IncidentTimelineResponse{Items: items})
}

func (h *API) GetTicketIncidentPostmortem(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectAccess(w, r, ticket.ProjectID) {
		return
	}

	activities, _ := h.store.ListActivities(r.Context(), ticketID)
	comments, _ := h.store.ListComments(r.Context(), ticketID)
	webhookEvents, _ := h.store.ListTicketWebhookEvents(r.Context(), ticketID)

	type line struct {
		at   time.Time
		text string
	}
	lines := make([]line, 0, len(activities)+len(comments)+len(webhookEvents))
	for _, a := range activities {
		lines = append(lines, line{
			at:   a.CreatedAt,
			text: fmt.Sprintf("- %s: %s changed %s from `%s` to `%s`", a.CreatedAt.Format(time.RFC3339), a.ActorName, derefString(a.Field), derefString(a.OldValue), derefString(a.NewValue)),
		})
	}
	for _, c := range comments {
		lines = append(lines, line{
			at:   c.CreatedAt,
			text: fmt.Sprintf("- %s: comment by %s\n\n  %s", c.CreatedAt.Format(time.RFC3339), c.AuthorName, strings.ReplaceAll(c.Message, "\n", "\n  ")),
		})
	}
	for _, e := range webhookEvents {
		lines = append(lines, line{
			at:   e.CreatedAt,
			text: fmt.Sprintf("- %s: webhook `%s` (delivered=%t)", e.CreatedAt.Format(time.RFC3339), e.Event, e.Delivered),
		})
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].at.Before(lines[j].at) })

	var b strings.Builder
	b.WriteString("# Postmortem Draft: " + ticket.Key + " " + ticket.Title + "\n\n")
	b.WriteString("## Incident Summary\n\n")
	b.WriteString("- Severity: " + derefString(ticket.IncidentSeverity) + "\n")
	b.WriteString("- Commander: " + derefString(ticket.IncidentCommanderName) + "\n")
	b.WriteString("- Impact: " + derefString(ticket.IncidentImpact) + "\n\n")
	b.WriteString("## Timeline\n\n")
	for _, l := range lines {
		b.WriteString(l.text + "\n")
	}
	b.WriteString("\n## Root Cause\n\n")
	b.WriteString("- TBD\n\n")
	b.WriteString("## Action Items\n\n")
	b.WriteString("- TBD\n")

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.String()))
}

func (h *API) recordTicketActivities(ctx context.Context, before, after store.Ticket, actorID uuid.UUID, actorName string) {
	type fieldChange struct {
		action   string
		field    string
		oldValue string
		newValue string
	}

	var changes []fieldChange

	if before.StateID != after.StateID {
		changes = append(changes, fieldChange{
			action:   "state_changed",
			field:    "state",
			oldValue: before.StateName,
			newValue: after.StateName,
		})
	}
	if before.Priority != after.Priority {
		changes = append(changes, fieldChange{
			action:   "priority_changed",
			field:    "priority",
			oldValue: before.Priority,
			newValue: after.Priority,
		})
	}
	assigneeChanged := (before.AssigneeID == nil) != (after.AssigneeID == nil) ||
		(before.AssigneeID != nil && after.AssigneeID != nil && *before.AssigneeID != *after.AssigneeID)
	if assigneeChanged {
		oldName := ""
		newName := ""
		if before.AssigneeName != nil {
			oldName = *before.AssigneeName
		}
		if after.AssigneeName != nil {
			newName = *after.AssigneeName
		}
		changes = append(changes, fieldChange{
			action:   "assignee_changed",
			field:    "assignee",
			oldValue: oldName,
			newValue: newName,
		})
	}
	if before.Type != after.Type {
		changes = append(changes, fieldChange{
			action:   "type_changed",
			field:    "type",
			oldValue: before.Type,
			newValue: after.Type,
		})
	}
	if before.Title != after.Title {
		changes = append(changes, fieldChange{
			action:   "title_changed",
			field:    "title",
			oldValue: before.Title,
			newValue: after.Title,
		})
	}
	if derefString(before.IncidentSeverity) != derefString(after.IncidentSeverity) {
		changes = append(changes, fieldChange{
			action:   "incident_severity_changed",
			field:    "incidentSeverity",
			oldValue: derefString(before.IncidentSeverity),
			newValue: derefString(after.IncidentSeverity),
		})
	}

	for _, c := range changes {
		field := c.field
		oldVal := c.oldValue
		newVal := c.newValue
		_ = h.store.CreateActivity(ctx, after.ID, store.ActivityCreateInput{
			ActorID:   actorID,
			ActorName: actorName,
			Action:    c.action,
			Field:     &field,
			OldValue:  &oldVal,
			NewValue:  &newVal,
		})
	}
}

func (h *API) dispatchTicketWebhook(ctx context.Context, projectID, ticketID uuid.UUID, event string, payload map[string]any) {
	if h.webhooks != nil {
		h.webhooks.Dispatch(ctx, projectID, event, payload)
		_ = h.store.CreateTicketWebhookEvent(ctx, store.TicketWebhookEventCreateInput{
			TicketID:  ticketID,
			Event:     event,
			Payload:   payload,
			Delivered: true,
		})
	}
}

func nullableString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func mapStringPtr[T ~string](value *T, mapper func(T) string) *string {
	if value == nil {
		return nil
	}
	mapped := mapper(*value)
	return &mapped
}

func parseOpenapiUUIDPtr(id *openapi_types.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	parsed := uuid.UUID(*id)
	return &parsed
}
