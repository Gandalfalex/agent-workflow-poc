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
- Scope:
  - Cross-project roll-up dashboard for open risk, throughput, and SLA breaches.
  - Milestone tracking across projects with confidence indicators.
  - Filters by group, owner, and objective/OKR.
- Acceptance criteria:
  - Supports at least 50 projects without timing out.
  - Drill-down from portfolio KPI to project/ticket details.
  - Snapshot export endpoint for weekly leadership reporting.

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
