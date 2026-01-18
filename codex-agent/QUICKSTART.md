# Quick Start Guide

## Local Development

```bash
# Install dependencies
npm install

# Build TypeScript
npm run build

# Run in development mode (with hot reload)
npm run dev

# Or run production build
npm start
```

## Docker Deployment

```bash
# Start just the MCP server with its dependencies
docker-compose up codex-agent

# Or start everything including ticketing system
docker-compose up

# View logs
docker-compose logs -f codex-agent

# Stop
docker-compose down
```

## Testing with MCP Inspector

1. Start the server:
   ```bash
   npm run dev
   ```

2. In another terminal, use MCP Inspector to connect to `stdio`

3. Test tools:
   - List projects first to get project IDs
   - Use those IDs to list tickets
   - Get a specific ticket
   - Try updating state with friendly name like "Done"

## Available Tools

### List all projects
```
Tool: list_projects (no parameters)
```

### Get a ticket
```
Tool: get_ticket
Input: { "ticketId": "uuid-here" }
```

### List tickets in a project
```
Tool: list_tickets
Input: {
  "projectId": "uuid-here",
  "query": "optional search text",
  "limit": 10
}
```

### Search tickets
```
Tool: search_tickets
Input: {
  "query": "search text",
  "projectId": "uuid-here (optional)"
}
```

### Add comment
```
Tool: add_comment
Input: {
  "ticketId": "uuid-here",
  "message": "Your comment here"
}
```

### Get workflow states
```
Tool: get_project_workflow
Input: { "projectId": "uuid-here" }
```

### Update ticket state
Option 1 - By state UUID:
```
Tool: update_ticket_state
Input: {
  "ticketId": "uuid-here",
  "stateId": "state-uuid-here"
}
```

Option 2 - By friendly state name:
```
Tool: update_ticket_state
Input: {
  "ticketId": "uuid-here",
  "stateName": "Done"
}
```

## Environment Variables

All configured in `docker-compose.yml`, but can be overridden:

- `KEYCLOAK_BASE_URL` - Keycloak server URL
- `KEYCLOAK_REALM` - Keycloak realm (default: ticketing)
- `KEYCLOAK_CLIENT_ID` - Keycloak client (default: myclient)
- `KEYCLOAK_USERNAME` - Username for authentication
- `KEYCLOAK_PASSWORD` - Password for authentication
- `TICKETING_API_BASE_URL` - Ticketing API URL

## Troubleshooting

**Build fails**
- Ensure Node.js 20+ is installed
- Run `npm clean-install` to clear cache

**Authentication errors**
- Check Keycloak is running: `docker-compose logs keycloak`
- Verify credentials in docker-compose.yml
- Check KEYCLOAK_BASE_URL points to running instance

**API errors**
- Check ticketing-api is running: `docker-compose logs ticketing-api`
- Verify ticketing database is initialized
- Check TICKETING_API_BASE_URL is correct

**MCP Server won't connect**
- Make sure server is running with `npm run dev`
- Check no other process is using stdio
- Verify no compile errors in TypeScript output
