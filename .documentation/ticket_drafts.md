# Ticket Drafts (Roadmap Batch 2)

Date: February 19, 2026
Source: `.documentation/feature_roadmap.md`

## Previously Completed (Batch 1)

| Draft | Title | Status |
|-------|-------|--------|
| TKT-001 | Webhook delivery end-to-end | Done (dispatch, signing, events, retry + delivery logs all complete via TKT-007) |
| TKT-002 | Board/workflow regression | Done (12+ E2E test files, contract-driven harness) |
| TKT-003 | Local dev compose and runbook | Done (full Docker Compose with all services) |
| TKT-004 | Role-based authorization audit | Mostly done (hierarchy enforced; granular checks are TKT-008) |
| TKT-005 | Ticket activity timeline | Carried forward as TKT-009 |
| TKT-006 | Workflow editor UX | Carried forward as TKT-010 |

---

## New Tickets

### TKT-007: Add Webhook Retry Logic and Delivery History — **DONE**
- Priority: `P0`
- Status: **Completed** (February 17, 2026)
- What shipped:
  - Exponential backoff retry (3 attempts: immediate, 30s, 5min) in `deliverWithRetry()`.
  - `webhook_deliveries` migration (`011_webhook_deliveries.sql`): webhook_id, event, attempt, status_code, response_body, error, delivered, duration_ms, created_at.
  - Store layer: `CreateWebhookDelivery`, `ListWebhookDeliveries` (latest 50).
  - API endpoint: `GET /projects/{projectId}/webhooks/{id}/deliveries`.
  - OpenAPI schema: `WebhookDelivery`, `WebhookDeliveryListResponse`.
  - Settings UI: "History" button per webhook, expandable delivery list with status dot, event, attempt, status code, duration, time ago, and expandable response/error details.
  - Response body truncated to 4KB to prevent bloat.
  - All existing tests pass. fakeStore stubs updated for new interface methods.

### TKT-008: Enforce Granular Per-Operation Role Permissions — **DONE**
- Priority: `P0`
- Status: **Completed** (February 17, 2026)
- What shipped:
  - SQL template `project_role_for_user.sql` resolving highest role via project_groups + group_memberships join.
  - `GetProjectRoleForUser` store method returning "" for no membership.
  - `requireProjectRole()` handler helper with role rank system (admin=3, contributor=2, viewer=1). System admins bypass all checks.
  - 17+ write handlers patched: ticket/story/comment/attachment CRUD → contributor, workflow/webhook/project-group management → admin.
  - `GET /projects/{projectId}/my-role` API endpoint returning current user's role.
  - Frontend: `currentUserRole` state in board store, `canEditTickets`/`canManageProject` getters.
  - UI gating: read-only ticket modal for viewers, hidden New Ticket/Story buttons, hidden Settings tab for non-admins.
  - 6 handler unit tests (viewer 403 on create/delete, contributor 201, viewer 200 on read, contributor 403 on workflow, my-role endpoint).
  - 4 E2E tests with multi-user support: viewer cannot create ticket (UI), viewer cannot see Settings tab, viewer API 403 on create ticket, viewer API 403 on update workflow.
  - Multi-user E2E infrastructure: `staticAuth` with multiple entries, `WithViewerUser()` harness option, `APIRequest()` helper, viewer seed data.

### TKT-009: Add Ticket Activity Timeline (Backend + UI) — **DONE**
- Priority: `P1`
- Status: **Completed** (February 18, 2026)
- What shipped:
  - `ticket_activities` migration (`012_ticket_activities.sql`): ticket_id, actor_id, actor_name, action, field, old_value, new_value, created_at.
  - Store layer: `ListActivities`, `CreateActivity` methods in `store/activities.go`.
  - SQL templates in `07_activities.go.templ`.
  - `GET /tickets/{id}/activities` API endpoint with OpenAPI spec + generated Go and TypeScript types.
  - Activity recording in `UpdateTicket` handler for: state_changed, priority_changed, assignee_changed, type_changed, title_changed.
  - Frontend: `listTicketActivities` API wrapper, `ticketActivities` store state, `loadTicketActivities` action.
  - TicketModal.vue: Activity section above comments with human-readable descriptions per action type.
  - 2 E2E tests: state change activity visible after ticket update, priority change activity visible.
  - Contract selectors: `ticket.activity_timeline`, `ticket.activity_item`.

### TKT-010: Build Visual Workflow State Editor — **DONE**
- Priority: `P1`
- Status: **Completed** (February 2026)
- What shipped:
  - Workflow editor tab in SettingsPage with full CRUD for workflow states.
  - Controls: add state, rename state, delete state, drag-and-drop reorder via `draggable` rows.
  - State properties: name input, isDefault radio (exactly one enforced), isClosed checkbox.
  - Client-side validation: at least 1 state, non-empty names, unique names, exactly 1 default.
  - `window.confirm()` dialog when deleting an existing state that may have tickets.
  - Backend `PUT /projects/{projectId}/workflow` with transactional `ReplaceWorkflowStates`.
  - Board store integration: `loadWorkflowEditor()` and `saveWorkflowEditor()` actions.
  - 4 E2E tests: add+save, rename+toggle closed, validation error, reorder via API.
  - Contract selectors for all workflow editor elements.

### TKT-011: Add Ticket File Attachments with MinIO CDN — **DONE**
- Priority: `P1`
- Status: **Completed** (February 17, 2026)
- What shipped:
  - MinIO S3-compatible object storage with swappable `ObjectStore` interface (MinIO for prod, in-memory for E2E).
  - 4 REST endpoints: upload (multipart), list, download (streaming), delete.
  - `ticket_attachments` migration, store CRUD, handler layer.
  - Frontend: file picker, attachment list with download links/delete buttons in ticket modal.
  - Docker Compose `minio` service. 10MB configurable upload limit.
  - 2 E2E tests (upload+list, delete).
- Not yet shipped: Nginx CDN caching layer (downloads go through backend).

### TKT-012: Add Project Dashboard Overview Page — **DONE** (fully closed)
- Recent activity feed added (February 19, 2026):
  - `GET /projects/{projectId}/activities?limit=N` endpoint with project-scoped SQL JOIN on tickets.
  - `ProjectActivity` schema with `ticketKey` + `ticketTitle` for feed context.
  - Dashboard page loads activities in parallel with stats.
  - Feed renders actor avatar, human-readable label, ticket key + title, timestamp.
  - Loading skeleton and empty state.

### TKT-012: Add Project Dashboard Overview Page — **DONE** (original)
- Priority: `P1`
- Status: **Completed** (February 17, 2026)
- What shipped:
  - `GET /projects/{projectId}/stats` API endpoint returning aggregate ticket counts.
  - `ProjectStats` and `StatCount` OpenAPI schemas with generated types.
  - Store layer with 5 SQL aggregation queries: by state, priority, type, assignee, and open/closed.
  - Frontend route `/projects/:projectId/dashboard` with lazy-loaded `DashboardPage.vue`.
  - Dashboard tab in header navigation alongside Board and Settings.
  - Summary cards (total, open, closed) with large numbers and color accents.
  - Horizontal bar charts for state, priority (color-coded), type (color-coded), and assignee.
  - Loading skeleton and empty state handling.
  - All existing tests pass. fakeStore stubs updated for new interface methods.
- Not yet shipped: Recent activity feed (depends on TKT-009).

---

## Suggested Milestone Split
1. **Milestone A (Harden):** ~~TKT-007~~, ~~TKT-008~~ — **Complete**
2. **Milestone B (Core Features):** TKT-009, ~~TKT-010~~, ~~TKT-011~~, ~~TKT-012~~

---

## Roadmap Batch 3 (New Ideas)

### TKT-013: Saved Filters and Shareable Board Views
- Priority: `P1`
- Scope:
  - Persist personal filter presets (assignee, state, priority, type, search text).
  - Add quick-select dropdown in board toolbar.
  - Add share link token for a preset (read-only for viewers).
- Acceptance criteria:
  - User can save, rename, delete, and apply presets.
  - Reload preserves last active preset.
  - Shared link opens board with preset applied.

### TKT-014: Bulk Ticket Operations
- Priority: `P1`
- Scope:
  - Multi-select mode on board cards with count badge.
  - Bulk actions: move state, assign user, set priority, delete.
  - Server-side permission checks per ticket.
- Acceptance criteria:
  - Mixed-permission batches return per-ticket success/error summary.
  - Optimistic UI updates with rollback on partial failures.
  - E2E coverage for contributor/viewer/admin role behavior.

### TKT-015: Mentions and Notification Inbox
- Priority: `P1`
- Scope:
  - Parse `@username` in comments and ticket description updates.
  - Add `notifications` table and unread count endpoint.
  - Header inbox panel with mark-read/mark-all-read actions.
- Acceptance criteria:
  - Mentioned users receive in-app notifications within 5 seconds.
  - Assignment changes generate notifications.
  - Notification preferences support mention-only and assignment-only.

### TKT-016: Dependency Graph and Blocked Work
- Priority: `P2`
- Scope:
  - Add ticket dependency relations (`blocks`, `blocked_by`, `related`).
  - Graph visualization in ticket modal and project dashboard.
  - Board badge and filter for blocked tickets.
- Acceptance criteria:
  - Cyclic dependencies are prevented with clear API error.
  - Blocked tickets are highlighted on board and in stats.
  - Graph view supports at least 2-hop expansion.

### TKT-017: Rule-Based Automation Engine
- Priority: `P2`
- Scope:
  - Project-level trigger/action rules (event + condition + actions).
  - Actions: set fields, add comment, trigger webhook.
  - Execution log with run status, latency, and error output.
- Acceptance criteria:
  - Rule execution is idempotent and loop-safe.
  - Dry-run mode shows intended changes without persisting.
  - Admin-only CRUD and audit trail for rule changes.

### TKT-018: Sprint Planner and Capacity Forecast
- Priority: `P2`
- Scope:
  - Sprint entity with date range, goal, and selected tickets.
  - Capacity settings per user/team and projected load calculation.
  - Confidence forecast using historical throughput simulation.
- Acceptance criteria:
  - Planner warns on over-capacity with explicit delta.
  - Forecast can run with configurable iteration count.
  - Dashboard shows committed vs projected completion.

### TKT-019: AI Triage Copilot
- Priority: `P3`
- Scope:
  - Suggest assignee, priority, and initial state on ticket creation.
  - Auto-generate concise summary for long descriptions/comments.
  - Confidence score and "apply suggestion" UX (never auto-apply by default).
- Acceptance criteria:
  - Suggestions can be accepted/rejected field-by-field.
  - Prompt/version metadata logged for each AI suggestion.
  - Feature can be disabled per project via settings.

### TKT-020: Incident Bridge and Postmortem Assistant
- Priority: `P2`
- Scope:
  - Incident mode for tickets (severity, impact, incident commander).
  - Timeline aggregation from comments, activities, and webhook events.
  - Generate postmortem draft (impact, timeline, root cause placeholders).
- Acceptance criteria:
  - Incident timeline is exportable as Markdown.
  - Severity changes are audited and visible in activity feed.
  - Integration hooks for Slack/Pager workflows via webhooks.

### TKT-021: Portfolio Command Center
- Priority: `P2`
- Scope:
  - Cross-project roll-up dashboard for open risk, throughput, and SLA breaches.
  - Milestone tracking across projects with confidence indicators.
  - Filters by group, owner, and objective/OKR.
- Acceptance criteria:
  - Supports at least 50 projects without timing out.
  - Drill-down from portfolio KPI to project/ticket details.
  - Snapshot export endpoint for weekly leadership reporting.
