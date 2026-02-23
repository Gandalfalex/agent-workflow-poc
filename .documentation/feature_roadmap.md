# Feature Roadmap

Date: February 20, 2026
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
- Introduce automation and intelligence features for high-scale teams.

## Prioritization
- `P0` Remaining gaps that affect reliability or security.
- `P1` High-value features with clear user benefit.
- `P2` Nice-to-have enhancements after core features land.
- `P3` Experimental and moonshot ideas with high upside.

---

## Phase 1 (P0): Close Remaining Gaps

### ~~1. Webhook retry and delivery logs~~ ✓ Completed (TKT-007)
- ~~Add exponential backoff retry on failed deliveries (max 3 attempts).~~
- Implemented: 3-attempt retry with backoff (0s, 30s, 5min), `webhook_deliveries` table, delivery history API + settings UI panel.

### ~~2. Granular role-based permission enforcement~~ ✓ Completed (TKT-008)
- ~~Enforce per-operation checks: viewers cannot create/edit/delete tickets; contributors cannot manage projects/groups.~~
- Implemented: `requireProjectRole()` helper with role rank system, 17+ handler patches, `GET /my-role` endpoint, frontend UI gating, 6 unit tests, 4 E2E tests.

## Phase 2 (P1): Core Product Features

### ~~3. Ticket activity timeline~~ ✓ Completed (TKT-009)
- Implemented: `ticket_activities` table, `GET /tickets/{id}/activities` endpoint, auto-recording on ticket update (state, priority, assignee, type, title), timeline section in TicketModal.vue, 2 E2E tests.
- Remaining: Dashboard recent activity feed (can now be built on top of this).

### ~~4. Workflow editor UI~~ ✓ Completed (TKT-010)
- ~~Visual workflow state editor in settings: add, rename, reorder, set default/closed flags.~~
- Implemented: Workflow editor tab in settings with add/rename/delete/drag-reorder, isDefault radio, isClosed checkbox, client-side validation, confirmation dialog on delete, 4 E2E tests.

### ~~5. Ticket attachments with MinIO~~ ✓ Completed (TKT-011)
- ~~MinIO as S3-compatible object storage, Nginx as caching CDN layer in front.~~
- Implemented: MinIO blob storage, REST API (upload/list/download/delete), frontend UI, E2E tests.
- Remaining: Nginx CDN caching layer for repeat downloads.

### ~~6. Dashboard and project overview page~~ ✓ Completed (TKT-012)
- ~~Project-level dashboard showing: open ticket count by state, ticket count by priority, recent activity.~~
- Implemented: Stats API endpoint, dashboard page with summary cards and bar charts by state/priority/type/assignee.
- Remaining gap closed: Recent activity feed added (February 19, 2026) — project-scoped feed with ticket context on dashboard.

## Phase 3 (P1-P2): Collaboration and Productivity

### ~~7. @mention notifications~~ ✓ Completed (TKT-015)
- Implemented: notification inbox in header with unread badge, per-item list, mark-read APIs, and mark-all-read support.
- Triggers: `@username` mentions in ticket comments and ticket description updates, plus assignment change notifications.
- Preferences: current-user notification preferences support mention-only / assignment-only behavior.
- API:
  - `GET /projects/{projectId}/notifications`
  - `GET /projects/{projectId}/notifications/unread-count`
  - `POST /projects/{projectId}/notifications/{notificationId}/read`
  - `POST /projects/{projectId}/notifications/read-all`
  - `GET/PATCH /projects/{projectId}/notification-preferences`

### ~~8. Saved board filters~~ ✓ Completed (TKT-013)
- Implemented: personal board filter presets with save/rename/delete/apply, quick-select toolbar controls, share-token links, and persisted active preset on reload.
- API: `GET/POST /projects/{projectId}/board-filters`, `PATCH/DELETE /projects/{projectId}/board-filters/{presetId}`, `GET /projects/{projectId}/board-filters/shared/{token}`.

### ~~9. Bulk ticket operations~~ ✓ Completed (TKT-014)
- Implemented: board multi-select mode with selected-count badge, bulk action toolbar, and optimistic update UX with per-ticket rollback on partial failures.
- Actions: `move_state`, `assign`, `set_priority`, `delete`.
- API: `POST /projects/{projectId}/tickets/bulk` with per-ticket success/error results.
- E2E: admin UI bulk flow (move/assign/priority/delete) and viewer per-ticket permission-failure summary coverage.

### ~~10. Dependency graph and blocked work~~ ✓ Completed (TKT-016)
- Implemented: ticket dependencies with relation types (`blocks`, `blocked_by`, `related`) and cycle-safe creation.
- API:
  - `GET /projects/{projectId}/dependency-graph?rootTicketId=&depth=`
  - `GET /tickets/{id}/dependencies`
  - `POST /tickets/{id}/dependencies`
  - `DELETE /tickets/{id}/dependencies/{dependencyId}`
- UI:
  - Ticket modal dependency manager + 2-hop graph summary.
  - Dashboard dependency graph panel.
  - Board blocked badge and blocked-only filter.
  - Dashboard `blockedOpen` metric in stats cards.
- E2E: dependency creation, cycle rejection, blocked filter behavior, and dashboard graph visibility.

### ~~11. Markdown editor upgrade~~ ✓ Completed
- ~~Replace plain textarea with a toolbar-equipped markdown editor.~~
- Implemented: Reusable `MarkdownEditor.vue` component with toolbar (bold, italic, code, link, lists, quote, heading), keyboard shortcuts (Ctrl+B/I/E/K, Tab/Shift+Tab), edit/preview toggle. Applied to TicketModal (description + comments), NewTicketModal, and StoryModal. Zero new dependencies.
- Remaining: image paste support.

## Phase 4 (P2): Integrations and Reporting

### ~~11. Outbound webhook payload versioning~~ ✓ Completed (TKT-023)
- Implemented: `v1` outbound envelope (`version`, `event`, `eventTimestamp`, `idempotencyKey`, `data`) and matching delivery headers (`X-Ticketing-Webhook-Version`, `X-Ticketing-Idempotency-Key`).
- Implemented: OpenAPI-documented payload schemas per webhook event type.
- Why: Integrators get stable, explicit delivery contracts and replay-safe idempotency.

### ~~12. Lightweight project reporting~~ ✓ Completed
- Implemented: reporting summary endpoint, settings reporting tab, and export endpoint (`json`/`csv`) with E2E coverage.
- API:
  - `GET /projects/{projectId}/reporting/summary`
  - `GET /projects/{projectId}/reporting/export?format=json|csv`

### 13. Real-time live updates (WebSocket transport)
- Replace periodic polling with WebSocket push for notification unread count, inbox updates, board refresh cues, and activity feed updates.
- Keep polling as fallback behind a feature flag until rollout is stable.
- Why: Lower latency and lower request overhead than fixed-interval polling.

Deferred (later): Multi-channel notifications (Slack/Teams/email)
- Defer channel-specific delivery until channel strategy is finalized.
- Keep current in-app notifications as primary channel for now.

## Phase 5 (P2-P3): Automation, Intelligence, and Scale

### 14. Dependency graph and blocked-work detection
- Add explicit ticket dependencies (`blocks`, `blocked_by`, `related`).
- Visual graph view and automatic "blocked" badge on board cards.
- Cross-story and cross-project dependency links.
- Why: Teams lose time when hidden dependencies stall delivery.

### 15. Rule-based automation engine
- If/then automation rules at project scope (e.g., "when moved to Done, assign QA group").
- Actions: set state, set assignee, set priority, add comment, call webhook.
- Dry-run mode and execution history for auditability.
- Why: Repetitive triage and handoff work should be automated.

### 16. Sprint planner with capacity simulation
- Plan a sprint by dragging tickets into a candidate sprint bucket.
- Team member capacity settings and workload heatmap.
- Forecast commit confidence with simple Monte Carlo simulation.
- Why: Better planning quality and fewer overcommitted sprints.

### 17. AI-assisted triage copilot
- Suggest priority, assignee, and workflow state from title/description/context.
- Auto-summarize long ticket threads and produce "next best action".
- Provide confidence score and require explicit user confirmation.
- Why: Speeds up intake while preserving human control.

### 18. Live collaboration mode
- Presence indicators in board/ticket views.
- Soft locks and conflict hints during concurrent editing.
- Real-time comment and activity updates without refresh.
- Why: Reduces accidental overwrites and stale decision-making.

### ~~19. Incident bridge integration~~ ✓ Completed (TKT-020)
- Implemented: incident-mode ticket fields, aggregated incident timeline endpoint, markdown postmortem export endpoint, severity-change audit activity, and ticket-modal incident controls/export UX.

### 20. Portfolio command center
- Multi-project dashboard with roll-up KPIs and risk scoring.
- Cross-project milestone tracking with drill-down.
- Objective/OKR linkage to stories and tickets.
- Why: Leadership needs portfolio visibility, not just project-level views.

### 21. Plugin marketplace and app extensions
- Safe extension points for custom panels, commands, and automation actions.
- Scoped API keys and permission model for third-party apps.
- In-app install/update flow for vetted plugins.
- Why: Enables domain-specific workflows without forking the core product.

### 22. Time-travel board replay
- Replay board evolution over a selected date range.
- Highlight transitions, churn hotspots, and bottlenecks.
- Export replay snapshots for sprint review.
- Why: Makes process issues visible and measurable.

---

## Recommended Next 5 Tickets
1. Attachment download caching layer (Nginx/CDN fronting backend) (P2)
2. RBAC/admin audit trail for sensitive actions (P2)
3. TKT-021: Portfolio command center (P2)
4. Rule-based automation engine (P2-P3)
5. Live collaboration mode (P3)

## Risks and Dependencies
- Schema changes for activity timeline need migration planning.
- Real-time collaboration still depends on robust WebSocket lifecycle handling (reconnect, auth refresh, backpressure, and conflict-aware UX).
- Channel integrations (Slack/Teams/email) are intentionally deferred pending product/channel decisions.
- Feature throughput depends on maintaining OpenAPI-first workflow and generated type sync.
- AI-assisted features require prompt/version governance and careful data privacy boundaries.
- Real-time collaboration features will build on the same transport layer and still need a presence/conflict state model.
- Automation engine needs strong guardrails to avoid rule loops and privilege escalation.

## Definition of Done
- API behavior implemented and covered by automated tests.
- Frontend UX added/updated with loading and error states.
- Documentation updated in `.documentation/`.
- E2E contract updated and tests pass.
