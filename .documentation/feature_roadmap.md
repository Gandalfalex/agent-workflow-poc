# Feature Roadmap

Date: February 19, 2026
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

### 19. Incident bridge integration
- Convert tickets to incidents with severity, timeline, and owner.
- Integrate with on-call channels (webhook/Slack/Pager workflows).
- Auto-generate postmortem draft from activity and comments.
- Why: Unifies planned work and unplanned operational incidents.

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
1. Saved board filters with persistence + share links (P1)
2. Bulk ticket operations with permission-safe multi-select (P1)
3. @mention notifications and unified inbox (P1)
4. Dependency graph with blocked-work highlighting (P2)
5. Rule-based automation engine (P2)

## Risks and Dependencies
- Schema changes for activity timeline need migration planning.
- Notification features depend on a notification infrastructure decision (polling vs WebSocket).
- Feature throughput depends on maintaining OpenAPI-first workflow and generated type sync.
- AI-assisted features require prompt/version governance and careful data privacy boundaries.
- Real-time collaboration features require transport decisions (SSE vs WebSocket) and presence state model.
- Automation engine needs strong guardrails to avoid rule loops and privilege escalation.

## Definition of Done
- API behavior implemented and covered by automated tests.
- Frontend UX added/updated with loading and error states.
- Documentation updated in `.documentation/`.
- E2E contract updated and tests pass.
