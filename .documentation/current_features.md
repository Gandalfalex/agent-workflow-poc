# Current Features

Snapshot date: February 15, 2026

## Core Platform
- OpenAPI-defined backend API (`ticketing-system/openapi.yaml`) with generated backend/frontend types.
- Go backend + PostgreSQL persistence with migrations.
- Vue 3 + TypeScript frontend using Pinia stores and route-based project views.

## Authentication
- Session-based authentication for web UI.
- Login endpoint (`/auth/login`) and logout endpoint (`/auth/logout`).
- Current user endpoint (`/auth/me`) for session restoration.
- Keycloak-backed auth integration in backend.

## Projects and Access Control
- Project CRUD:
  - List, create, get, update, delete projects.
- Group CRUD:
  - List, create, get, update, delete groups.
- Group membership management:
  - List members, add member, remove member.
- Project-group role mapping:
  - List project groups, add group to project, update role, remove group from project.
- User directory search endpoint for assignment/member workflows (`/users?q=`).

## Ticketing and Board
- Kanban board API and UI for project board view.
- Ticket CRUD:
  - List tickets, create ticket, get ticket, update ticket, delete ticket.
- Ticket fields supported in UI/API:
  - Title, description, priority, type (`feature`/`bug`), state, assignee, story linkage.
- Ticket key/number model supported by backend schema and migrations.
- Story support:
  - List stories, create story, get story, update story, delete story.
  - Board groups tickets under stories in the UI.
- Ticket comments:
  - List ticket comments, add ticket comment, delete ticket comment.

## Workflow Management
- Workflow state retrieval and update per project.
- UI flow to initialize default workflow states when needed.

## Webhooks
- Project-scoped webhook management:
  - List, create, get, update, delete webhooks.
- Webhook test endpoint support.
- UI settings support for:
  - Creating webhooks, enabling/disabling, and testing delivery.

## Admin/Operations
- Admin endpoint to sync users from Keycloak to local database (`/admin/sync-users`).
- Health check endpoint (`/health`).
- Docker/dev orchestration files for local and production-like setup.

## Frontend UX Currently Present
- Login view and session bootstrap.
- Project board page with:
  - Ticket/stories display by workflow state.
  - Ticket detail modal/editor.
  - New ticket modal.
  - Story creation modal.
  - Board search/filtering.
- Settings page with:
  - Project and group management actions.
  - Group member management.
  - Project role assignment for groups.
  - Webhook management.

## Noted Gaps / In-Progress Areas
- `ticketing-system/STATUS.md` marks some areas as partially complete/integration pending:
  - End-to-end webhook integration verification.
  - Final QA and some dev setup cleanup.
- Status file still includes duplicate older unchecked items even though much functionality is implemented.
