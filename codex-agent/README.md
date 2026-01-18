# Ticketing System MCP Server

An MCP (Model Context Protocol) server that provides tools for interacting with the ticketing system API. Authenticate with Keycloak and perform operations like reading tickets, adding comments, and changing ticket states.

## Features

- **Ticket Management**: Get, list, and search tickets across projects
- **Comments**: Add comments to tickets
- **Workflow**: Change ticket states and view available states
- **Authentication**: Automatic Keycloak OAuth2 authentication with token refresh
- **Type-Safe**: Full TypeScript support with Zod validation

## Available Tools

### `get_ticket`
Get a specific ticket with all its details including comments.

**Input:**
- `ticketId` (string): The ticket ID (UUID)

**Output:** Full ticket object with comments

### `list_tickets`
List tickets in a project with optional filtering.

**Input:**
- `projectId` (string): The project ID (UUID)
- `stateId` (string, optional): Filter by state ID
- `assigneeId` (string, optional): Filter by assignee ID
- `query` (string, optional): Search query
- `limit` (number, optional): Results limit (default: 50)
- `offset` (number, optional): Results offset (default: 0)

**Output:** `{ items: Ticket[], total: number }`

### `search_tickets`
Search for tickets across projects or within a specific project.

**Input:**
- `query` (string): Search query
- `projectId` (string, optional): Project ID to search in

**Output:** `{ items: Ticket[], total: number }`

### `add_comment`
Add a comment to a ticket.

**Input:**
- `ticketId` (string): The ticket ID (UUID)
- `message` (string): Comment message

**Output:** Comment object

### `update_ticket_state`
Update a ticket's state/status. Can use either `stateId` (direct) or `stateName` (friendly).

**Input:** Either:
- `ticketId` (string): The ticket ID (UUID)
- `stateId` (string): The state ID (UUID)

Or:
- `ticketId` (string): The ticket ID (UUID)
- `stateName` (string): The state name (e.g., 'Done', 'In Review')

**Output:** Updated ticket object

### `get_project_workflow`
Get all workflow states available in a project.

**Input:**
- `projectId` (string): The project ID (UUID)

**Output:** Array of WorkflowState objects

## Setup

### Environment Variables

Required:
- `KEYCLOAK_USERNAME`: Keycloak username
- `KEYCLOAK_PASSWORD`: Keycloak password

Optional:
- `KEYCLOAK_BASE_URL`: Keycloak server URL (default: `http://keycloak:8080`)
- `KEYCLOAK_REALM`: Keycloak realm name (default: `ticketing`)
- `KEYCLOAK_CLIENT_ID`: Keycloak client ID (default: `myclient`)
- `TICKETING_API_BASE_URL`: Ticketing API base URL (default: `http://ticketing-api:8080`)

### Installation

```bash
npm install
```

### Development

```bash
npm run dev
```

### Build

```bash
npm run build
```

### Production

```bash
npm start
```

## Docker

Build and run with Docker:

```bash
docker build -t ticketing-mcp .
docker run --env KEYCLOAK_USERNAME=AdminUser --env KEYCLOAK_PASSWORD=admin123 ticketing-mcp
```

Or use with docker-compose:

```bash
docker-compose up codex-agent
```

## Testing with MCP Inspector

1. Start the server:
```bash
npm run dev
```

2. Use the MCP Inspector tool to connect to `stdio` and test the tools

## Architecture

- **Authentication** (`src/auth.ts`): Keycloak OAuth2 client with automatic token refresh
- **API Client** (`src/api-client.ts`): Typed HTTP client for the ticketing system API
- **Tools** (`src/tools/`):
  - `tickets.ts`: Ticket retrieval and search operations
  - `comments.ts`: Comment operations
  - `workflow.ts`: Workflow state management
- **Server** (`src/index.ts`): MCP server implementation

## Error Handling

The server returns errors with descriptive messages including:
- Authentication failures
- API errors with status codes
- Validation errors from Zod schemas
- Missing state names with available options
