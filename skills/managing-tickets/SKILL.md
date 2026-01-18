---
name: managing-tickets
description: Retrieves ticket information, updates ticket states, adds comments, and searches tickets. Use when you need to interact with the ticketing system to get details, search for tickets, add updates, or change ticket status.
---

# Managing Tickets

This skill provides access to the ticketing system for retrieving, searching, and updating ticket information.

## Overview

This skill enables you to:

1. **Get ticket details** - Retrieve full ticket information including comments
2. **List tickets** - View all tickets in a project with filtering
3. **Search tickets** - Find tickets across projects by keyword
4. **Add comments** - Add notes and updates to tickets
5. **Update state** - Change ticket status through workflow states

## When to Use

Use this skill when you need to:
- Look up ticket information
- Find related tickets by searching
- Add status updates or notes to tickets
- Move tickets through workflow states
- Get project information and available states

## Usage

### Get Ticket Details

Retrieve full information about a specific ticket:

```
Get ticket PROJ-001
```

Returns:
- Ticket title, description, type, priority
- Current state and assigned user
- All comments and discussion
- Linked story if applicable

### List Tickets in Project

View all tickets in a specific project:

```
List tickets in project [project-id]
```

Optional filters:
```
List tickets in project [id] assigned to [user]
List high-priority tickets in project [id]
Search for authentication in project [id]
```

### Search Tickets

Find tickets across projects:

```
Search tickets for authentication
Search for bug reports
Find all tickets about API integration
```

### Add Comment

Add a note or update to a ticket:

```
Add comment to PROJ-001: Implementation complete, ready for review
```

### Update Ticket State

Move a ticket to a different state in the workflow:

```
Update PROJ-001 state to In Review
Change PROJ-001 status to Done
Move PROJ-001 to Review
```

Supported state names:
- Todo
- In Progress
- In Review
- Done
- Blocked (project-specific)

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

## Project Information

Get available projects:

```
List projects
```

Returns all accessible projects with:
- Project ID and key
- Project name and description
- Available workflow states

## Workflow States

Each project has custom workflow states. View available states:

```
Get workflow states for project [id]
```

Typical states:
- Todo - Newly created tickets
- In Progress - Being worked on
- In Review - Waiting for code review
- Done - Completed and merged
- Blocked - Waiting for dependencies

## Search Operators

Search queries support:
- Keywords: `authentication`, `bug`, `dashboard`
- Multiple terms: `authentication AND api`
- Quoted phrases: `"user profile"`
- Project filter: `project:PROJ tickets`

## Integration with Other Skills

### With Implementing Features

1. Get ticket details to understand requirements
2. Run implementing-features skill to auto-implement
3. Add comment with status update
4. Update state to "In Review"

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

No configuration needed - uses existing ticketing system connection.

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
