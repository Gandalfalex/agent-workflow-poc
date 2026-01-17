# Ticketing System (Jira-like) Requirements

## Product Goals
- Provide a lightweight Jira-style ticketing system with authentication, multi-project support, a Kanban board, ticket state transitions, and webhooks.
- Support project-scoped access via groups and roles.

## Functional Requirements
- User authentication
  - Login/logout with email + password via Keycloak.
  - Session-based auth (cookie) for the web UI.
- Project + group access control
  - Users are assigned to groups.
  - Groups are assigned to projects with a role (admin, contributor, viewer).
  - Role permissions cover read, create, update, delete, admin.
- Project management
  - Create, update, list, and delete projects.
  - Project keys are 4-character uppercase alphanumeric strings.
- Group management
  - Create, update, list, and delete groups.
  - Add and remove users in groups.
  - Add and update group roles within projects.
- Kanban board
  - Columns represent workflow states per project (e.g., Backlog, In Progress, Done).
  - Drag-and-drop cards to change state.
  - Ticket detail view for editing title, description, assignee, priority, and state.
- Ticket management
  - Create, edit, and delete tickets.
  - Assign tickets to users.
  - Categorize tickets by type (feature, bug).
  - Group tickets into stories with shared descriptions.
  - Capture ticket comments with author + timestamp.
  - Track status history (at least current state).
  - Ticket keys follow `PROJ-123` format (4-char project key + 3-digit sequence, per project).
- Webhooks
  - Configure outbound webhooks per project.
  - Emit events on ticket create/update/state change.
  - Allow enabling/disabling endpoints and setting secret token.
- Basic admin configuration
  - Manage workflow states per project (add/reorder/rename).
  - Manage webhook endpoints.

## Non-Functional Requirements
- Frontend: Vue 3 + TypeScript + ShadCN Vue (or equivalent ShadCN-style component library).
- Backend: Go (REST API).
- OpenAPI-first approach:
  - Define API spec in `openapi.yaml` first.
  - Generate server stubs and types for backend + frontend.
  - Use `oapi-codegen` (Go) and `openapi-typescript` (frontend).
- Persistence: PostgreSQL. Store SQL must live in Go template files under `ticketing-system/backend/internal/store/sql`.
- Local dev via Docker Compose.

## Out of Scope (Initial)
- SSO/OAuth integrations.
- Multi-tenant orgs (beyond project/group scoping).
- Advanced reporting.
- Email notifications.

## UX Notes
- Clean, minimal Kanban board with smooth drag interactions.
- Fast ticket creation with inline form.
- Simple settings screens for workflow and webhooks.

## API Surface (High-level)
- Auth: login/logout/me
- Projects: list/create/update/delete, manage project groups
- Groups: list/create/update/delete, manage group members
- Tickets: list/create/update/delete
- Users: list (for assignee selection)
- Workflow: list/update
- Webhooks: list/create/update/delete/test
