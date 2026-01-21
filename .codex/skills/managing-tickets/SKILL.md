---
name: managing-tickets
description: Retrieves ticket information, updates ticket states, adds comments, and searches tickets. Use when you need to interact with the ticketing system to get details, search for tickets, add updates, or change ticket status.
mcp_server: codex-agent
mcp_tools:
  - get_ticket
  - list_tickets
  - search_tickets
  - add_comment
  - update_ticket_state
  - get_project_workflow
---

# Managing Tickets

This skill provides access to the ticketing system for retrieving, searching, and updating ticket information using the **Codex Agent MCP Server**.

## Overview

This skill enables you to:

1. **Get ticket details** - Retrieve full ticket information including comments (`get_ticket`)
2. **List tickets** - View all tickets in a project with filtering (`list_tickets`)
3. **Search tickets** - Find tickets across projects by keyword (`search_tickets`)
4. **Add comments** - Add notes and updates to tickets (`add_comment`)
5. **Update state** - Change ticket status through workflow states (`update_ticket_state`)
6. **Get workflow states** - View available states for a project (`get_project_workflow`)

## When to Use

Use this skill when you need to:
- Look up ticket information
- Find related tickets by searching
- Add status updates or notes to tickets
- Move tickets through workflow states
- Get project information and available states

## MCP Server Connection

This skill uses the **Codex Agent MCP Server** (`ticketing-mcp`) which provides authenticated access to the ticketing system API via Keycloak OAuth2.

**Server:** `codex-agent` (TypeScript/Node.js)  
**Location:** `/codex-agent/`  
**Tools:** 7 MCP tools for ticket management

## Usage

### Get Ticket Details

Retrieve full information about a specific ticket using the `get_ticket` tool:

**MCP Tool:** `get_ticket`

**Input:**
```json
{
  "ticketId": "550e8400-e29b-41d4-a716-446655440000"
}
```

Returns:
- Ticket title, description, type, priority
- Current state and assigned user
- All comments and discussion
- Linked story if applicable

### List Tickets in Project

View all tickets in a specific project using the `list_tickets` tool:

**MCP Tool:** `list_tickets`

**Input:**
```json
{
  "projectId": "550e8400-e29b-41d4-a716-446655440000",
  "stateId": "optional-state-id",
  "assigneeId": "optional-assignee-id",
  "query": "optional search query",
  "limit": 50,
  "offset": 0
}
```

**Output:**
```json
{
  "items": [/* array of tickets */],
  "total": 42
}
```

### Search Tickets

Find tickets across projects using the `search_tickets` tool:

**MCP Tool:** `search_tickets`

**Input:**
```json
{
  "query": "authentication",
  "projectId": "optional-project-id-to-limit-search"
}
```

**Output:**
```json
{
  "items": [/* array of matching tickets */],
  "total": 15
}
```

### Add Comment

Add a note or update to a ticket using the `add_comment` tool:

**MCP Tool:** `add_comment`

**Input:**
```json
{
  "ticketId": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Implementation complete, ready for review"
}
```

### Update Ticket State

Move a ticket to a different state in the workflow using the `update_ticket_state` tool:

**MCP Tool:** `update_ticket_state`

**Input (by state ID):**
```json
{
  "ticketId": "550e8400-e29b-41d4-a716-446655440000",
  "stateId": "state-uuid-here"
}
```

**Input (by state name - friendly):**
```json
{
  "ticketId": "550e8400-e29b-41d4-a716-446655440000",
  "stateName": "In Review"
}
```

Supported state names (project-specific):
- Todo
- In Progress
- In Review
- Done
- Blocked

## Examples

### Example 1: Check Implementation Status

```
Get ticket PROJ-001
```

Returns full ticket details including all comments showing implementation progress.

### Example 2: Find Related Tickets

```
Search tickets for authentication
```

Returns all tickets mentioning authentication across all projects.

### Example 3: Add Status Update

```
Add comment to PROJ-001: Fixed merge conflicts, tests passing, ready for review
```

Adds comment to ticket for team visibility.

### Example 4: Move Through Workflow

```
Get ticket PROJ-001
[Review implementation details]
Update PROJ-001 state to In Review
Add comment to PROJ-001: Implementation reviewed, merged to develop
```

## Ticket Information

Each ticket contains:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "key": "PROJ-001",
  "title": "Add user authentication",
  "description": "Implement OAuth2 authentication...",
  "type": "feature",
  "priority": "high",
  "status": "In Progress",
  "assignee": "John Doe",
  "comments": [
    {
      "author": "Jane",
      "text": "Started implementation",
      "timestamp": "2024-01-18T10:30:00Z"
    }
  ]
}
```

## Workflow States

Each project has custom workflow states. View available states using the `get_project_workflow` tool:

**MCP Tool:** `get_project_workflow`

**Input:**
```json
{
  "projectId": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Output:**
```json
{
  "states": [
    {
      "id": "state-uuid",
      "projectId": "project-uuid",
      "name": "Todo",
      "order": 0,
      "isDefault": true,
      "isClosed": false
    },
    {
      "id": "state-uuid-2",
      "name": "In Progress",
      "order": 1,
      "isDefault": false,
      "isClosed": false
    },
    {
      "id": "state-uuid-3",
      "name": "Done",
      "order": 2,
      "isDefault": false,
      "isClosed": true
    }
  ]
}
```

Typical workflow states:
- **Todo** - Newly created tickets
- **In Progress** - Being worked on
- **In Review** - Waiting for code review
- **Done** - Completed and merged
- **Blocked** - Waiting for dependencies (project-specific)

## Search Operators

Search queries support:
- Keywords: `authentication`, `bug`, `dashboard`
- Multiple terms: `authentication AND api`
- Quoted phrases: `"user profile"`
- Project filter: `project:PROJ tickets`

## Integration with Other Skills

### With Implementing Features

1. Use `get_ticket` to retrieve full ticket details and requirements
2. Use `implement_ticket` MCP tool (from implementing-features skill) to auto-implement
3. Use `add_comment` to add status updates
4. Use `update_ticket_state` to move to "In Review"

### With n8n Workflows

```bash
ssh user@host "bash scripts/skill-cli.sh \"add comment to PROJ-001: Automated implementation started\""
```

### Batch Operations

```bash
# Get all high-priority tickets
List high-priority tickets

# For each ticket, add comment
for ticket in $results; do
  Add comment to $ticket: Ready for implementation
done
```

## Output Formats

### Ticket Details

```json
{
  "success": true,
  "ticket": {
    "key": "PROJ-001",
    "title": "Feature name",
    "description": "...",
    "status": "In Progress",
    "comments": [...]
  }
}
```

### Ticket List

```json
{
  "success": true,
  "tickets": [
    {"key": "PROJ-001", "title": "..."},
    {"key": "PROJ-002", "title": "..."}
  ],
  "total": 2
}
```

### Search Results

```json
{
  "success": true,
  "results": [
    {"key": "PROJ-001", "title": "..."},
    {"key": "OTHER-123", "title": "..."}
  ],
  "total": 2
}
```

## Error Handling

Common errors and solutions:

- **Ticket not found** - Check ticket key format (PROJ-001 or full UUID)
- **Project not found** - Verify project ID is correct
- **State not found** - Use exact state name or get available states first
- **No permission** - Verify access to project

## Advanced Usage

### Monitor Multiple Tickets

```
List tickets in project [id]
[For each high-priority ticket]
Get ticket [key]
Add comment: Status review - ready for implementation?
```

### Aggregate Status

```
Search for "in progress"
[Get count and details]
Add summary to dashboard
```

### Update Batch

```
Search for feature tickets
[For each not yet in Review]
Get details
Update to Done
Add comment: Verified complete
```

## Configuration

The Codex Agent MCP Server requires the following environment variables:

**Required:**
- `KEYCLOAK_USERNAME` - Keycloak username
- `KEYCLOAK_PASSWORD` - Keycloak password

**Optional:**
- `KEYCLOAK_BASE_URL` - Keycloak server URL (default: `http://localhost:8081`)
- `KEYCLOAK_REALM` - Keycloak realm (default: `ticketing`)
- `KEYCLOAK_CLIENT_ID` - Keycloak client ID (default: `myclient`)
- `TICKETING_API_BASE_URL` - Ticketing API URL (default: `http://localhost:8080`)

The server handles automatic Keycloak OAuth2 authentication with token refresh.

## References

- [Implementing Features](/skills/implementing-features/SKILL.md) - Auto-implement tickets
- [Ticketing System API](/codex-agent/README.md) - API details
- [n8n Integration](/N8N_SKILL_INTEGRATION.md) - Workflow automation

## Related Skills

- [Implementing Features](/skills/implementing-features/SKILL.md)

## Limitations

- **Read-only search** - Cannot delete tickets
- **State names** - Must use exact workflow state names
- **Access control** - Can only access tickets you have permission for
- **Comment format** - Plain text only (no rich formatting)

## Support

For issues:
1. Verify ticket ID/key format
2. Check project access permissions
3. Confirm workflow state names are correct
4. Review ticketing system status
