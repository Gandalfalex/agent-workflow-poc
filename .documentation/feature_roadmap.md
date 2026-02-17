# Feature Roadmap

Date: February 17, 2026
Source baseline: `.documentation/current_features.md`

## Completed

### Local dev environment (was P0)
- Docker Compose fully working with backend, frontend, Postgres, Keycloak, n8n, codex-agent.
- Migrations run automatically on startup. Keycloak realm imported via volume.
- Production compose with Traefik also available.

### Board and workflow regression (was P0)
- 12+ E2E test files with contract-driven Playwright harness.
- Coverage: ticket CRUD, state transitions, stories, comments, drag-and-drop, webhook events, navigation, unhappy paths.
- No critical regressions remain.

### Webhook delivery (was P0, complete)
- Dispatcher, HMAC signing, event filtering, async delivery all working.
- E2E tests validate `ticket.created` and `ticket.state_changed` events.
- Exponential backoff retry (3 attempts: immediate, 30s, 5min) with delivery logging.
- `webhook_deliveries` table tracks every attempt with status, response, duration.
- Delivery history API endpoint and settings UI panel with expandable detail rows.

### Role-based access control (was P1, complete)
- Admin/contributor/viewer role hierarchy enforced.
- `requireAdmin()` and `requireProjectAccess()` middleware in place.
- Per-operation `requireProjectRole()` enforcement: viewers read-only, contributors CRUD tickets, admins manage settings.
- Frontend role-aware UI gating (read-only modals, hidden controls, restricted tabs).
- Multi-user E2E tests with RBAC negative-path coverage.
- Remaining gap: no audit trail.

### Markdown editor upgrade (was P1-P2)
- Reusable `MarkdownEditor.vue` with toolbar, keyboard shortcuts, and preview toggle.
- Applied to all markdown text areas: ticket description, comments, new ticket, story.
- No new dependencies — uses existing `marked` + `lucide-vue-next`.
- Remaining gap: image paste support.

### Project dashboard (was P1, complete)
- Project-level statistics API and dashboard page.
- Aggregate ticket counts by state, priority, type, and assignee.
- Dashboard tab in header navigation alongside Board and Settings.

### Ticket file attachments (was P1)
- MinIO S3-compatible object storage with swappable `ObjectStore` interface.
- Upload, list, download, delete via REST API (multipart form upload, streaming download).
- `ticket_attachments` table with metadata in Postgres, blobs in MinIO.
- Frontend UI in ticket modal: file picker, attachment list with download links, delete buttons.
- In-memory ObjectStore for E2E tests (no MinIO container needed).
- Docker Compose `minio` service added. 10MB configurable upload limit.
- 2 E2E tests: upload+list, delete.
- Remaining gap: no Nginx CDN caching layer (downloads served through backend).

---

## Roadmap Goals
- Complete remaining gaps in webhooks and RBAC.
- Add high-impact product features for planning, visibility, and usability.
- Improve frontend polish and self-service administration.

## Prioritization
- `P0` Remaining gaps that affect reliability or security.
- `P1` High-value features with clear user benefit.
- `P2` Nice-to-have enhancements after core features land.

---

## Phase 1 (P0): Close Remaining Gaps

### ~~1. Webhook retry and delivery logs~~ ✓ Completed (TKT-007)
- ~~Add exponential backoff retry on failed deliveries (max 3 attempts).~~
- Implemented: 3-attempt retry with backoff (0s, 30s, 5min), `webhook_deliveries` table, delivery history API + settings UI panel.

### ~~2. Granular role-based permission enforcement~~ ✓ Completed (TKT-008)
- ~~Enforce per-operation checks: viewers cannot create/edit/delete tickets; contributors cannot manage projects/groups.~~
- Implemented: `requireProjectRole()` helper with role rank system, 17+ handler patches, `GET /my-role` endpoint, frontend UI gating, 6 unit tests, 4 E2E tests.

## Phase 2 (P1): Core Product Features

### 3. Ticket activity timeline
- Add immutable activity records for state changes, assignee changes, field edits.
- New `ticket_activities` table with migration.
- API endpoint to retrieve activity for a ticket.
- Render chronological timeline in ticket detail modal.
- Why: Comments exist but change history is not auditable.

### 4. Workflow editor UI
- Visual workflow state editor in settings: add, rename, reorder, set default/closed flags.
- Drag-and-drop state reordering.
- Client-side and backend validation (must have exactly one default state).
- Why: Workflow API exists but editing requires raw API calls.

### ~~5. Ticket attachments with MinIO~~ ✓ Completed (TKT-011)
- ~~MinIO as S3-compatible object storage, Nginx as caching CDN layer in front.~~
- Implemented: MinIO blob storage, REST API (upload/list/download/delete), frontend UI, E2E tests.
- Remaining: Nginx CDN caching layer for repeat downloads.

### ~~6. Dashboard and project overview page~~ ✓ Completed (TKT-012)
- ~~Project-level dashboard showing: open ticket count by state, ticket count by priority, recent activity.~~
- Implemented: Stats API endpoint, dashboard page with summary cards and bar charts by state/priority/type/assignee.
- Remaining: Recent activity feed (depends on TKT-009 activity timeline).

## Phase 3 (P1-P2): Collaboration and Productivity

### 7. @mention notifications
- Parse `@username` in comments and trigger in-app notifications.
- Notification list accessible from header with unread count.
- Notification on ticket assignment changes.
- Why: Increases responsiveness without needing external integrations.

### 8. Saved board filters
- Save named filter presets (assignee, type, priority, state combinations).
- Quick-switch between personal filter presets.
- Filter state persists across page refresh.
- Why: Existing search resets on every page load.

### 9. Bulk ticket operations
- Multi-select tickets on the board.
- Bulk actions: move to state, assign user, change priority, delete.
- Permission checks applied per operation.
- Why: Managing larger backlogs one ticket at a time is slow.

### ~~10. Markdown editor upgrade~~ ✓ Completed
- ~~Replace plain textarea with a toolbar-equipped markdown editor.~~
- Implemented: Reusable `MarkdownEditor.vue` component with toolbar (bold, italic, code, link, lists, quote, heading), keyboard shortcuts (Ctrl+B/I/E/K, Tab/Shift+Tab), edit/preview toggle. Applied to TicketModal (description + comments), NewTicketModal, and StoryModal. Zero new dependencies.
- Remaining: image paste support.

## Phase 4 (P2): Integrations and Reporting

### 11. Outbound webhook payload versioning
- Add `v1` envelope with schema version, event timestamp, idempotency key.
- Document payload schema per event type.
- Optional per-event-type subscription granularity.
- Why: Integrators need stable, documented contracts.

### 12. Lightweight project reporting
- Basic metrics: ticket throughput, average cycle time, open-by-state over time.
- Read-only reporting endpoint and settings page panel.
- Why: Teams need sprint/project health without external BI tools.

### 13. Email and Slack notification channels
- Configurable notification delivery: in-app, email, Slack webhook.
- Per-user notification preferences.
- Why: Not everyone watches the app in real time.

---

## Recommended Next 5 Tickets
1. ~~Webhook retry logic and delivery log table (P0).~~ ✓ Done (TKT-007)
2. ~~Granular per-operation RBAC enforcement (P0, TKT-008).~~ ✓ Done
3. Ticket activity timeline - backend + UI (P1, TKT-009).
4. Workflow editor UI with drag-and-drop states (P1, TKT-010).
5. ~~Project dashboard overview page (P1, TKT-012).~~ ✓ Done

## Risks and Dependencies
- Schema changes for activity timeline need migration planning.
- Notification features depend on a notification infrastructure decision (polling vs WebSocket).
- Feature throughput depends on maintaining OpenAPI-first workflow and generated type sync.

## Definition of Done
- API behavior implemented and covered by automated tests.
- Frontend UX added/updated with loading and error states.
- Documentation updated in `.documentation/`.
- E2E contract updated and tests pass.
