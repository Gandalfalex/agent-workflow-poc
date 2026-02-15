# Feature Roadmap

Date: February 15, 2026
Source baseline: `.documentation/current_features.md`

## Roadmap Goals
- Stabilize and productionize what already exists.
- Close known integration gaps (especially webhooks and QA).
- Add high-impact product features that improve planning, execution, and visibility.

## Prioritization Method
- `P0` Critical foundation or release blocker.
- `P1` High-value feature with moderate implementation risk.
- `P2` Valuable enhancement after core stability and workflow completion.

## Phase 1 (P0): Stabilization and Release Readiness

### 1. End-to-end webhook reliability
- Scope:
  - Verify event emission paths for create/update/delete/state-change.
  - Add retry/backoff behavior validation and delivery result logging.
  - Expand integration tests covering signature/secret behavior.
- Why now:
  - Webhooks are partially complete and explicitly marked as pending integration verification.
- Exit criteria:
  - Webhook events are consistently delivered in local and staging.
  - Test coverage includes success/failure/retry scenarios.
  - Operational troubleshooting info is available in logs.

### 2. Workflow and board QA hardening
- Scope:
  - Full regression pass for ticket CRUD, state moves, stories, comments, assignee, and search.
  - Validate default workflow initialization and edit behavior.
  - Fix edge-case bugs discovered during QA.
- Why now:
  - Core workflow is functional but needs confidence for wider use.
- Exit criteria:
  - No critical regressions in board flows.
  - Smoke test checklist passes in clean environment.

### 3. Local developer environment consistency
- Scope:
  - Reconcile and finalize local `docker-compose` path for ticketing backend/frontend/db/keycloak.
  - Document single-command startup and seed/sync steps.
- Why now:
  - Faster onboarding and repeatable QA require one reliable dev setup.
- Exit criteria:
  - New contributor can run full stack from docs without manual fixes.

## Phase 2 (P1): Core Product Completeness

### 4. Role-based permission enforcement audit
- Scope:
  - Verify backend authorization checks align with project-group role model.
  - Ensure viewer/contributor/admin behavior is enforced consistently in API and UI.
  - Add missing negative tests (forbidden access) for critical endpoints.
- Why now:
  - Access control is central to multi-project usage and security.
- Exit criteria:
  - Permission matrix is documented and test-backed.
  - Unauthorized operations return correct error codes.

### 5. Ticket activity timeline
- Scope:
  - Add immutable activity entries for state changes, assignee changes, and key field edits.
  - Show ticket history in the ticket modal.
- Why now:
  - Current comments exist, but users need auditable change context.
- Exit criteria:
  - Timeline entries are generated automatically for tracked actions.
  - Users can view chronological ticket history in UI.

### 6. Workflow administration UX improvements
- Scope:
  - Improve workflow editor for add/reorder/rename/close-state controls.
  - Guardrails for invalid workflow states (for example, no default state).
- Why now:
  - Workflow API exists; admin UX can be made safer and faster.
- Exit criteria:
  - Workflow updates are intuitive and validated before save.
  - Error states are clear and recoverable in UI.

## Phase 3 (P1-P2): Planning and Team Productivity

### 7. Backlog planning enhancements
- Scope:
  - Story-centric backlog view with ticket counts and basic progress indicators.
  - Bulk actions for moving tickets between states or assigning users.
- Why now:
  - Improves throughput for teams managing larger ticket sets.
- Exit criteria:
  - Backlog can be planned without opening each ticket individually.
  - Bulk operations work with permission checks.

### 8. Saved filters and board views
- Scope:
  - Save named filters (assignee, type, priority, state).
  - Quick switching between personal/team presets.
- Why now:
  - Existing search is useful but not persistent for daily workflows.
- Exit criteria:
  - Users can create, apply, and delete saved filters.
  - Filter state persists across refresh/session.

### 9. Notifications and mention basics
- Scope:
  - Add comment mentions (for example `@user`) and in-app notification list.
  - Trigger notifications for assignment and mention events.
- Why now:
  - Increases responsiveness without needing external integrations first.
- Exit criteria:
  - Mentioned/assigned users see actionable notifications.
  - Notifications link back to relevant ticket context.

## Phase 4 (P2): Integrations and Reporting

### 10. Outbound integration expansion
- Scope:
  - Add richer webhook event payloads and versioned schema notes.
  - Optional event subscriptions by type granularity.
- Exit criteria:
  - Integrators can reliably consume versioned payload contracts.

### 11. Lightweight reporting
- Scope:
  - Basic project metrics: ticket throughput, cycle time estimates, open-by-state.
  - Read-only dashboard endpoint + settings page panel.
- Exit criteria:
  - Teams can inspect sprint/project health without external BI tooling.

## Recommended Execution Order (Next 6 Tickets)
1. Webhook integration verification and retry test coverage (P0).
2. End-to-end board/workflow regression pass with bug fixes (P0).
3. Finalize local docker/dev runbook and startup scripts (P0).
4. Role-based permission audit and forbidden-path tests (P1).
5. Ticket activity timeline (backend + UI) (P1).
6. Workflow editor UX/validation improvements (P1).

## Risks and Dependencies
- Keycloak and local auth setup remain a dependency for reliable end-to-end testing.
- Schema changes for timeline/reporting need migration planning to avoid breaking existing data.
- Feature throughput depends on maintaining OpenAPI-first workflow and generated client/server sync.

## Definition of Done for Each Roadmap Item
- API behavior implemented and covered by automated tests.
- Frontend UX added/updated with loading and error states.
- Documentation updated (feature docs + operator/developer notes).
- Manual verification checklist completed in local docker environment.
