# Ticket Drafts (Roadmap Batch 2)

Date: February 20, 2026
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

### TKT-013: Saved Filters and Shareable Board Views — **DONE**
- Priority: `P1`
- Scope:
  - Persist personal filter presets (assignee, state, priority, type, search text).
  - Add quick-select dropdown in board toolbar.
  - Add share link token for a preset (read-only for viewers).
- Acceptance criteria:
  - ✅ User can save, rename, delete, and apply presets.
  - ✅ Reload preserves last active preset.
  - ✅ Shared link opens board with preset applied.
  - ✅ E2E suite remains green via `make -C ticketing-system e2e`.

### TKT-014: Bulk Ticket Operations — **DONE**
- Priority: `P1`
- Scope:
  - Multi-select mode on board cards with count badge.
  - Bulk actions: move state, assign user, set priority, delete.
  - Server-side permission checks per ticket.
- Acceptance criteria:
  - ✅ Mixed-permission batches return per-ticket success/error summary.
  - ✅ Optimistic UI updates with rollback on partial failures.
  - ✅ E2E coverage for admin/viewer role behavior + backend RBAC tests for contributor/viewer.

### TKT-015: Mentions and Notification Inbox — **DONE**
- Priority: `P1`
- Scope:
  - Parse `@username` in comments and ticket description updates.
  - Add `notifications` table and unread count endpoint.
  - Header inbox panel with mark-read/mark-all-read actions.
- Acceptance criteria:
  - ✅ Mentioned users receive in-app notifications within 5 seconds (polling + immediate server writes).
  - ✅ Assignment changes generate notifications.
  - ✅ Notification preferences support mention-only and assignment-only.

### TKT-016: Dependency Graph and Blocked Work — **DONE**
- Priority: `P2`
- Scope:
  - Add ticket dependency relations (`blocks`, `blocked_by`, `related`).
  - Graph visualization in ticket modal and project dashboard.
  - Board badge and filter for blocked tickets.
- Acceptance criteria:
  - ✅ Cyclic dependencies are prevented with clear API error (`dependency_cycle`, HTTP 409).
  - ✅ Blocked tickets are highlighted on board and in stats (blocked badge/filter + dashboard blocked-open metric).
  - ✅ Graph view supports at least 2-hop expansion (ticket modal and dashboard graph sections).

### TKT-017: WebSocket Live Updates with Polling Fallback — **DONE**
- Priority: `P1-P2`
- Scope:
  - Add authenticated project-scoped WebSocket endpoint for live UI update signals.
  - Push event types for:
    - notifications unread-count changes
    - notifications list changes (mention/assignment/new-read-state)
    - board refresh cues on ticket/story mutations
    - activity feed refresh cues on ticket/story mutations
  - Frontend WebSocket subscription manager with reconnect and backoff.
  - Feature flag to keep polling fallback (`VITE_USE_WS_LIVE_UPDATES=false` disables WS).
- Acceptance criteria:
  - ✅ When WS is connected, unread badge updates without periodic polling.
  - ✅ On disconnect/failure, app resumes polling automatically within one retry interval.
  - ✅ Ticket/story mutations emit live refresh events to other active sessions in the same project.
  - ✅ Existing notification APIs and polling behavior remain backward compatible.
  - ✅ E2E coverage includes WS live-event delivery and 426-upgrade fallback with polling endpoint verification.
  - ✅ Inbox-open optimization: `notifications.changed` refreshes list only when inbox panel is open.

### TKT-018: Sprint Planner and Capacity Forecast — **DONE**
- Priority: `P2`
- Status: **Completed** (February 19, 2026)
- What shipped:
  - Sprint planner data model (`sprints`, `sprint_tickets`, `capacity_settings`) via `016_sprint_planner.sql`.
  - OpenAPI-first endpoints:
    - `GET/POST /projects/{projectId}/sprints`
    - `GET/PUT /projects/{projectId}/capacity-settings`
    - `GET /projects/{projectId}/sprint-forecast`
  - Forecast simulation based on historical daily throughput sampling from ticket activity history.
  - Configurable simulation iteration count (`iterations`, clamped 10..5000).
  - Explicit over-capacity delta in forecast response.
  - Dashboard sprint forecast panel showing committed, projected completion, capacity, and over-capacity values.
  - E2E coverage for sprint creation, capacity replacement, forecast API assertions, and dashboard panel visibility.

### TKT-019: AI Triage Copilot — **DONE**
- Priority: `P3`
- Status: **Completed** (February 19, 2026)
- What shipped:
  - OpenAPI-first AI triage endpoints:
    - `GET/PATCH /projects/{projectId}/ai-triage/settings`
    - `POST /projects/{projectId}/ai-triage/suggestions`
    - `POST /projects/{projectId}/ai-triage/suggestions/{suggestionId}/decision`
  - Backend persistence:
    - `ai_triage_settings`
    - `ai_triage_suggestions`
    - `ai_triage_suggestion_decisions`
  - Suggestion engine (heuristic local model) returning:
    - summary, priority, state, optional assignee
    - per-field confidence
    - prompt/version metadata
  - Field-by-field decision logging for accepted vs rejected suggestion fields.
  - Project settings toggle in Settings UI to enable/disable AI triage per project.
  - New Ticket modal AI suggestion panel with per-field apply checkboxes.
  - E2E coverage for toggle + suggestion panel + API suggestion/decision flow.

### TKT-020: Incident Bridge and Postmortem Assistant — **DONE**
- Priority: `P2`
- Status: **Completed** (February 19, 2026)
- What shipped:
  - OpenAPI-first incident APIs:
    - `GET /tickets/{id}/incident-timeline`
    - `GET /tickets/{id}/incident-postmortem` (`text/markdown`)
  - Incident mode fields on tickets:
    - `incidentEnabled`, `incidentSeverity`, `incidentImpact`, `incidentCommanderId`
  - Timeline aggregation from ticket comments, ticket activities, and webhook-trigger events.
  - Postmortem markdown draft generation with summary, timeline, root-cause placeholder, and action-item placeholder.
  - Severity change auditing added to ticket activity feed (`incident_severity_changed`).
  - Ticket modal incident controls + incident timeline + postmortem export button.
  - E2E coverage for severity change audit visibility + incident timeline API + markdown export API.

### TKT-021: Portfolio Command Center
- Priority: `P2`
- Status: **Done** (March 4, 2026)
- Context: Core portfolio infrastructure (GET /portfolio/stats, PortfolioDashboard.vue, portfolio store, E2E test skeleton) was already shipped. This ticket closes the remaining gaps: throughput metrics, group/owner filters, and performance indexes.
- Scope:
  1. **OpenAPI**: Add `weeklyThroughput` (int) and `avgCycleTimeHours` (float) to `ProjectPortfolioEntry`; add `ownerId` and `groupId` query params to `GET /portfolio/stats`.
  2. **Migration `021_portfolio_indexes.sql`**: Add indexes `idx_sprints_project_status`, `idx_ticket_deps_blocked_ticket`, `idx_tickets_project_state` to keep portfolio query fast at 50+ projects.
  3. **SQL template** (`13_portfolio.go.templ`): Add `throughput_stats` CTE computing weekly closed count and avg cycle time (hours) per project; add filter-variant SQL templates for `groupId` and `ownerId` filters.
  4. **Store** (`store/portfolio.go`): Add `PortfolioFilter{OwnerID, GroupID}` struct; extend `ProjectPortfolioEntry` with two new fields; update `GetPortfolioStats` signature and scan; update Store interface in `handlers.go` and fakeStore stub in `handlers_test.go`.
  5. **Handler**: Wire `ownerId`/`groupId` params from `GetPortfolioStatsParams` into `PortfolioFilter`; pass to store.
  6. **Frontend** (`PortfolioDashboard.vue`): Add group filter dropdown (loads groups via existing `ListGroups` API) and owner filter; re-call `portfolioStore.load(filter)` on change. Show `weeklyThroughput` in `ProjectHealthCard.vue`.
  7. **API** (`api.ts`): Update `getPortfolioStats` to accept optional `{ownerId?, groupId?}` params and append as query string.
  8. **i18n**: Add 5 keys each for en and de (`portfolio.filter.allGroups`, `portfolio.filter.ownerAll`, `portfolio.closedThisWeek`, `portfolio.avgCycle`, `portfolio.avgCycleHours`).
  9. **E2E contract**: Add `portfolio.filter_group_select` and `portfolio.throughput_stat` selectors; run `make e2e-contract`.
  10. **E2E test** (`portfolio_test.go`): Fix CSV content-type prefix check; add group-filter subtest; add throughput field presence assertion.
- Key files:
  - `ticketing-system/openapi.yaml`
  - `ticketing-system/backend/migrations/021_portfolio_indexes.sql`
  - `ticketing-system/backend/internal/store/sql/13_portfolio.go.templ`
  - `ticketing-system/backend/internal/store/portfolio.go`
  - `ticketing-system/backend/internal/httpapi/handlers.go`, `handlers_test.go`, `map.go`
  - `ticketing-system/frontend/src/views/PortfolioDashboard.vue`
  - `ticketing-system/frontend/src/components/app/portfolio/ProjectHealthCard.vue`
  - `ticketing-system/frontend/src/lib/api.ts`, `i18n.ts`
  - `ticketing-system/backend/e2e/contracts/frontend_contract.source.json`
  - `ticketing-system/backend/e2e/portfolio_test.go`
- Acceptance criteria:
  - Portfolio query stays under 200ms for 50 projects (verified by E2E timing assertion or manual test).
  - `weeklyThroughput` and `avgCycleTimeHours` populated for each project entry.
  - Group filter returns only projects in that group; owner filter returns only projects where user is a member.
  - Snapshot export (existing `?format=csv`) includes the two new fields.
  - E2E suite remains green via `make -C ticketing-system e2e`.
- Validation:
  - `make generate` — no errors
  - `go build ./...` — compiles
  - `npx tsc --noEmit` — types check
  - `make e2e-frontend-build` — frontend builds

### TKT-022: Board UI Clarity and Density Refresh — **HIGH PRIORITY**
- Priority: `P0`
- Status: **Done** (implemented February 20, 2026)
- Source: UX review (February 20, 2026)
- Scope:
  - Information density and typography:
    - Truncate long numeric IDs in card titles; keep descriptive text first and move full IDs to metadata/tooltip.
    - Keep monospace for IDs only; use sans-serif emphasis for card titles.
    - Reduce story column width target from current wide layout to approximately 15% (or support collapse).
  - Visual hierarchy and layout:
    - Replace always-visible bulk action row with a contextual floating bottom action bar shown only when tickets are selected.
    - Soften empty-state drop zones (lower contrast/opacity and reveal stronger affordance on hover).
    - Improve first column distinction by renaming to `Story Group` and styling it differently from workflow state columns.
  - Color and contrast:
    - Add stronger priority scanning cues via card edge/stripe color coding (especially medium/high/urgent).
    - Strengthen selected card state beyond checkmark (visible tint/glow + border).
    - Raise metadata/subtext contrast for readability on dark theme.
  - Interaction:
    - Add hover quick actions for common edits (assign, priority, move state).
    - Improve shortcut discoverability (`/`, `N`) with a help affordance/tooltips.
    - Add explicit drag handle (`⋮⋮`) to indicate draggable cards.
  - Data presentation:
    - Consolidate filter/preset controls into a single `Views` model to reduce toolbar rows.
    - Use clearer ticket-type iconography for fast scanning (with accessible labels/tooltips).
    - Ensure assignee avatars are consistently visible in card footer.
  - Micro-fixes:
    - Move selection checkboxes to the left edge of cards.
    - Add mini per-story progress indicator across states (not just total count).
    - Hide preset name input until save/create preset action is invoked.
- Acceptance criteria:
  - ✅ Board top chrome is reduced to a single primary toolbar in default (non-selection) mode.
  - ✅ Bulk actions only appear contextually when one or more tickets are selected.
  - ✅ Card readability improves for long identifiers without loss of traceability (truncated display + full ID/title metadata line).
  - ✅ Dark-theme readability and selection contrast improved (priority stripe, stronger selected state, clearer metadata text).
  - ✅ E2E selectors/contracts and tests updated for toolbar/preset interaction changes.
- Implementation notes:
  - Board filter panel now defaults to collapsed; toolbar remains primary entry point.
  - Preset-name input is hidden until explicit save/edit intent (`Save view` action).
  - Added card hover quick actions (move-to-next-state, cycle priority, assign-to-me).
  - Preserved contract-driven E2E compatibility by adding selector keys:
    - `board.filter_toggle_button`
    - `board.preset_open_editor_button`
  - Validation: frontend build and full E2E suite pass (`make -C ticketing-system e2e`).

### TKT-023: Webhook Payload Versioning (`v1` envelope + idempotency key) — **DONE**
- Priority: `P2`
- Status: **Completed** (February 20, 2026)
- Scope:
  - Introduce stable `v1` outbound webhook envelope with schema versioning.
  - Include explicit event timestamp and idempotency key per delivered webhook payload.
  - Document payload schema variants per event type for integrator contracts.
- What shipped:
  - Dispatcher envelope upgraded to:
    - `version`
    - `event`
    - `eventTimestamp`
    - `idempotencyKey`
    - `data`
  - Delivery headers now include:
    - `X-Ticketing-Webhook-Version`
    - `X-Ticketing-Idempotency-Key`
  - `openapi.yaml` now documents `WebhookPayloadV1` and event-specific data shapes:
    - `WebhookTicketCreatedData`
    - `WebhookTicketUpdatedData`
    - `WebhookTicketDeletedData`
    - `WebhookTicketStateChangedData`
  - Regenerated backend/frontend OpenAPI types.
  - Added/updated webhook dispatcher unit tests for envelope serialization and delivery headers.
- Validation:
  - `go test ./internal/webhook`
  - `go test -tags=e2e -count=1 -run 'TestWebhookFiresOnTicketCreation|TestWebhookFiresOnTicketStateChange' ./e2e -v`
  - `npm run -s build` (frontend)

### TKT-024: User Onboarding Sync Flow (Identity Provider -> App Users) — **DONE**
- Priority: `P1`
- Status: **Completed** (February 20, 2026)
- Scope:
  - Provide an easy, admin-accessible way to introduce newly created identity-provider users into the app's searchable user directory.
  - Ensure sync operation is properly admin-gated.
- What shipped:
  - Backend:
    - `POST /admin/sync-users` now enforces `requireAdmin()`.
    - New OpenAPI-backed endpoint `POST /admin/users` to create identity-provider users and upsert local app directory in one step.
    - Added handler tests for admin success and non-admin forbidden behavior.
  - Frontend:
    - Settings > Groups > Add members now includes a one-click `Sync users` button.
    - Sync action calls `/admin/sync-users` and shows `synced/total` feedback.
    - New Settings > Users tab with admin user creation form (username, email, optional first/last name, password).
    - User creation wired to `/admin/users` with success feedback and ready-for-group-assignment workflow.
    - Added selector: `settings.sync-users-button`.
- Validation:
  - `go test ./internal/httpapi`
  - `go test ./internal/auth`
  - `npm run -s build` (frontend)

### TKT-025: Sprint Lifecycle Management (Start, Complete, Ticket Rollover)
- Priority: `P1`
- Status: **Done** (March 4, 2026)
- Source: Sprint management UX review (February 23, 2026)
- Context: Sprints exist (TKT-018) but are static — create-only with no lifecycle. Users need to start sprints, complete them, and roll incomplete tickets to the next sprint. This is critical for iterative planning workflows.
- Scope:
  - **Sprint status lifecycle**: `planned` → `active` → `completed` (explicit transitions, one active sprint per project).
  - **DB migration** (`020_sprint_status.sql`): Add `status` column (enum: planned/active/completed, default planned) and `completed_at` (nullable timestamptz) to `sprints` table.
  - **API endpoints**:
    - `POST /projects/{pid}/sprints/{sid}/start` — activates a planned sprint (fails if another is already active).
    - `POST /projects/{pid}/sprints/{sid}/complete` — completes the active sprint, accepts optional `moveTicketIds: uuid[]` to roll selected tickets into the next planned sprint.
  - **OpenAPI schema changes**: Add `status` (enum) and `completedAt` to Sprint; new `SprintStartRequest` and `SprintCompleteRequest` schemas.
  - **Store methods**: `StartSprint`, `CompleteSprint`, `GetActiveSprint`, `GetNextPlannedSprint`. CompleteSprint sets status=completed + completed_at=now(), then moves specified tickets to next planned sprint via existing `sprint_tickets` junction.
  - **SQL templates**: `sprints_update_status`, `sprints_active_for_project`, `sprints_next_planned`. Update `sprints_list` and `sprints_get` to include status + completed_at columns.
  - **Handler layer**: StartSprint and CompleteSprint handlers (contributor+ role). Update mapSprint to include new fields. Add fakeStore stubs.
  - **Board Sprint Sidebar** (new `SprintSidebar.vue`): Collapsible side panel on the board showing:
    - Active sprint: name, date range, progress bar (done/total tickets), "Complete Sprint" button.
    - Planned sprints: list with name, dates, ticket count, "Start" button (disabled if active sprint exists).
    - Completed sprints (collapsed by default): name, dates, completed_at.
  - **Complete Sprint dialog**: Modal/inline with checkboxes listing all incomplete tickets in the active sprint (pre-checked). Shows next planned sprint name as rollover target. User unchecks tickets they don't want moved. Confirm/Cancel.
  - **Frontend wiring**: Load sprints on board mount. Computed getters for activeSprint, plannedSprints, completedSprints. Toggle sidebar via toolbar button. Wire start/complete handlers.
  - **i18n**: ~15 keys (en + de) for sprint sidebar, lifecycle actions, and complete dialog.
  - **E2E contract selectors**: `board.sprint_sidebar`, `board.sprint_sidebar_toggle`, `sprint.active_section`, `sprint.planned_section`, `sprint.completed_section`, `sprint.start_button`, `sprint.complete_button`, `sprint.complete_dialog`, `sprint.complete_ticket_checkbox`, `sprint.complete_confirm_button`, `sprint.complete_cancel_button`, `sprint.select_all_button`, `sprint.deselect_all_button`.
- Acceptance criteria:
  - Only one sprint can be active per project at a time.
  - Starting a sprint when another is active returns an error.
  - Completing a sprint shows a dialog with all incomplete tickets (not in closed workflow states), pre-checked.
  - Selected tickets are moved to the next planned sprint (by start_date ASC). If no next sprint exists, tickets stay unassigned.
  - Sprint sidebar is togglable on the board and shows accurate state for all sprint statuses.
  - Existing sprint forecast (TKT-018) continues to work with status-aware sprints.
- Key files:
  - `migrations/020_sprint_status.sql`
  - `openapi.yaml`
  - `store/sql/12_sprint_planner.go.templ`
  - `store/sprint_planner.go`
  - `handlers.go`, `types.go`, `handlers_sprint_planner.go`, `handlers_test.go`, `map.go`
  - `api.ts`, `board.ts`, `SprintSidebar.vue` (new), `BoardPage.vue`, `i18n.ts`
  - `frontend_contract.source.json`
- Validation:
  - `make generate` — no errors
  - `go build ./...` — compiles
  - `npx tsc --noEmit` — types check
  - `make e2e-frontend-build` — frontend builds
  - Manual: create sprint → start it → add tickets → complete with rollover → verify next sprint received moved tickets

### TKT-026: RBAC / Admin Audit Trail
- Priority: `P1`
- Status: **Done** (March 5, 2026)
- Source: Security and compliance review
- Context: Admin actions (project creation/deletion, group management, webhook configuration, user creation) had no audit trail. Needed an immutable, queryable log for compliance and incident review.
- What shipped:
  - **Migration** (`022_audit_log.sql`): `audit_log` table with actor_id, actor_name, action, resource_type, resource_id, resource_name, project_id (nullable FK), details (jsonb), created_at. Indexes on actor_id, project_id, resource_type, created_at DESC.
  - **SQL templates** (`16_audit_log.go.templ`): `audit_log_insert`, `audit_log_list` (nullable filter params via `$1::uuid IS NULL OR field = $1`), `audit_log_count`.
  - **Store** (`store/audit_log.go`): `AuditLogEntry`, `AuditLogCreateInput`, `AuditLogFilter` types; `CreateAuditLog`, `ListAuditLog`, `scanAuditLogEntry`.
  - **OpenAPI**: `GET /admin/audit-log` endpoint (admin-only, limit/offset/projectId/resourceType/actorId filter params). `AuditLogEntry` and `AuditLogListResponse` schemas.
  - **Handler** (`handlers_audit.go`): `recordAudit()` fire-and-forget helper. `ListAuditLog` handler.
  - **Instrumented actions**: project.created, project.deleted, group.created, group.deleted, project_group.added, project_group.updated, webhook.created, webhook.deleted, users.synced, user.created, ticket.deleted.
  - **Frontend**: `listAuditLog` in api.ts. `AuditLogEntry`/`AuditLogListResponse` type exports. `loadAuditLog` action + state in admin.ts. "Audit Log" tab in SettingsPage.vue with table (time, actor, action, resource) and load-more button.
  - **i18n**: 10 keys en + de (settings.tab.audit, settings.audit.*).
  - **E2E contract**: 4 new selectors (settings.audit_tab, settings.audit_view, settings.audit_table, settings.audit_load_more). 174 total selectors.
  - **E2E tests** (`audit_log_test.go`): API test verifies project.created entry is recorded; viewer 403 test; UI test verifies audit tab and view render.
- Key files:
  - `migrations/022_audit_log.sql`
  - `store/sql/16_audit_log.go.templ`
  - `store/audit_log.go`
  - `openapi.yaml`
  - `handlers_audit.go`, `handlers.go`, `handlers_test.go`, `types.go`, `map.go`
  - `frontend/src/lib/api.ts`, `frontend/src/stores/admin.ts`
  - `frontend/src/views/SettingsPage.vue`, `frontend/src/lib/i18n.ts`
  - `e2e/contracts/frontend_contract.source.json`, `e2e/audit_log_test.go`

### TKT-027: Rule-Based Automation Engine
- Priority: `P2`
- Status: **Done** (March 5, 2026)
- Source: Feature roadmap item 15 (March 5, 2026)
- Context: Teams do repetitive triage and handoff work manually — moving a ticket to "Done" should auto-notify QA, a bug with "urgent" priority should auto-assign a specific group, a stale ticket should auto-comment. The existing webhook system handles external integrations; this feature handles internal automated reactions to ticket events within the system itself.

#### Core Model

An **automation rule** belongs to a project and has:
- **trigger**: an event + optional condition (e.g. `ticket.state_changed` + `toState = "Done"`)
- **actions**: ordered list of internal operations to execute
- **enabled** flag
- **execution_count** and **last_executed_at** for monitoring

**Triggers** (event + optional field condition):
| Event key | Description |
|-----------|-------------|
| `ticket.created` | Any new ticket |
| `ticket.state_changed` | State transition — supports `fromState` and/or `toState` condition |
| `ticket.priority_changed` | Priority transition — supports `toPriority` condition |
| `ticket.assigned` | Assignee set or changed |
| `ticket.unassigned` | Assignee cleared |
| `ticket.type_changed` | Type changed (bug/feature/task/incident) |

Trigger conditions are optional key-value string maps (e.g. `{"toState": "<state-id>", "toPriority": "urgent"}`). Empty condition = match all.

**Actions** (executed in order, failures do not block subsequent actions):
| Action type | Parameters |
|-------------|------------|
| `set_state` | `stateId: uuid` |
| `set_assignee` | `assigneeId: uuid` |
| `set_priority` | `priority: string` |
| `add_comment` | `body: string` (supports `{{ticket.key}}`, `{{ticket.title}}`, `{{actor.name}}` template vars) |
| `call_webhook` | `webhookId: uuid` (re-uses existing webhook infrastructure) |

#### Scope

**Migration** (`023_automation_rules.sql`):
```sql
CREATE TABLE automation_rules (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL,
  enabled bool NOT NULL DEFAULT true,
  trigger_event text NOT NULL,
  trigger_conditions jsonb NOT NULL DEFAULT '{}',
  actions jsonb NOT NULL DEFAULT '[]',
  execution_count int NOT NULL DEFAULT 0,
  last_executed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX automation_rules_project_id_idx ON automation_rules(project_id) WHERE enabled = true;
```

**Execution log** (`023_automation_rules.sql` continued):
```sql
CREATE TABLE automation_executions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_id uuid NOT NULL REFERENCES automation_rules(id) ON DELETE CASCADE,
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  trigger_event text NOT NULL,
  actions_run jsonb NOT NULL DEFAULT '[]',  -- [{type, params, success, error}]
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX automation_executions_rule_id_idx ON automation_executions(rule_id);
CREATE INDEX automation_executions_ticket_id_idx ON automation_executions(ticket_id);
```

**SQL templates** (`17_automation.go.templ`):
- `automation_rules_list` — list by project_id (all, not just enabled — for settings UI)
- `automation_rules_list_enabled` — list by project_id WHERE enabled = true (for engine)
- `automation_rules_get` — by id + project_id
- `automation_rules_insert`
- `automation_rules_update`
- `automation_rules_delete`
- `automation_rules_bump_stats` — increment execution_count, set last_executed_at
- `automation_executions_insert`
- `automation_executions_list` — by rule_id, limit 50

**Store** (`store/automation.go`):
```go
type AutomationRule struct {
    ID                uuid.UUID
    ProjectID         uuid.UUID
    Name              string
    Enabled           bool
    TriggerEvent      string
    TriggerConditions map[string]string
    Actions           []AutomationAction
    ExecutionCount    int
    LastExecutedAt    *time.Time
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
type AutomationAction struct {
    Type   string         // set_state | set_assignee | set_priority | add_comment | call_webhook
    Params map[string]string
}
type AutomationExecution struct {
    ID           uuid.UUID
    RuleID       uuid.UUID
    TicketID     uuid.UUID
    TriggerEvent string
    ActionsRun   []AutomationActionResult
    CreatedAt    time.Time
}
type AutomationActionResult struct {
    Type    string
    Params  map[string]string
    Success bool
    Error   string
}
```
Methods: `ListAutomationRules`, `ListEnabledAutomationRules`, `GetAutomationRule`, `CreateAutomationRule`, `UpdateAutomationRule`, `DeleteAutomationRule`, `CreateAutomationExecution`, `ListAutomationExecutions`.

**Automation Engine** (`internal/automation/engine.go`):
- `Engine` struct holds store ref + webhook dispatcher ref.
- `func (e *Engine) Run(ctx context.Context, projectID uuid.UUID, event string, ticket store.Ticket, extra map[string]string)` — called after ticket mutations.
- Fetches enabled rules for the project, evaluates trigger conditions, executes matching rules' actions in order.
- Each action is executed via a typed `actionExecutor` (interface). Failures are logged to `automation_executions` but don't abort subsequent actions.
- Comment template expansion: replace `{{ticket.key}}`, `{{ticket.title}}`, `{{actor.name}}` with values from context.
- Non-blocking: called with `go engine.Run(...)` from handlers (fire-and-forget, same pattern as audit log).

**OpenAPI** (`openapi.yaml`):
- `AutomationRule`, `AutomationRuleCreateRequest`, `AutomationRuleUpdateRequest`, `AutomationRuleListResponse` schemas.
- `AutomationAction` schema (type + params map).
- `AutomationExecution`, `AutomationExecutionListResponse` schemas.
- Endpoints (all scoped under `/projects/{projectId}/automation/rules`):
  - `GET /projects/{projectId}/automation/rules` → listAutomationRules
  - `POST /projects/{projectId}/automation/rules` → createAutomationRule (admin/contributor)
  - `GET /projects/{projectId}/automation/rules/{ruleId}` → getAutomationRule
  - `PUT /projects/{projectId}/automation/rules/{ruleId}` → updateAutomationRule
  - `DELETE /projects/{projectId}/automation/rules/{ruleId}` → deleteAutomationRule
  - `GET /projects/{projectId}/automation/rules/{ruleId}/executions` → listAutomationRuleExecutions

**Handler wiring** (`handlers_automation.go`):
- Standard CRUD handlers.
- `Engine.Run` called (goroutine) from `CreateTicket`, `UpdateTicket`, and `BulkTicketOperation` handlers after the mutation succeeds.
- Engine is injected via `HandlerOptions`.

**Frontend**:
- `AutomationRulePage.vue` (or section in SettingsPage.vue under new "Automation" tab).
- Rule list: name, trigger event badge, enabled toggle, execution count, last run time, edit/delete buttons.
- Rule builder form:
  - Name input.
  - Trigger event select (dropdown of event keys with human labels).
  - Trigger conditions: dynamic key-value inputs shown based on selected event (e.g., "To state" select populated from workflow states for `ticket.state_changed`).
  - Actions list: add action button, each action has a type select + rendered param inputs (state select, priority select, user search, comment textarea, webhook select).
  - Drag-to-reorder actions (or up/down buttons).
  - Enabled toggle.
  - Save / Cancel.
- Execution history drawer per rule: last 50 executions with timestamp, ticket key, per-action success/error indicators.
- API functions: `listAutomationRules`, `createAutomationRule`, `updateAutomationRule`, `deleteAutomationRule`, `listAutomationExecutions`.
- Pinia store (`automation.ts`): rules list, selected rule, execution history, loading states.
- i18n: ~20 keys en + de (settings.tab.automation, automation.rule.*, automation.action.*, automation.trigger.*).

**Engine wiring in `main.go`/`server.go`**:
- Construct `automation.Engine` and inject into `HandlerOptions`.

**E2E contract selectors** (~8 new):
- `settings.automation_tab`, `settings.automation_view`, `automation.rule_list`, `automation.rule_row`, `automation.add_rule_button`, `automation.rule_name_input`, `automation.rule_trigger_select`, `automation.rule_save_button`, `automation.execution_history`.

**E2E tests** (`automation_rules_test.go`):
- API: create rule → trigger ticket state change → verify execution logged via `/executions`.
- API: disabled rule is not executed.
- UI: create rule via settings, verify it appears in list, toggle enabled.

#### Acceptance Criteria
1. Admin/contributor can create, update, enable/disable, and delete automation rules per project.
2. On `ticket.state_changed` to target state, matching enabled rules execute all actions in order.
3. `set_state`, `set_assignee`, `set_priority`, `add_comment`, `call_webhook` all function correctly.
4. Failed actions are recorded in execution log with error message; subsequent actions still run.
5. Execution count and last_executed_at are updated after each rule fires.
6. Execution history is queryable per rule via API (last 50).
7. Rules with empty trigger conditions match all events of that type.
8. Engine is non-blocking — ticket mutation response time is unaffected.
9. All CRUD endpoints return correct HTTP status codes; delete is idempotent.
10. Frontend rule builder accurately reflects available trigger events, conditions, and actions using live project data (workflow states, users, webhooks).

#### Key Files
- `migrations/023_automation_rules.sql`
- `store/sql/17_automation.go.templ`
- `store/automation.go`
- `internal/automation/engine.go`
- `openapi.yaml`
- `handlers_automation.go`, `handlers.go` (Store interface + engine wiring), `handlers_test.go`, `types.go`, `map.go`
- `frontend/src/lib/api.ts`
- `frontend/src/stores/automation.ts`
- `frontend/src/views/SettingsPage.vue` (new Automation tab)
- `frontend/src/components/app/automation/RuleBuilder.vue` (new)
- `frontend/src/lib/i18n.ts`
- `e2e/contracts/frontend_contract.source.json`
- `e2e/automation_rules_test.go`

#### Validation Checklist
- `make generate` — no errors
- `go build ./...` — compiles
- `npx tsc --noEmit` — types check
- `go test ./internal/...` — passes
- `make e2e-contract` — contract regenerates
- Manual: create rule → trigger → verify ticket updated + execution log entry visible in UI

---

### TKT-028: Live Collaboration MVP — Presence, Edit Locks, Real-Time Comments
- **Priority:** `P3`
- **Type:** `feature`
- **Status:** `Todo`

#### Problem
Multiple users working simultaneously on the same board or ticket have no visibility of each other and no protection against concurrent edits. The result: accidental overwrites, stale reads, and wasted triage effort when two people resolve the same ticket at the same time.

#### Scope
- **Presence indicators** — show who is currently viewing the board or a specific ticket
- **DB-enforced edit locks** — prevent concurrent saves; second editor gets a read-only view
- **Real-time comment updates** — new comments appear live in open ticket modals without a manual refresh

Out of scope for this ticket: field-level conflict resolution, typing indicators, cross-project presence.

---

#### Backend Changes

**1. In-memory presence store** (`internal/presence/store.go`)
- `PresenceEntry { UserID, UserName, ProjectID, View ("board"|"ticket"), TicketID *string, LastSeen time.Time }`
- TTL-based expiry (30s); clients heartbeat every 15s
- Methods: `Upsert`, `Delete`, `ListByProject`
- Ephemeral — no DB migration needed

**2. DB edit lock table** (`migrations/024_ticket_locks.sql`)
```sql
CREATE TABLE ticket_locks (
    ticket_id    UUID PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
    locked_by    UUID NOT NULL,
    locked_by_name TEXT NOT NULL,
    acquired_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON ticket_locks (expires_at);
```
- Lock TTL: 2 minutes, renewable every 30s
- Expired locks are ignored on read and cleaned up lazily on acquire

**3. Lock REST endpoints**
- `POST /tickets/{ticketId}/lock` — acquire lock
  - Returns `200` with lock record if acquired or already held by this user
  - Returns `409 Conflict` if held by another user:
    ```json
    { "lockedBy": "Alice", "expiresAt": "2026-03-06T18:55:00Z" }
    ```
- `PUT /tickets/{ticketId}/lock` — renew lock (extend TTL); returns `403` if not lock holder
- `DELETE /tickets/{ticketId}/lock` — release lock; returns `403` if not lock holder

**4. Lock enforcement in UpdateTicket handler**
- Before applying any field update, check `ticket_locks` for an active, non-expired lock held by a different user
- If blocked: return `423 Locked` with body `{ "lockedBy": "Alice", "expiresAt": "..." }`
- If lock holder saves successfully: release lock automatically (or keep it — client decides)

**5. New WebSocket event types** (extend existing `events/ws` transport)
- `presence.update`
  ```json
  { "type": "presence.update", "payload": { "users": [{ "userId": "...", "userName": "Alice", "view": "ticket", "ticketId": "..." }] } }
  ```
- `ticket.lock_acquired` — broadcast when a lock is taken
  ```json
  { "type": "ticket.lock_acquired", "payload": { "ticketId": "...", "lockedBy": "Alice", "expiresAt": "..." } }
  ```
- `ticket.lock_released` — broadcast on explicit release or expiry
  ```json
  { "type": "ticket.lock_released", "payload": { "ticketId": "..." } }
  ```
- `ticket.comment_added`
  ```json
  { "type": "ticket.comment_added", "payload": { "ticketId": "...", "comment": { ...CommentResponse } } }
  ```

**6. New REST endpoint — presence heartbeat**
- `POST /projects/{projectId}/presence`
  - Body: `{ "view": "board"|"ticket", "ticketId": "uuid|null" }`
  - Response: `{ "users": [...PresenceEntry] }`

**7. OpenAPI**
- Schemas: `PresenceEntry`, `PresenceRequest`, `PresenceResponse`, `TicketLock`, `TicketLockConflict`
- Document WS event payload shapes in spec comments
- Next migration: `024`, next SQL template: `18`

---

#### Frontend Changes

**1. Presence composable** (`src/composables/usePresence.ts`)
- POST heartbeat every 15s; clear on unmount / route change
- Reactive `presenceUsers` updated from `presence.update` WS events

**2. Board presence bar** (in `BoardPage.vue` toolbar area)
- Avatar row for users currently on this board
- Tooltip per avatar: "Alice — viewing board", "Bob — editing TKT-014"
- Show up to 5 avatars; remainder as "+N others"

**3. Ticket modal edit lock** (`TicketModal.vue`)
- On open: `POST /tickets/{id}/lock`
  - Success → editable, renew every 30s via `PUT /tickets/{id}/lock`
  - 409 → open in **read-only mode** with banner:
    > **Alice has this ticket open for editing.**  
    > Your changes cannot be saved until she closes it or the lock expires ({expiresAt}).
- On save: release lock via `DELETE /tickets/{id}/lock` after successful update
- On close without saving: release lock
- Subscribe to `ticket.lock_released` — if the current ticket's lock is released while the modal is open in read-only mode, show toast: "Alice's lock was released — you can now edit this ticket." and re-attempt acquire

**4. Real-time comments** (`TicketModal.vue`)
- Subscribe to `ticket.comment_added` for the open ticket
- Append new comments (deduplicate by ID); use existing `TransitionGroup` for slide-in animation
- If user has scrolled away from the bottom, show a "1 new comment ↓" pill instead of auto-scrolling

**5. Toast messages** (use existing `useToast`)

| Trigger | Type | Message |
|---------|------|---------|
| Lock acquired successfully | — | *(silent — normal flow)* |
| Lock acquisition fails (409) | warning | "Alice is editing this ticket — opened in read-only mode" |
| Lock released by other user while you wait | success | "Alice's lock expired — you can now edit this ticket" |
| Your lock renewal fails (e.g. session expired) | error | "Your edit session expired. Refresh to continue editing." |
| New comment arrives | info | "Bob added a comment" *(only if modal is not focused)* |
| Save blocked by server (423) | error | "Could not save — Alice re-acquired the edit lock. Reload the ticket to see the latest version." |

---

#### Architecture Notes
- Lock store: Postgres (durable, survives backend restarts; presence store stays in-memory)
- Lock TTL enforced at read time — no background job required
- `HandlerOptions` gets a `PresenceStore` field (same pattern as `AutomationEngine`)
- WS broadcasts reuse existing `broadcastProjectEvent` helper
- No new dependencies

---

#### Definition of Done
- [ ] `024_ticket_locks.sql` migration
- [ ] `PresenceStore` in-memory implementation with TTL
- [ ] `POST/PUT/DELETE /tickets/{id}/lock` endpoints with OpenAPI schemas
- [ ] `POST /projects/{id}/presence` endpoint
- [ ] `423` enforcement in `UpdateTicket` handler
- [ ] WS events: `presence.update`, `ticket.lock_acquired`, `ticket.lock_released`, `ticket.comment_added`
- [ ] Board presence avatar bar
- [ ] Ticket modal: lock acquire on open, renew, release on close/save
- [ ] Ticket modal: read-only mode with lock-held banner when 409
- [ ] Real-time comment append with "new comment" pill for off-screen arrivals
- [ ] Toast messages per table above
- [ ] E2E contract updated; at least 3 E2E tests:
  - Presence avatars appear on board
  - Second user sees read-only mode when lock is held
  - Comment added by one user appears live for the other
- [ ] Feature flag: `VITE_LIVE_COLLAB_ENABLED` (default `false`)

### TKT-029: SLA Targets, Due Dates, and Breach Escalation
- **Priority:** `P1`
- **Type:** `feature`
- **Status:** `Done`
- **Problem:** Teams cannot reliably see which tickets are at risk of missing expected response or completion timelines.
- **Scope:**
  - Add per-project SLA policy config by priority/type (target durations for first response and completion).
  - Add ticket due date field and board/card countdown badges.
  - Add SLA states (`on_track`, `at_risk`, `breached`) computed server-side.
  - Add automation hooks for escalation actions on breach (comment, assign, priority bump, webhook).
  - Add dashboard widgets for breached and at-risk counts.
- **Acceptance Criteria:**
  - SLA status is visible on board cards and ticket modal.
  - Breach transitions emit activities and optionally trigger automation.
  - Dashboard includes trend counts for at-risk and breached tickets.
  - OpenAPI, frontend types, and E2E selectors are updated.

### TKT-030: Custom Fields and Ticket-Type Form Schemas
- **Priority:** `P1`
- **Type:** `feature`
- **Status:** `Done` (March 7, 2026)
- **Problem:** Fixed ticket fields are limiting for teams with domain-specific workflows.
- **Scope:**
  - Add project-level custom field definitions (`text`, `number`, `date`, `enum`, `user`, `boolean`).
  - Add ticket-type schema mapping (which fields appear for `bug`, `feature`, etc.).
  - Add required-by-state rules for validation during transitions.
  - Render dynamic sections in New Ticket and Ticket Modal forms.
  - Include custom fields in filtering and export endpoints.
- **Acceptance Criteria:**
  - Admin can define and reorder custom fields in Settings.
  - Dynamic fields render correctly and persist in ticket payloads.
  - Validation blocks state changes when required fields are missing.
  - API and generated frontend types remain OpenAPI-first.

### TKT-031: WIP Limits and Flow Health Insights
- **Priority:** `P1`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** Teams can overload columns without feedback, increasing cycle time and context switching.
- **Scope:**
  - Add per-workflow-state WIP limit settings.
  - Show board warnings when a state exceeds limit.
  - Add flow health panel: aging WIP, cycle-time percentiles, throughput trend.
  - Add optional enforcement mode preventing drag/drop into overloaded states.
- **Acceptance Criteria:**
  - WIP limits can be configured in workflow settings.
  - Over-limit states are clearly highlighted in board header and story rows.
  - Dashboard flow panel updates from live project data.
  - E2E covers warning mode and enforcement mode behavior.

### TKT-032: Release and Version Management
- **Priority:** `P1-P2`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** Planning and delivery are disconnected; teams cannot package tickets into releasable versions with clear notes.
- **Scope:**
  - Add release entities (`name`, `version`, `status`, `targetDate`, `notes`).
  - Allow linking tickets to a release from board and modal.
  - Auto-generate draft release notes from completed tickets.
  - Add release status transitions (`planned`, `in_progress`, `released`, `archived`).
  - Add release export endpoint (`markdown` and `json`).
- **Acceptance Criteria:**
  - Users can create releases and assign/unassign tickets.
  - Release notes generation includes ticket key, title, type, and notable changes.
  - Released tickets and release history are filterable in board/dashboard.
  - API schemas and frontend route contract are documented in OpenAPI.

### TKT-033: Approval Gates for Sensitive Workflow Transitions
- **Priority:** `P1-P2`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** High-impact transitions (for example to production-ready states) need explicit approval and traceability.
- **Scope:**
  - Add approval policies per workflow transition.
  - Support approver scopes: specific users, groups, or role-based approvers.
  - Add approval requests, decision log, and rejection reasons.
  - Enforce transition blocking until required approvals are met.
  - Integrate with existing audit log and notifications.
- **Acceptance Criteria:**
  - Attempted gated transitions return clear blocked responses when approval is missing.
  - Approval decisions are visible in ticket activity and admin audit log.
  - Approvers get inbox notifications with deep links to approve/reject.
  - E2E covers allow/deny and multi-approver flows.

### TKT-034: Automation Rule Simulator and Dry-Run Replay
- **Priority:** `P2`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** Rule mistakes are risky; users need confidence before enabling automations in production.
- **Scope:**
  - Add simulation mode for rules without mutating live tickets.
  - Replay selected historical events (for example last 7/30 days) against draft rules.
  - Show predicted action outcomes and failure reasons.
  - Add "compare result" diff view for each affected ticket.
  - Store simulation runs for later review.
- **Acceptance Criteria:**
  - Users can run simulation from rule editor before saving/enabling.
  - Simulation output includes matched event count and per-action success/failure prediction.
  - No ticket mutations occur during dry-run.
  - API exposes simulation results with pagination for large runs.

### TKT-035: Shared Operational Views by Team Role
- **Priority:** `P2`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** Personal presets are useful but teams still lack standardized role-based working views.
- **Scope:**
  - Add project-scoped shared views managed by admins/contributors.
  - Support view ownership and permissions (private, team-shared, project-shared).
  - Allow locking parts of view config (required filters/sort/group).
  - Add default-per-role view selection (Support, QA, Product, Engineering).
  - Track usage metrics for views.
- **Acceptance Criteria:**
  - Shared views appear in board "Views" dropdown with ownership metadata.
  - Users can duplicate shared views into personal variants.
  - Locked filters cannot be overridden in restricted views.
  - E2E verifies visibility rules for viewer/contributor/admin.

### TKT-036: Similar Ticket and Knowledge Link Suggestions
- **Priority:** `P2-P3`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** Duplicate tickets and missing historical context increase resolution time.
- **Scope:**
  - Add similar-ticket suggestions during ticket creation/edit.
  - Add knowledge links section for postmortems, runbooks, and related docs.
  - Score suggestions by text similarity + metadata overlap (type, component, labels when available).
  - Allow one-click linking of suggested items into ticket metadata.
  - Add suppression action ("not relevant") to improve ranking.
- **Acceptance Criteria:**
  - Suggestion panel appears in New Ticket and Ticket Modal.
  - Suggestions include confidence indicators and quick-link actions.
  - Linked knowledge items are shown in ticket detail and exports.
  - Performance target: suggestion call under 300ms for common projects.

### TKT-037: Inbound Email to Ticket Gateway
- **Priority:** `P2-P3`
- **Type:** `feature`
- **Status:** `Todo`
- **Problem:** Work requests often arrive by email and are manually copied into the system.
- **Scope:**
  - Add per-project inbound email addresses/aliases.
  - Parse inbound messages into new tickets or comments by thread key.
  - Support basic attachment ingestion and sender mapping.
  - Add anti-spoof checks and allowlist settings.
  - Add processing log with retry status and errors.
- **Acceptance Criteria:**
  - New inbound emails create tickets with subject/body mapping.
  - Replies on known threads append comments to correct ticket.
  - Attachments are stored through existing attachment pipeline.
  - Admin can inspect processing logs and reprocess failed messages.
