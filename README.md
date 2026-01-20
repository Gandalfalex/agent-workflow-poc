# Coding Agent Workflow - Complete System

A comprehensive system for autonomous feature implementation, ticket management, and workflow automation using Claude AI and n8n.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Architecture](#architecture)
3. [Features](#features)
4. [Skills](#skills)
5. [Configuration](#configuration)
6. [User Management](#user-management)
7. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Prerequisites

- Node.js 18+ (frontend)
- Go 1.21+ (backend)
- Docker & Docker Compose
- Claude CLI installed and authenticated (for feature implementation)

### Installation & Deployment

```bash
# Clone and navigate to project
cd ~/projects/coding-agent-workflow

# Build and start all services
docker-compose up -d

# Services will be available at:
# - Frontend: http://localhost:5173
# - Backend API: http://localhost:8080
# - n8n: http://localhost:5678
# - Keycloak: http://localhost:8081
```

### Testing

```bash
# 1. Log in to frontend
# Go to http://localhost:5173
# Email: ich@ich.ich
# Password: admin123

# 2. Create a project and tickets

# 3. Test feature implementation
# Settings → Groups → Add members (uses fuzzy search)
# Create a feature ticket and assign to a group

# 4. Via SSH (for n8n integration)
ssh user@host "bash ~/projects/coding-agent-workflow/codex-agent/scripts/skill-cli.sh \"implement PROJ-001\""
```

---

## Architecture

### Three-Layer System

```
┌─────────────────────────────────────────────────────┐
│ Frontend (Vue 3)                                    │
│ - Board view (kanban)                              │
│ - Ticket management                                 │
│ - Settings (groups, users, webhooks)               │
└────────────────┬────────────────────────────────────┘
                 │
┌─────────────────────────────────────────────────────┐
│ Backend API (Go)                                    │
│ - Ticket CRUD & state management                   │
│ - Group & user management                          │
│ - Webhook dispatch                                 │
│ - Keycloak integration                             │
└────────────────┬────────────────────────────────────┘
                 │
┌─────────────────────────────────────────────────────┐
│ Codex Agent MCP Server (TypeScript/Go)             │
│ - 7 MCP tools for ticket operations                │
│ - Feature implementation via Claude subagents      │
│ - Keycloak OAuth2 authentication                   │
│ - Go CLI standalone option                         │
└─────────────────────────────────────────────────────┘
```

### Technology Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | Vue 3, TypeScript, Tailwind | User interface |
| **Backend** | Go, PostgreSQL, Keycloak | API & authentication |
| **Codex Agent** | TypeScript/Go, MCP SDK | Claude integration & MCP tools |
| **Automation** | n8n, Webhooks | Workflow orchestration |
| **Auth** | Keycloak + OAuth2 | Identity management |

---

## Features

### 1. Ticket Management
- **Board View**: Kanban-style ticket organization by state
- **Ticket CRUD**: Create, read, update, delete tickets
- **State Management**: Workflow-based ticket states
- **Comments**: Ticket discussions and updates
- **Stories**: Group related tickets

### 2. Group & User Management
- **Groups**: Create groups and assign users
- **Fuzzy Search**: Intelligent user search (via API)
  - Matches: `ich` → `ich@ich.ich`, `admin` → `AdminUser`
  - Case-insensitive, searches name and email
  - Results sorted by relevance
- **Project Access Control**: Assign groups to projects with roles (Admin, Contributor, Viewer)

### 3. Autonomous Feature Implementation

**Available Implementations:**
- **TypeScript MCP Server**: Full MCP server with 7 tools (`codex-agent/src/`)
- **Go CLI**: Standalone binary (`codex-agent/cmd/implement-ticket/`)

**Features:**
- **MCP Tool**: `implement_ticket` spawns Claude subagent for autonomous implementation
- **Workspace Isolation**: Creates git worktrees per ticket (e.g., `feature/PROJ-001`)
- **Auto-Implementation**: Claude reads existing code patterns, implements feature, writes tests
- **Ticket Updates**: Auto-updates ticket state to "In Review" with implementation summary
- **Keycloak Auth**: Automatic OAuth2 authentication with token refresh

### 4. Webhook Integration
- **Event Types**: ticket.created, ticket.updated, ticket.deleted, ticket.state_changed
- **n8n Compatible**: Send ticket events to n8n workflows
- **Signatures**: Optional HMAC-SHA256 signing
- **Test Send**: Verify webhook configuration

---

## Skills

### Managing Tickets

**MCP Server**: `codex-agent` (TypeScript/Node.js)

**MCP Tools**:
- `get_ticket` - Get ticket details with comments
- `list_tickets` - List tickets in project with filtering
- `search_tickets` - Search tickets across projects
- `add_comment` - Add comment to ticket
- `update_ticket_state` - Update ticket state (by ID or name)
- `get_project_workflow` - Get available workflow states

**Usage via MCP Client** (e.g., Claude Code):
```json
{
  "tool": "get_ticket",
  "input": {"ticketId": "550e8400-..."}
}
```

**See**: `/skills/managing-tickets/SKILL.md`

### Implementing Features

**MCP Tool**: `implement_ticket`

**Usage via MCP**:
```json
{
  "tool": "implement_ticket",
  "input": {
    "ticketId": "PROJ-001",
    "workspaceRoot": "~/worktrees",
    "repoPath": "/path/to/repo"
  }
}
```

**Usage via Go CLI**:
```bash
./bin/implement-ticket --ticket PROJ-001 --repo /path/to/repo
```

**Usage via SSH (n8n)**:
```bash
ssh user@host "bash .../scripts/skill-cli.sh \"implement PROJ-001\""
```

**Implementation Process** (8 steps):
1. Resolves ticket ID/key and fetches full details
2. Fetches all ticket comments for context
3. Creates isolated git worktree: `feature/PROJ-001`
4. Generates comprehensive implementation prompt
5. Spawns Claude subagent with ticket context
6. Subagent implements feature, writes tests, commits changes
7. Updates ticket state to "In Review"
8. Adds implementation summary comment with files, test results, commit SHA

**Requirements**:
- Ticket must be type "feature" (not bug)
- `claude` CLI installed and authenticated (`claude auth`)
- `KEYCLOAK_USERNAME` and `KEYCLOAK_PASSWORD` for API access
- Valid git repository

**See**: `/skills/implementing-features/SKILL.md`

---

## Configuration

### Environment Variables

#### Backend (docker-compose.yml)
```yaml
KEYCLOAK_BASE_URL=http://keycloak:8080
KEYCLOAK_REALM=ticketing
KEYCLOAK_CLIENT_ID=myclient
KEYCLOAK_USERNAME=AdminUser
KEYCLOAK_PASSWORD=admin123
TICKETING_API_BASE_URL=http://ticketing-api:8080
```

#### Frontend (.env)
```
VITE_API_BASE=http://localhost:8080
VITE_PROJECT_ID=<project-uuid>
```

#### Codex Agent MCP Server
```bash
# Authentication (Required)
KEYCLOAK_USERNAME=AdminUser
KEYCLOAK_PASSWORD=admin123

# Keycloak Configuration (Optional - defaults shown)
KEYCLOAK_BASE_URL=http://localhost:8081
KEYCLOAK_REALM=ticketing
KEYCLOAK_CLIENT_ID=myclient

# API Configuration
TICKETING_API_BASE_URL=http://localhost:8080

# Workspace Configuration
WORKSPACE_ROOT=~/worktrees
REPO_PATH=~/projects/coding-agent-workflow/ticketing-system
SUBAGENT_TIMEOUT=30m
AUTO_UPDATE_STATE=true

# Note: Requires 'claude' CLI to be installed and authenticated
# Run: claude auth
```

### Keycloak Setup

Users in `keycloak/realm.json`:
- **AdminUser** / `admin123` (role: admin)
- **Codex/Claude** / `nichtWeiterWichtig` (role: admin)
- **NormalUser** / `user123` (role: user)

---

## User Management

### Creating Groups

1. **Settings → Groups**
2. Click "➕ Create new group"
3. Enter name and optional description
4. Click "Create group"

### Adding Users to Groups

1. **Select a group** from dropdown
2. **Search for user** (fuzzy search):
   - Type: `ich`, `admin`, `normal`, `llm`, etc.
   - Press Enter or click Search
   - Results sorted by relevance
3. **Click "Add"** on search result
4. **Confirm** user appears in members table

### Assigning Groups to Projects

1. **Right column → Project access**
2. **Select group** from dropdown
3. **Choose role**: Admin, Contributor, or Viewer
4. **Click "Add to project"**
5. Verify in project access table

### User Search (Fuzzy Matching)

**Algorithm**: All characters in query must appear in text in order

| Query | Matches | Doesn't Match |
|-------|---------|---------------|
| `ich` | ich@ich.ich, lich | normal@... |
| `adm` | AdminUser, admin123 | Normal |
| `agent` | llm@agent-workflow.com | ich@... |

---

## n8n Integration

### Setting Up SSH Connection

1. **n8n → Add credential → SSH**
2. **Host**: Your server address
3. **Username**: SSH user
4. **Authentication**: SSH key or password
5. **Test connection**

### Creating Skill Workflows

#### Example: Implement Feature on Ticket Created

```
Webhook (ticket.created)
    ↓
Extract ticket ID
    ↓
SSH Node: bash .../skill-cli.sh "implement {{ticket_id}}"
    ↓
Log result
    ↓
Send notification
```

#### Example: Search Users

```
HTTP (GET /users?q=ich)
    ↓
Parse results
    ↓
For each user: Add to group {{group_id}}
```

### SSH Command Format

```bash
ssh user@host "bash ~/projects/coding-agent-workflow/codex-agent/scripts/skill-cli.sh \"<skill-command>\""
```

**Skill Commands**:
- `implement PROJ-001`
- `get ticket PROJ-001`
- `search tickets for authentication`
- `list projects`

---

## Troubleshooting

### Users Not Found in Search

**Issue**: Search returns no results for realm.json users

**Solution**: Users are synced when they log in. To make all users immediately searchable:
1. Have each user log in once (e.g., `AdminUser`, `Codex/Claude`, `NormalUser`)
2. Or implement user sync on backend startup (future feature)

### Search Not Working

**Check**:
- [ ] At least one user has logged in
- [ ] Search uses fuzzy matching (e.g., `ich` for `ich@ich.ich`)
- [ ] Press Enter or click Search button
- [ ] Check browser console for errors

### Feature Implementation Fails

**Check**:
- [ ] Ticket type is "feature" (not bug)
- [ ] Claude CLI installed: `which claude && claude --version`
- [ ] Claude CLI authenticated: `claude auth` (check status with `claude auth status`)
- [ ] Git repository valid: `cd /repo && git status`
- [ ] Subagent timeout not exceeded (30 minutes default)

### Webhook Not Firing

**Check**:
- [ ] Webhook URL is correct and reachable
- [ ] Webhook is enabled
- [ ] Check API logs: `docker-compose logs ticketing-api`
- [ ] Test webhook: Settings → Webhooks → Send test

### SSH Connection Errors

**Check**:
- [ ] Server reachable: `ssh -v user@host echo "test"`
- [ ] Credentials correct in n8n
- [ ] Script path exists: `ssh user@host ls -la ~/projects/.../scripts/`
- [ ] Bash installed on server: `ssh user@host which bash`

---

## Development

### Building from Source

```bash
# Backend
cd ticketing-system/backend
go build -o api ./cmd/api

# Frontend
cd ticketing-system/frontend
npm install && npm run build

# Codex Agent MCP Server (TypeScript)
cd codex-agent
npm install && npm run build

# Codex Agent CLI (Go)
cd /path/to/project
go build -o bin/implement-ticket ./codex-agent/cmd/implement-ticket/

# Or use the build script
bash codex-agent/scripts/build-go.sh
```

### Local Testing

**TypeScript MCP Server:**
```bash
# Start in dev mode
cd codex-agent
KEYCLOAK_USERNAME=AdminUser \
KEYCLOAK_PASSWORD=admin123 \
npm run dev

# Or use MCP Inspector
npx @modelcontextprotocol/inspector node dist/index.js
```

**Go CLI:**
```bash
# Ensure claude CLI is authenticated first
claude auth

# Test ticket implementation
KEYCLOAK_USERNAME=AdminUser \
KEYCLOAK_PASSWORD=admin123 \
./bin/implement-ticket --ticket PROJ-001
```

---

## API Endpoints

### Tickets
- `GET /projects/:id/tickets` - List tickets
- `POST /projects/:id/tickets` - Create ticket
- `PATCH /tickets/:id` - Update ticket
- `DELETE /tickets/:id` - Delete ticket

### Groups
- `GET /groups` - List groups
- `POST /groups` - Create group
- `GET /groups/:id/members` - List members
- `POST /groups/:id/members` - Add member

### Users
- `GET /users?q=search` - Search users (fuzzy)
- `POST /auth/login` - Authenticate

### Webhooks
- `GET /projects/:id/webhooks` - List webhooks
- `POST /projects/:id/webhooks` - Create webhook
- `POST /projects/:id/webhooks/:id/test` - Send test

---

## Support & Documentation

- **Backend API**: See `ticketing-system/backend/openapi.yaml`
- **Codex Agent MCP Server**: See `codex-agent/README.md`
- **MCP Tools Documentation**: See `codex-agent/src/tools/`
- **Go CLI Implementation**: See `codex-agent/cmd/implement-ticket/main.go`
- **Skills Documentation**:
  - Managing Tickets: `skills/managing-tickets/SKILL.md`
  - Implementing Features: `skills/implementing-features/SKILL.md`
  - Git Workspace Bootstrap: `skills/git-workspace-bootstrap/SKILL.md`
- **Frontend**: See `ticketing-system/frontend/README.md`

---

## License

MIT
