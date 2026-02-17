package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/auth"
	"ticketing-system/backend/internal/store"
	"ticketing-system/backend/internal/webhook"
)

type fakeStore struct {
	pingErr error
	listErr error
	states  []store.WorkflowState

	listTickets      []store.Ticket
	listTicketsErr   error
	listTicketsTotal int

	boardTickets    []store.Ticket
	boardTicketsErr error

	getTicket    store.Ticket
	getTicketErr error

	createTicket    store.Ticket
	createTicketErr error
	createInput     store.TicketCreateInput

	updateTicket    store.Ticket
	updateTicketErr error
	updateInput     store.TicketUpdateInput

	deleteTicketErr error

	webhooks         []store.Webhook
	webhookErr       error
	getWebhook       store.Webhook
	getWebhookErr    error
	createWebhook    store.Webhook
	createWebhookErr error
	updateWebhook    store.Webhook
	updateWebhookErr error
	deleteWebhookErr error

	replaceErr    error
	replaceResult []store.WorkflowState
	replaceInputs []store.WorkflowStateInput

	projects          []store.Project
	projectErr        error
	getProject        store.Project
	getProjectErr     error
	createProject     store.Project
	createProjErr     error
	updateProject     store.Project
	updateProjErr     error
	deleteProjErr     error
	projectIDsForUser []uuid.UUID

	projectGroups         []store.ProjectGroup
	projectGroupErr       error
	addProjectGroup       store.ProjectGroup
	addProjectGroupErr    error
	updateProjectGroup    store.ProjectGroup
	updateProjectGroupErr error
	deleteProjectGroup    error

	groups       []store.Group
	groupErr     error
	getGroup     store.Group
	getGroupErr  error
	createGroup  store.Group
	createGrpErr error
	updateGroup  store.Group
	updateGrpErr error
	deleteGrpErr error

	groupMembers      []store.GroupMember
	groupMemberErr    error
	addGroupMember    store.GroupMember
	addGroupMemberErr error
	deleteGroupMember error

	upsertUserErr error
	users         []store.UserSummary
	usersErr      error

	stories        []store.Story
	storiesErr     error
	getStory       store.Story
	getStoryErr    error
	createStory    store.Story
	createStoryErr error
	updateStory    store.Story
	updateStoryErr error
	deleteStoryErr error

	comments         []store.Comment
	commentsErr      error
	createComment    store.Comment
	createCommentErr error
	deleteCommentErr error

	webhookDeliveries    []store.WebhookDelivery
	webhookDeliveriesErr error

	projectRoleForUser    string
	projectRoleForUserErr error
}

func (f *fakeStore) Ping(ctx context.Context) error {
	return f.pingErr
}

func (f *fakeStore) ListWorkflowStates(ctx context.Context, projectID uuid.UUID) ([]store.WorkflowState, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.states, nil
}

func (f *fakeStore) ListProjects(ctx context.Context) ([]store.Project, error) {
	if f.projectErr != nil {
		return nil, f.projectErr
	}
	return f.projects, nil
}

func (f *fakeStore) ListProjectsForUser(ctx context.Context, userID uuid.UUID) ([]store.Project, error) {
	if f.projectErr != nil {
		return nil, f.projectErr
	}
	return f.projects, nil
}

func (f *fakeStore) ListProjectIDsForUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if f.projectErr != nil {
		return nil, f.projectErr
	}
	if len(f.projectIDsForUser) > 0 {
		return f.projectIDsForUser, nil
	}
	ids := make([]uuid.UUID, len(f.projects))
	for i, project := range f.projects {
		ids[i] = project.ID
	}
	return ids, nil
}

func (f *fakeStore) GetProject(ctx context.Context, id uuid.UUID) (store.Project, error) {
	if f.getProjectErr != nil {
		return store.Project{}, f.getProjectErr
	}
	return f.getProject, nil
}

func (f *fakeStore) CreateProject(ctx context.Context, input store.ProjectCreateInput) (store.Project, error) {
	if f.createProjErr != nil {
		return store.Project{}, f.createProjErr
	}
	return f.createProject, nil
}

func (f *fakeStore) UpdateProject(ctx context.Context, id uuid.UUID, input store.ProjectUpdateInput) (store.Project, error) {
	if f.updateProjErr != nil {
		return store.Project{}, f.updateProjErr
	}
	return f.updateProject, nil
}

func (f *fakeStore) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return f.deleteProjErr
}

func (f *fakeStore) ListProjectGroups(ctx context.Context, projectID uuid.UUID) ([]store.ProjectGroup, error) {
	if f.projectGroupErr != nil {
		return nil, f.projectGroupErr
	}
	return f.projectGroups, nil
}

func (f *fakeStore) AddProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID, role string) (store.ProjectGroup, error) {
	if f.addProjectGroupErr != nil {
		return store.ProjectGroup{}, f.addProjectGroupErr
	}
	return f.addProjectGroup, nil
}

func (f *fakeStore) UpdateProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID, role string) (store.ProjectGroup, error) {
	if f.updateProjectGroupErr != nil {
		return store.ProjectGroup{}, f.updateProjectGroupErr
	}
	return f.updateProjectGroup, nil
}

func (f *fakeStore) DeleteProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID) error {
	return f.deleteProjectGroup
}

func (f *fakeStore) ListGroups(ctx context.Context) ([]store.Group, error) {
	if f.groupErr != nil {
		return nil, f.groupErr
	}
	return f.groups, nil
}

func (f *fakeStore) GetGroup(ctx context.Context, id uuid.UUID) (store.Group, error) {
	if f.getGroupErr != nil {
		return store.Group{}, f.getGroupErr
	}
	return f.getGroup, nil
}

func (f *fakeStore) CreateGroup(ctx context.Context, input store.GroupCreateInput) (store.Group, error) {
	if f.createGrpErr != nil {
		return store.Group{}, f.createGrpErr
	}
	return f.createGroup, nil
}

func (f *fakeStore) UpdateGroup(ctx context.Context, id uuid.UUID, input store.GroupUpdateInput) (store.Group, error) {
	if f.updateGrpErr != nil {
		return store.Group{}, f.updateGrpErr
	}
	return f.updateGroup, nil
}

func (f *fakeStore) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	return f.deleteGrpErr
}

func (f *fakeStore) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]store.GroupMember, error) {
	if f.groupMemberErr != nil {
		return nil, f.groupMemberErr
	}
	return f.groupMembers, nil
}

func (f *fakeStore) AddGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (store.GroupMember, error) {
	if f.addGroupMemberErr != nil {
		return store.GroupMember{}, f.addGroupMemberErr
	}
	return f.addGroupMember, nil
}

func (f *fakeStore) DeleteGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	return f.deleteGroupMember
}

func (f *fakeStore) UpsertUser(ctx context.Context, input store.UserUpsertInput) error {
	return f.upsertUserErr
}

func (f *fakeStore) ListUsers(ctx context.Context, query string) ([]store.UserSummary, error) {
	if f.usersErr != nil {
		return nil, f.usersErr
	}
	return f.users, nil
}

func (f *fakeStore) ListStories(ctx context.Context, projectID uuid.UUID) ([]store.Story, error) {
	if f.storiesErr != nil {
		return nil, f.storiesErr
	}
	return f.stories, nil
}

func (f *fakeStore) GetStory(ctx context.Context, id uuid.UUID) (store.Story, error) {
	if f.getStoryErr != nil {
		return store.Story{}, f.getStoryErr
	}
	return f.getStory, nil
}

func (f *fakeStore) CreateStory(ctx context.Context, projectID uuid.UUID, input store.StoryCreateInput) (store.Story, error) {
	if f.createStoryErr != nil {
		return store.Story{}, f.createStoryErr
	}
	return f.createStory, nil
}

func (f *fakeStore) UpdateStory(ctx context.Context, id uuid.UUID, input store.StoryUpdateInput) (store.Story, error) {
	if f.updateStoryErr != nil {
		return store.Story{}, f.updateStoryErr
	}
	return f.updateStory, nil
}

func (f *fakeStore) DeleteStory(ctx context.Context, id uuid.UUID) error {
	return f.deleteStoryErr
}

func (f *fakeStore) ListComments(ctx context.Context, ticketID uuid.UUID) ([]store.Comment, error) {
	if f.commentsErr != nil {
		return nil, f.commentsErr
	}
	return f.comments, nil
}

func (f *fakeStore) CreateComment(ctx context.Context, ticketID uuid.UUID, input store.CommentCreateInput) (store.Comment, error) {
	if f.createCommentErr != nil {
		return store.Comment{}, f.createCommentErr
	}
	return f.createComment, nil
}

func (f *fakeStore) DeleteComment(ctx context.Context, id uuid.UUID) error {
	return f.deleteCommentErr
}

func (f *fakeStore) ListAttachments(ctx context.Context, ticketID uuid.UUID) ([]store.Attachment, error) {
	return nil, nil
}

func (f *fakeStore) GetAttachment(ctx context.Context, id uuid.UUID) (store.Attachment, error) {
	return store.Attachment{}, nil
}

func (f *fakeStore) CreateAttachment(ctx context.Context, ticketID uuid.UUID, input store.AttachmentCreateInput) (store.Attachment, error) {
	return store.Attachment{}, nil
}

func (f *fakeStore) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (f *fakeStore) ListWebhookDeliveries(ctx context.Context, webhookID uuid.UUID) ([]store.WebhookDelivery, error) {
	if f.webhookDeliveriesErr != nil {
		return nil, f.webhookDeliveriesErr
	}
	return f.webhookDeliveries, nil
}

func (f *fakeStore) GetProjectStats(ctx context.Context, projectID uuid.UUID) (store.ProjectStats, error) {
	return store.ProjectStats{}, nil
}

func (f *fakeStore) GetProjectRoleForUser(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	if f.projectRoleForUser != "" {
		return f.projectRoleForUser, f.projectRoleForUserErr
	}
	return "admin", f.projectRoleForUserErr
}

func (f *fakeStore) ReplaceWorkflowStates(ctx context.Context, projectID uuid.UUID, inputs []store.WorkflowStateInput) ([]store.WorkflowState, error) {
	f.replaceInputs = inputs
	if f.replaceErr != nil {
		return nil, f.replaceErr
	}
	return f.replaceResult, nil
}

func (f *fakeStore) ListTickets(ctx context.Context, filter store.TicketFilter) ([]store.Ticket, int, error) {
	if f.listTicketsErr != nil {
		return nil, 0, f.listTicketsErr
	}
	return f.listTickets, f.listTicketsTotal, nil
}

func (f *fakeStore) ListTicketsForBoard(ctx context.Context, projectID uuid.UUID) ([]store.Ticket, error) {
	if f.boardTicketsErr != nil {
		return nil, f.boardTicketsErr
	}
	return f.boardTickets, nil
}

func (f *fakeStore) GetTicket(ctx context.Context, id uuid.UUID) (store.Ticket, error) {
	if f.getTicketErr != nil {
		return store.Ticket{}, f.getTicketErr
	}
	return f.getTicket, nil
}

func (f *fakeStore) CreateTicket(ctx context.Context, projectID uuid.UUID, input store.TicketCreateInput) (store.Ticket, error) {
	f.createInput = input
	if f.createTicketErr != nil {
		return store.Ticket{}, f.createTicketErr
	}
	return f.createTicket, nil
}

func (f *fakeStore) UpdateTicket(ctx context.Context, id uuid.UUID, input store.TicketUpdateInput) (store.Ticket, error) {
	f.updateInput = input
	if f.updateTicketErr != nil {
		return store.Ticket{}, f.updateTicketErr
	}
	return f.updateTicket, nil
}

func (f *fakeStore) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	return f.deleteTicketErr
}

func (f *fakeStore) ListWebhooks(ctx context.Context, projectID uuid.UUID) ([]store.Webhook, error) {
	if f.webhookErr != nil {
		return nil, f.webhookErr
	}
	return f.webhooks, nil
}

func (f *fakeStore) GetWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) (store.Webhook, error) {
	if f.getWebhookErr != nil {
		return store.Webhook{}, f.getWebhookErr
	}
	return f.getWebhook, nil
}

func (f *fakeStore) CreateWebhook(ctx context.Context, projectID uuid.UUID, input store.WebhookCreateInput) (store.Webhook, error) {
	if f.createWebhookErr != nil {
		return store.Webhook{}, f.createWebhookErr
	}
	return f.createWebhook, nil
}

func (f *fakeStore) UpdateWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID, input store.WebhookUpdateInput) (store.Webhook, error) {
	if f.updateWebhookErr != nil {
		return store.Webhook{}, f.updateWebhookErr
	}
	return f.updateWebhook, nil
}

func (f *fakeStore) DeleteWebhook(ctx context.Context, projectID uuid.UUID, id uuid.UUID) error {
	return f.deleteWebhookErr
}

type fakeAuth struct {
	loginUser  auth.User
	loginToken auth.TokenSet
	loginErr   error
	verifyUser auth.User
	verifyErr  error
	listUsers  []auth.User
	listErr    error
}

func (f *fakeAuth) Login(ctx context.Context, username, password string) (auth.User, auth.TokenSet, error) {
	if f.loginErr != nil {
		return auth.User{}, auth.TokenSet{}, f.loginErr
	}
	return f.loginUser, f.loginToken, nil
}

func (f *fakeAuth) Verify(ctx context.Context, token string) (auth.User, error) {
	if f.verifyErr != nil {
		return auth.User{}, f.verifyErr
	}
	return f.verifyUser, nil
}

func (f *fakeAuth) ListUsers(ctx context.Context) ([]auth.User, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listUsers, nil
}

type fakeWebhookDispatcher struct {
	events []string
	data   []any
}

func openapiUUID(value string) openapi_types.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil
	}
	return parsed
}

func (f *fakeWebhookDispatcher) Dispatch(ctx context.Context, projectID uuid.UUID, event string, data any) {
	f.events = append(f.events, event)
	f.data = append(f.data, data)
}

func (f *fakeWebhookDispatcher) Test(ctx context.Context, hook store.Webhook, event string, data any) (webhook.Result, error) {
	f.events = append(f.events, event)
	f.data = append(f.data, data)
	return webhook.Result{Delivered: true, StatusCode: 200}, nil
}

func newHandlerWith(store Store) *API {
	return NewHandler(store, &fakeAuth{}, &fakeWebhookDispatcher{}, HandlerOptions{})
}

func newTestRequest(method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	user := auth.User{
		ID:    uuid.NewString(),
		Email: "test@example.com",
		Name:  "Test User",
		Roles: []string{"admin"},
	}
	return req.WithContext(context.WithValue(req.Context(), authUserKey, user))
}

// newTestRequestAsUser creates a request with a non-admin user (subject to role checks).
func newTestRequestAsUser(method, url string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, url, body)
	user := auth.User{
		ID:    "22222222-2222-2222-2222-222222222222",
		Email: "regular@example.com",
		Name:  "Regular User",
		Roles: []string{"default-roles-ticketing"},
	}
	return req.WithContext(context.WithValue(req.Context(), authUserKey, user))
}

func TestHealth(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{})
		req := newTestRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.HealthCheck(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var payload HealthResponse
		if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if payload.Status != "ok" {
			t.Fatalf("expected status ok, got %q", payload.Status)
		}
	})

	t.Run("db down", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{pingErr: errors.New("db down")})
		req := newTestRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		h.HealthCheck(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", rec.Code)
		}
	})
}

func TestGetWorkflow(t *testing.T) {
	id := uuid.New()
	projectID := openapiUUID("11111111-1111-1111-1111-111111111111")
	states := []store.WorkflowState{
		{
			ID:        id,
			Name:      "Backlog",
			Order:     1,
			IsDefault: true,
			IsClosed:  false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	h := newHandlerWith(&fakeStore{states: states})
	req := newTestRequest(http.MethodGet, "/workflow", nil)
	rec := httptest.NewRecorder()

	h.GetWorkflow(rec, req, projectID)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		States []workflowState `json:"states"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(resp.States))
	}
	if resp.States[0].Id != openapiUUID(id.String()) {
		t.Fatalf("expected id %s, got %s", id.String(), resp.States[0].Id)
	}
}

func TestUpdateWorkflow(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{})
		req := newTestRequest(http.MethodPut, "/workflow", strings.NewReader("{"))
		rec := httptest.NewRecorder()

		h.UpdateWorkflow(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("empty states", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{})
		req := newTestRequest(http.MethodPut, "/workflow", strings.NewReader(`{"states":[]}`))
		rec := httptest.NewRecorder()

		h.UpdateWorkflow(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		storeState := store.WorkflowState{
			ID:        id,
			Name:      "Done",
			Order:     3,
			IsDefault: false,
			IsClosed:  true,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		fs := &fakeStore{replaceResult: []store.WorkflowState{storeState}}
		h := newHandlerWith(fs)
		body := `{"states":[{"id":"` + id.String() + `","name":"Done","order":3,"isDefault":false,"isClosed":true}]}`
		req := newTestRequest(http.MethodPut, "/workflow", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.UpdateWorkflow(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if len(fs.replaceInputs) != 1 {
			t.Fatalf("expected 1 input, got %d", len(fs.replaceInputs))
		}
		if fs.replaceInputs[0].Name != "Done" {
			t.Fatalf("expected name Done, got %q", fs.replaceInputs[0].Name)
		}

		var resp struct {
			States []workflowState `json:"states"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(resp.States) != 1 || resp.States[0].Name != "Done" {
			t.Fatalf("unexpected response states: %+v", resp.States)
		}
	})
}

func TestGetBoard(t *testing.T) {
	id := uuid.New()
	projectID := openapiUUID("11111111-1111-1111-1111-111111111111")
	states := []store.WorkflowState{
		{
			ID:        id,
			Name:      "Backlog",
			Order:     1,
			IsDefault: true,
			IsClosed:  false,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	ticketID := uuid.New()
	boardTickets := []store.Ticket{
		{
			ID:          ticketID,
			Key:         "TIC-1001",
			Number:      1001,
			Title:       "Test ticket",
			Description: "Desc",
			StateID:     id,
			StateName:   "Backlog",
			StateOrder:  1,
			Priority:    "medium",
			Position:    1,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	h := newHandlerWith(&fakeStore{states: states, boardTickets: boardTickets})
	req := newTestRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()

	h.GetBoard(rec, req, projectID)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		States  []workflowState  `json:"states"`
		Tickets []ticketResponse `json:"tickets"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(resp.States))
	}
	if len(resp.Tickets) != 1 {
		t.Fatalf("expected 1 ticket, got %d", len(resp.Tickets))
	}
	if resp.Tickets[0].Key != "TIC-1001" {
		t.Fatalf("expected ticket key, got %q", resp.Tickets[0].Key)
	}
}

func TestCreateTicket(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{})
		req := newTestRequest(http.MethodPost, "/tickets", strings.NewReader("{"))
		rec := httptest.NewRecorder()

		h.CreateTicket(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("missing title", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{})
		req := newTestRequest(http.MethodPost, "/tickets", strings.NewReader(`{"description":"x"}`))
		rec := httptest.NewRecorder()

		h.CreateTicket(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		stateID := uuid.New()
		fs := &fakeStore{createTicket: store.Ticket{ID: id, Key: "TIC-2000", Number: 2000, Title: "Hello", Description: "", StateID: stateID, StateName: "Backlog", StateOrder: 1, Priority: "medium", Position: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}}
		h := newHandlerWith(fs)
		req := newTestRequest(http.MethodPost, "/tickets", strings.NewReader(`{"title":"Hello"}`))
		rec := httptest.NewRecorder()

		h.CreateTicket(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rec.Code)
		}
		if fs.createInput.Title != "Hello" {
			t.Fatalf("expected title passed, got %q", fs.createInput.Title)
		}

		var resp ticketResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.Key != "TIC-2000" {
			t.Fatalf("expected key, got %q", resp.Key)
		}
	})
}

func TestUpdateTicketNotFound(t *testing.T) {
	id := uuid.New()
	fs := &fakeStore{updateTicketErr: pgx.ErrNoRows}
	h := newHandlerWith(fs)
	req := newTestRequest(http.MethodPatch, "/tickets/"+id.String(), strings.NewReader(`{"title":"New"}`))
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", id.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	rec := httptest.NewRecorder()

	h.UpdateTicket(rec, req, openapiUUID(id.String()))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestCreateWebhook(t *testing.T) {
	t.Run("invalid json", func(t *testing.T) {
		h := newHandlerWith(&fakeStore{})
		req := newTestRequest(http.MethodPost, "/webhooks", strings.NewReader("{"))
		rec := httptest.NewRecorder()

		h.CreateWebhook(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		id := uuid.New()
		now := time.Now().UTC()
		fs := &fakeStore{createWebhook: store.Webhook{ID: id, URL: "https://example.com/hook", Events: []string{"ticket.created"}, Enabled: true, CreatedAt: now, UpdatedAt: now}}
		h := newHandlerWith(fs)
		req := newTestRequest(http.MethodPost, "/webhooks", strings.NewReader(`{"url":"https://example.com/hook","events":["ticket.created"]}`))
		rec := httptest.NewRecorder()

		h.CreateWebhook(rec, req, openapiUUID("11111111-1111-1111-1111-111111111111"))

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rec.Code)
		}
		var resp webhookResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.Id != openapiUUID(id.String()) {
			t.Fatalf("expected id %s, got %s", id.String(), resp.Id)
		}
	})
}

func TestRBAC(t *testing.T) {
	projectID := openapiUUID("11111111-1111-1111-1111-111111111111")

	t.Run("viewer cannot create ticket", func(t *testing.T) {
		fs := &fakeStore{projectRoleForUser: "viewer"}
		h := newHandlerWith(fs)
		req := newTestRequestAsUser(http.MethodPost, "/tickets", strings.NewReader(`{"title":"Hello"}`))
		rec := httptest.NewRecorder()

		h.CreateTicket(rec, req, projectID)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("contributor can create ticket", func(t *testing.T) {
		id := uuid.New()
		stateID := uuid.New()
		fs := &fakeStore{
			projectRoleForUser: "contributor",
			createTicket: store.Ticket{
				ID: id, Key: "TIC-3000", Number: 3000, Title: "Hello",
				StateID: stateID, StateName: "Backlog", StateOrder: 1,
				Priority: "medium", Position: 1,
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			},
		}
		h := newHandlerWith(fs)
		req := newTestRequestAsUser(http.MethodPost, "/tickets", strings.NewReader(`{"title":"Hello"}`))
		rec := httptest.NewRecorder()

		h.CreateTicket(rec, req, projectID)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", rec.Code)
		}
	})

	t.Run("viewer can list tickets", func(t *testing.T) {
		fs := &fakeStore{
			projectRoleForUser: "viewer",
			projectIDsForUser:  []uuid.UUID{uuid.UUID(projectID)},
		}
		h := newHandlerWith(fs)
		req := newTestRequestAsUser(http.MethodGet, "/tickets", nil)
		rec := httptest.NewRecorder()

		h.ListTickets(rec, req, projectID, ListTicketsParams{})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("contributor cannot update workflow", func(t *testing.T) {
		fs := &fakeStore{projectRoleForUser: "contributor"}
		h := newHandlerWith(fs)
		body := `{"states":[{"name":"Done","order":1,"isDefault":true,"isClosed":true}]}`
		req := newTestRequestAsUser(http.MethodPut, "/workflow", strings.NewReader(body))
		rec := httptest.NewRecorder()

		h.UpdateWorkflow(rec, req, projectID)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("viewer cannot delete ticket", func(t *testing.T) {
		ticketID := uuid.New()
		fs := &fakeStore{
			projectRoleForUser: "viewer",
			getTicket: store.Ticket{
				ID: ticketID, ProjectID: uuid.UUID(projectID),
				Key: "TIC-1", Title: "Test",
				CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
			},
		}
		h := newHandlerWith(fs)
		req := newTestRequestAsUser(http.MethodDelete, "/tickets/"+ticketID.String(), nil)
		rec := httptest.NewRecorder()

		h.DeleteTicket(rec, req, openapiUUID(ticketID.String()))

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("get my project role", func(t *testing.T) {
		fs := &fakeStore{projectRoleForUser: "contributor"}
		h := newHandlerWith(fs)
		req := newTestRequestAsUser(http.MethodGet, "/my-role", nil)
		rec := httptest.NewRecorder()

		h.GetMyProjectRole(rec, req, projectID)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		var resp map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp["role"] != "contributor" {
			t.Fatalf("expected role contributor, got %q", resp["role"])
		}
	})
}
