# Implementation Plan

## Phase 1: Foundation
1. Define `openapi.yaml` with auth, projects, groups, tickets, workflow, and webhooks.
2. Add OpenAPI generators for backend + frontend and wire generation scripts.
3. Set up Go service with database migrations and baseline models.
4. Create Vue 3 + TypeScript app with ShadCN-style UI kit.

## Phase 2: Core Features
5. Implement auth (login/logout/session via Keycloak).
6. Implement projects + groups CRUD and membership management (store + handlers).
7. Implement ticket CRUD and Kanban board API scoped per project.
8. Add ticket types, story grouping, and comments (API + persistence).
9. Build Kanban UI with drag-and-drop and ticket detail drawer.
10. Implement workflow state configuration UI + API per project.

## Phase 3: Webhooks
11. Add project-scoped webhook config API + persistence.
12. Emit webhook events on ticket create/update/state change.
13. Add webhook management UI (create/edit/test).

## Phase 4: Polish
14. Add basic error handling, empty states, and loading UI.
15. Add seed data for local dev (projects, groups, roles).
16. Write minimal docs and runbook.
