# Ticket Drafts (Roadmap Batch 2)

Date: February 17, 2026
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

### TKT-009: Add Ticket Activity Timeline (Backend + UI)
- Priority: `P1`
- Problem:
  - Ticket changes (state moves, assignee changes, field edits) are not recorded or visible.
- Scope:
  - Create `ticket_activities` table: ticket_id, actor_id, action, field, old_value, new_value, created_at.
  - Generate activity entries automatically in ticket update handler.
  - Add API endpoint: `GET /projects/{projectId}/tickets/{id}/activities`.
  - Render chronological timeline in ticket detail modal, interleaved with comments.
  - Add OpenAPI spec entry and generated types.
- Acceptance Criteria:
  - State changes, assignee changes, and priority changes generate activity records.
  - Timeline is visible in ticket modal in chronological order.
  - Activity records are immutable.
  - E2E test validates timeline entries after ticket update.
- Dependencies:
  - Ticket update handlers. Database migration pipeline.

### TKT-010: Build Visual Workflow State Editor
- Priority: `P1`
- Problem:
  - Workflow states can only be edited via raw API. No UI for managing states.
- Scope:
  - Add workflow editor section to settings page.
  - Controls: add state, rename state, delete state, drag-and-drop reorder.
  - State properties: name, isDefault (exactly one required), isClosed.
  - Client-side validation before save. Backend validation on PATCH.
  - Confirmation dialog before deleting a state that has tickets.
- Acceptance Criteria:
  - Users can add, rename, reorder, and delete workflow states from settings.
  - Exactly one default state is enforced.
  - Reordering persists and reflects on the board.
  - Deleting a state with tickets shows a warning/confirmation.
  - E2E test covers add + reorder + save flow.
- Dependencies:
  - Existing `PATCH /projects/{projectId}/workflow` endpoint.

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

### TKT-012: Add Project Dashboard Overview Page — **DONE**
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
2. **Milestone B (Core Features):** TKT-009, TKT-010, ~~TKT-011~~, ~~TKT-012~~
