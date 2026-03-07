# Current Features

Snapshot date: March 7, 2026

## Core Platform
- OpenAPI-defined backend API (`ticketing-system/openapi.yaml`) with generated backend/frontend types.
- Go backend + PostgreSQL persistence with automatic migrations on startup.
- Vue 3 + TypeScript frontend using Pinia stores and route-based project views.
- Full Docker Compose orchestration: backend, frontend, Postgres, Keycloak, n8n, codex-agent.

## Authentication
- Keycloak-backed OAuth2/OIDC integration.
- Session-based authentication with token cookies and configurable TTL.
- Login endpoint (`/auth/login`), logout endpoint (`/auth/logout`), current user endpoint (`/auth/me`).
- Context-based user injection via auth middleware.

## Projects and Access Control
- Project CRUD: list, create, get, update, delete.
- Group CRUD: list, create, get, update, delete.
- Group membership management: list members, add member, remove member.
- Project-group role mapping: list project groups, add group, update role, remove group.
- Role hierarchy enforced in backend: `admin`, `contributor`, `viewer`.
- Per-operation role enforcement via `requireProjectRole()` helper: viewers (read-only), contributors (CRUD tickets/stories/comments/attachments), admins (manage workflow/webhooks/project groups).
- `GET /projects/{projectId}/my-role` endpoint returns the current user's role for a project.
- Frontend role-aware UI: read-only ticket modal for viewers, hidden create/delete controls, Settings tab restricted to admins.
- Admin operations gated by `requireAdmin()` middleware.
- Project access gated by `requireProjectAccess()` middleware.
- User directory search endpoint (`/users?q=`) with fuzzy matching.

## Ticketing and Board
- Kanban board API and UI with drag-and-drop ticket reordering.
- Ticket CRUD: list, create, get, update, delete.
- Ticket fields: title, description, priority (urgent/high/medium/low), type (feature/bug), state, assignee, story linkage.
- Ticket key/number model in backend schema.
- Story support: list, create, get, update, delete. Board groups tickets under stories.
- Ticket comments: list, create, delete. Markdown rendering with toolbar-equipped editor.
- Ticket file attachments: upload, list, download, delete. MinIO S3-compatible object storage with swappable ObjectStore interface (in-memory for E2E tests). 10MB file size limit.
- Board search and filtering.
- Bulk ticket operations:
  - Multi-select mode on board cards with selected-count badge.
  - Bulk action toolbar for move state, assign user, set priority, and delete.
  - Optimistic UI updates with partial-failure rollback and per-ticket error messaging.
  - API endpoint: `POST /projects/{projectId}/tickets/bulk`.
- Saved board filter presets:
  - Project-scoped personal presets with persisted filter fields (`assignee`, `state`, `priority`, `type`, `q`, `blocked`).
  - Preset CRUD API endpoints:
    - `GET /projects/{projectId}/board-filters`
    - `POST /projects/{projectId}/board-filters`
    - `PATCH /projects/{projectId}/board-filters/{presetId}`
    - `DELETE /projects/{projectId}/board-filters/{presetId}`
    - `GET /projects/{projectId}/board-filters/shared/{token}`
  - Share-token links that open board with preset applied (`?share=<token>`).
  - Last active preset persisted across reloads.
- Ticket dependencies and blocked-work visibility:
  - Dependency relations: `blocks`, `blocked_by`, `related`.
  - Dependency APIs:
    - `GET /projects/{projectId}/dependency-graph` (supports `rootTicketId` + up to 2-hop `depth`).
    - `GET /tickets/{id}/dependencies`
    - `POST /tickets/{id}/dependencies`
    - `DELETE /tickets/{id}/dependencies/{dependencyId}`
  - Cycle prevention on dependency creation with explicit API error (`dependency_cycle`, 409).
  - Ticket model includes `blockedByCount` and `isBlocked`.
  - Board blocked badge on cards and blocked-only filter toggle.
  - Ticket modal dependency management controls and graph summary panel.

## Ticket Activity Timeline
- Immutable `ticket_activities` table recording every field change on ticket update.
- Tracked actions: `state_changed`, `priority_changed`, `assignee_changed`, `type_changed`, `title_changed`, `incident_severity_changed`.
- API endpoint: `GET /tickets/{id}/activities` returning chronological list.
- OpenAPI schema: `TicketActivity`, `TicketActivityListResponse` with generated Go and TypeScript types.
- Activity section in ticket detail modal: human-readable labels per action, shown above comments, reloads automatically after ticket save without closing the modal.

## Incident Bridge and Postmortem
- Ticket incident mode fields: `incidentEnabled`, `incidentSeverity` (`sev1..sev4`), `incidentImpact`, `incidentCommanderId`.
- Incident timeline aggregation endpoint: `GET /tickets/{id}/incident-timeline`.
- Markdown postmortem draft export endpoint: `GET /tickets/{id}/incident-postmortem`.
- Timeline aggregation sources: ticket activities, ticket comments, and ticket webhook trigger events.
- Ticket updates now audit `incident_severity_changed` in the activity feed.
- Ticket modal includes incident controls, incident timeline section, and postmortem export action.

## AI Triage Copilot
- Project-scoped AI triage feature toggle in Settings (enable/disable per project).
- AI triage APIs:
  - `GET /projects/{projectId}/ai-triage/settings`
  - `PATCH /projects/{projectId}/ai-triage/settings`
  - `POST /projects/{projectId}/ai-triage/suggestions`
  - `POST /projects/{projectId}/ai-triage/suggestions/{suggestionId}/decision`
- Suggestion response includes:
  - suggested summary, priority, state, optional assignee
  - confidence scores per field
  - prompt version + model metadata
- New ticket modal includes AI suggestion panel with field-by-field apply controls.
- Suggestion decisions are logged with accepted/rejected fields for auditability.

## Notifications and Inbox
- Project-scoped in-app notifications with persisted `notifications` and `notification_preferences` tables.
- Mention parsing on comments and ticket description updates using `@username` tokens.
- Assignment-change notifications on ticket assignee updates.
- Notification APIs:
  - `GET /projects/{projectId}/notifications`
  - `GET /projects/{projectId}/notifications/unread-count`
  - `POST /projects/{projectId}/notifications/{notificationId}/read`
  - `POST /projects/{projectId}/notifications/read-all`
  - `GET /projects/{projectId}/notification-preferences`
  - `PATCH /projects/{projectId}/notification-preferences`
- Header inbox UI:
  - Unread badge, notification list, and mark-all-read behavior.
  - Preference toggles for mention and assignment notifications.
- Live updates endpoint: `GET /projects/{projectId}/events/ws` (WebSocket).
- Live event types: `heartbeat`, `notifications.unread_count`, `notifications.changed`, `board.refresh`, `activity.changed`.
- WebSocket-first updates with automatic fallback to unread polling every 5 seconds while authenticated.
- Inbox optimization: notification list reload on `notifications.changed` only when inbox panel is open.

## Workflow Management
- Workflow state retrieval and update per project via API.
- Default workflow state initialization when needed.
- Board columns driven by workflow states.

## Webhooks
- Project-scoped webhook CRUD: list, create, get, update, delete.
- Async webhook dispatcher with goroutine-based delivery.
- Versioned outbound payload envelope (`v1`) with:
  - `version`
  - `event`
  - `eventTimestamp`
  - `idempotencyKey`
  - `data` (event-specific payload)
- HMAC-SHA256 request signing (`X-Ticketing-Signature` header) when secret is configured.
- Delivery metadata headers: `X-Ticketing-Webhook-Version` and `X-Ticketing-Idempotency-Key`.
- Supported events: `ticket.created`, `ticket.updated`, `ticket.deleted`, `ticket.state_changed`.
- Exponential backoff retry on failed deliveries: 3 attempts (immediate, 30s, 5min).
- `webhook_deliveries` table logging every delivery attempt with status code, response body, error, duration, and timestamp.
- Delivery history API endpoint: `GET /projects/{projectId}/webhooks/{id}/deliveries` (latest 50).
- Webhook test endpoint for manual verification.
- UI for creating, editing, enabling/disabling, and testing webhooks.
- Delivery history panel in webhook settings: expandable rows with status dot, event, attempt number, status code, duration, time ago, and response/error details.

## Project Dashboard
- Project-level statistics API endpoint: `GET /projects/{projectId}/stats`.
- Aggregate ticket counts by state, priority, type, and assignee computed from existing ticket data.
- Open vs closed ticket totals derived from workflow state `isClosed` flag.
- Blocked-open ticket count (`blockedOpen`) derived from dependency graph.
- Dashboard page accessible from header navigation tab alongside Board and Settings.
- Summary cards showing total, open, and closed ticket counts.
- Summary cards include blocked-open count.
- Horizontal bar charts for each dimension (state, priority, type, assignee) with color-coded bars.
- Dependency graph panel listing graph node/edge totals and project-level dependency edges.
- Sprint forecast panel:
  - Sprint APIs: `GET/POST /projects/{projectId}/sprints`.
  - Capacity APIs: `GET/PUT /projects/{projectId}/capacity-settings`.
  - Forecast API: `GET /projects/{projectId}/sprint-forecast?sprintId=&iterations=`.
  - Dashboard shows committed tickets, projected completion (historical-throughput simulation), capacity, and over-capacity delta.
  - Configurable simulation iteration count (10-5000).
- Sprint lifecycle management (TKT-025):
  - Sprint status machine: `planned` → `active` → `completed` (one active sprint per project enforced).
  - `POST /projects/{projectId}/sprints/{sprintId}/start` — activates a planned sprint (fails if another is active).
  - `POST /projects/{projectId}/sprints/{sprintId}/complete` — completes active sprint; accepts `moveTicketIds[]` to roll incomplete tickets to next planned sprint.
  - Sprint sidebar on board: progress bar, start/complete lifecycle buttons, collapsible completed history.
  - Complete sprint dialog: lists all incomplete tickets (pre-checked), lets user pick which to roll over; shows target sprint name.
  - i18n keys (en + de). E2E lifecycle and rollover coverage.
- Loading skeleton and empty state handling.

## Admin and Operations
- Admin endpoint to sync users from Keycloak (`/admin/sync-users`).
- Admin endpoint to create users in identity provider and upsert app directory (`/admin/users`).
- Admin authorization enforced for user sync endpoint (`/admin/sync-users`).
- Settings UI includes:
  - one-click "Sync users" action in group member management to make new identity-provider users discoverable before assignment.
  - dedicated Users tab for admin user creation (username, email, optional names, password) backed by `/admin/users`.
- Health check endpoint (`/health`).
- Docker Compose for local dev: Postgres, Keycloak (with realm import), n8n, backend API, codex-agent, MinIO.
- Production Docker Compose with Traefik reverse proxy and HTTPS.
- **Audit Log (TKT-026)**: Immutable JSONB-backed audit trail for all admin/security-relevant actions:
  - Recorded actions: project created/deleted, group created/deleted, project group added/updated, webhook created/deleted, users synced, user created, ticket deleted.
  - `GET /admin/audit-log` (admin-only) with limit/offset/projectId/resourceType/actorId filters.
  - Settings → Audit Log tab: paginated table with time, actor, action, resource; load-more pagination.

## Codex Agent (MCP Server)
- TypeScript MCP server providing authenticated ticket management tools.
- Keycloak OAuth2 token management with automatic refresh.
- MCP tools: `list_projects`, `list_tickets`, `get_ticket`, `search_tickets`, `add_comment`, `update_ticket_state`, `get_project_workflow`.

## E2E Testing
- Contract-driven Go + Playwright test harness (`ticketing-system/backend/e2e/`).
- Frontend contract file (`contracts/frontend_contract.json`) with routes, selectors, and flows.
- BDD-style scenario builder for readable test definitions.
- PostgreSQL testcontainers for isolated test environments.
- Webhook event capture mechanism for integration validation.
- Multi-user E2E support: admin and viewer user seeding, `WithViewerUser()` harness option, API request helper with auth cookies.
- Test coverage: login/logout, project selection, ticket CRUD, story management, comments, file attachments (upload, delete), webhook events, drag-and-drop, form validation, unhappy paths, RBAC negative-path tests (viewer cannot create/delete tickets, cannot access settings/workflow), activity timeline (state change, priority change visible after ticket update).
- Additional bulk-operation coverage: admin bulk flow (move/assign/set priority/delete) and viewer bulk API permission-failure summaries.
- Sprint planner coverage: API create/list sprint + capacity replacement + forecast endpoint assertions, and dashboard sprint forecast panel selectors.

## Frontend UX
- Login view with session bootstrap.
- Project board page:
  - Kanban columns with story-group rows, drag-and-drop, and board search/filter controls.
  - Single primary toolbar in default mode with `Views`-first interaction and collapsible filters.
  - Contextual floating bulk-action bar that appears only in selection mode.
  - Story-group progress indicator and tighter story-column layout (~15% width target).
  - Ticket-card clarity updates: priority edge stripe, stronger selected state, drag handle, type iconography, assignee avatar, and long-ID truncation with full ID/title metadata retained.
  - Hover quick actions for common operations (move to next state, cycle priority, assign to current user).
- Reusable MarkdownEditor component with formatting toolbar (bold, italic, code, link, lists, quote, heading), keyboard shortcuts (Ctrl+B/I/E/K, Tab indent), and edit/preview toggle. Used in all description and comment fields.
- Ticket detail modal: inline editing, comments with markdown, file attachments (upload/download/delete), assignee/priority/type/story fields, activity timeline showing state/priority/assignee/type/title changes.
- New ticket and story creation modals with markdown-enabled description fields.
- Settings page (two tabs):
  - Projects: project CRUD, group management, member management, project-group role assignment.
  - Webhooks: webhook CRUD, test delivery, enable/disable toggle, delivery history panel.
- Project dashboard page: summary cards (total/open/closed), bar charts by state, priority, type, and assignee.
- Project drawer for switching between projects.
- Header live-update status indicator (`WS` or `POLL`) showing active transport mode.

## Live Collaboration (TKT-028, March 7, 2026)
- Ticket soft-locking: `ticket_locks` table (migration `024_ticket_locks.sql`) with 2-minute TTL. REST endpoints: `POST /tickets/{id}/lock` (acquire, returns 409 on conflict), `PUT /tickets/{id}/lock` (renew), `DELETE /tickets/{id}/lock` (release, 204).
- In-memory presence store (`internal/presence/store.go`) with 30-second TTL and concurrent-safe access. REST endpoint: `POST /projects/{projectId}/presence` (upsert, broadcasts `presence.update` WS event).
- New WS live-event types: `presence.update`, `ticket.lock_acquired`, `ticket.lock_released`, `ticket.comment_added` — added to `ProjectLiveEventType` enum in `openapi.yaml`.
- Frontend composables: `usePresence` (15s heartbeat), `useLock` (acquire/renew every 90s/release). Both gated by `VITE_LIVE_COLLAB_ENABLED` env var.
- `BoardPresenceBar.vue`: avatar bubbles showing active users (max 5 + overflow count).
- OpenAPI-first restoration: `openapi.yaml` (deleted in commit 0cbec5a) restored and extended with all new schemas/paths; `make generate` regenerates `generated.gen.go` and `api.schema.ts` — router.go is clean with only `HandlerWithOptions`.

## Automation Engine (TKT-027, March 5, 2026)
- Per-project automation rules with trigger events + conditions + ordered actions.
- `automation_rules` and `automation_executions` tables (migration `023_automation_rules.sql`).
- Store: `ListAutomationRules`, `ListEnabledAutomationRules`, `GetAutomationRule`, `CreateAutomationRule`, `UpdateAutomationRule`, `DeleteAutomationRule`, `BumpAutomationRuleStats`, `CreateAutomationExecution`, `ListAutomationExecutions`.
- Automation engine (`internal/automation/engine.go`): fire-and-forget via `go engine.Fire(...)` from `CreateTicket` and `UpdateTicket` handlers; dispatches `ticket.created`, `ticket.updated`, `ticket.state_changed` events.
- Supported action types: `set_state`, `set_assignee`, `set_priority`, `add_comment` (with `{{ticket.key}}`/`{{ticket.title}}` template vars), `call_webhook`.
- REST API: 6 endpoints under `/projects/{projectId}/automation/rules` (CRUD + executions list).
- Frontend: Automation tab in SettingsPage with rule list, inline rule-builder form (event select, conditions for state_changed, action builder), execution history per rule.
- Pinia store: `useAutomationStore` with CRUD actions and execution loading.
- i18n: ~18 keys in en + de.
- E2E tests: API CRUD + UI tab navigation (`automation_rules_test.go`).
