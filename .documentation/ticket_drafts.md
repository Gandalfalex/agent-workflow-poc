# Ticket Drafts (Roadmap Batch 1)

Date: February 15, 2026  
Source: `.documentation/feature_roadmap.md`

## TKT-001: Verify Webhook Delivery End-to-End and Add Retry Coverage
- Priority: `P0`
- Problem:
  - Webhook CRUD and dispatch exist, but end-to-end reliability is not fully validated.
- Scope:
  - Validate webhook events for `ticket.created`, `ticket.updated`, `ticket.deleted`, `ticket.state_changed`.
  - Add/expand integration tests for failure + retry + success flows.
  - Verify request signing behavior when secret is configured.
  - Add structured delivery logs with status and attempt count.
- Acceptance Criteria:
  - Events are emitted for all supported ticket event types.
  - Failed webhook deliveries are retried according to configured policy.
  - Tests cover at least one transient failure scenario and one permanent failure scenario.
  - When a webhook secret is set, outbound requests include expected signature header.
  - Logs include webhook id, event type, delivery outcome, and attempt number.
- Dependencies:
  - Existing webhook dispatcher/store implementation.
  - Local test receiver or mock endpoint for integration tests.

## TKT-002: Run Full Board/Workflow Regression and Fix Critical Bugs
- Priority: `P0`
- Problem:
  - Core board features are present, but QA and integration confidence is incomplete.
- Scope:
  - Execute regression checklist for ticket CRUD, state transitions, stories, comments, assignee updates, search/filter.
  - Validate workflow initialization and updates from UI/API.
  - Fix critical and high-severity defects found during testing.
  - Add test coverage for identified regressions.
- Acceptance Criteria:
  - Regression checklist is documented and completed with results.
  - No open critical defects for board/workflow flows.
  - All fixed regressions have corresponding automated tests.
  - Board operations remain functional in both happy-path and common edge cases.
- Dependencies:
  - Auth-ready local environment with backend, frontend, database, Keycloak.

## TKT-003: Finalize Local Dev Compose and Runbook
- Priority: `P0`
- Problem:
  - Local setup is partially complete and can cause inconsistent onboarding/verification.
- Scope:
  - Consolidate/validate local compose flow for backend, frontend, Postgres, and Keycloak.
  - Document startup, migration, seed/sync, and shutdown commands.
  - Ensure one-command bootstrap path is available and tested.
  - Document common failure modes and recovery steps.
- Acceptance Criteria:
  - A new contributor can start the full stack using documented steps only.
  - Migrations apply successfully in a clean environment.
  - Auth login flow works in local setup.
  - Runbook includes troubleshooting for at least 3 common setup failures.
- Dependencies:
  - Existing docker-compose files and setup scripts.

## TKT-004: Audit and Enforce Role-Based Authorization Matrix
- Priority: `P1`
- Problem:
  - Project-group roles exist, but permission enforcement may not be uniformly tested across endpoints.
- Scope:
  - Define explicit permission matrix for `viewer`, `contributor`, `admin`.
  - Audit protected endpoints for expected authorization checks.
  - Add negative tests for forbidden operations.
  - Align frontend controls with backend authorization (hide/disable disallowed actions).
- Acceptance Criteria:
  - Permission matrix is documented and committed.
  - Protected endpoints return correct `403` responses for unauthorized roles.
  - Tests validate at least one blocked action per role where applicable.
  - UI does not expose actions users cannot perform.
- Dependencies:
  - Current auth and project/group role assignment behavior.

## TKT-005: Add Ticket Activity Timeline (Backend + UI)
- Priority: `P1`
- Problem:
  - Comments exist, but ticket change history is not fully visible/auditable.
- Scope:
  - Add immutable activity records for state changes, assignee changes, and key field edits.
  - Expose activity timeline via API endpoint or ticket detail expansion.
  - Render timeline in ticket modal with timestamp and actor.
  - Add migration(s) and store logic for activity persistence.
- Acceptance Criteria:
  - Activity entries are created automatically for tracked ticket changes.
  - Timeline is visible in ticket detail UI in chronological order.
  - Activity records are immutable after creation.
  - API and UI tests cover timeline creation and display.
- Dependencies:
  - Ticket update handlers and ticket modal UI.
  - Database migration pipeline.

## TKT-006: Improve Workflow Editor UX and Validation
- Priority: `P1`
- Problem:
  - Workflow management works but needs stronger UX and validation guardrails.
- Scope:
  - Improve workflow editor controls for add, rename, reorder, default/closed state flags.
  - Add client-side validation before save.
  - Add backend validation for invalid workflow definitions.
  - Improve error messaging and recovery in settings UI.
- Acceptance Criteria:
  - Users can safely edit workflows without creating invalid state sets.
  - Saving invalid workflow configurations is blocked with clear error messages.
  - At least one default workflow state is always enforced.
  - Reordering states persists and reflects correctly on board view.
- Dependencies:
  - Existing workflow API and settings page.

## Suggested Labels (Optional)
- `area/backend`
- `area/frontend`
- `area/auth`
- `area/webhooks`
- `area/workflow`
- `priority/p0`
- `priority/p1`

## Suggested Milestone Split
1. Milestone A (Stabilization): `TKT-001`, `TKT-002`, `TKT-003`
2. Milestone B (Completeness): `TKT-004`, `TKT-005`, `TKT-006`
