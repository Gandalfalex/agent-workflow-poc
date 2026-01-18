# MCP Server Startup Guide

## Quick Start

The MCP server has been successfully built and is ready to run.

### Prerequisites

1. **Install dependencies** (one-time):
   ```bash
   cd ~/projects/coding-agent-workflow/codex-agent
   npm install
   ```

2. **Build the TypeScript** (one-time or after changes):
   ```bash
   npm run build
   ```

3. **Set environment variables**:
   ```bash
   export KEYCLOAK_USERNAME=AdminUser
   export KEYCLOAK_PASSWORD=admin123
   export KEYCLOAK_BASE_URL=http://localhost:8081
   export KEYCLOAK_REALM=ticketing
   export KEYCLOAK_CLIENT_ID=myclient
   export TICKETING_API_BASE_URL=http://localhost:8080
   export WORKSPACE_ROOT=~/worktrees
   export REPO_PATH=~/projects/coding-agent-workflow/ticketing-system
   ```

### Starting the Server

**Option 1: With environment variables (Local Development)**
```bash
cd ~/projects/coding-agent-workflow/codex-agent
KEYCLOAK_USERNAME=AdminUser \
KEYCLOAK_PASSWORD=admin123 \
KEYCLOAK_BASE_URL=http://localhost:8081 \
KEYCLOAK_REALM=ticketing \
KEYCLOAK_CLIENT_ID=myclient \
TICKETING_API_BASE_URL=http://localhost:8080 \
node dist/index.js
```

**Note:** Subagent spawning via `implement_ticket` uses the existing Claude/Codex CLI running on the server. No API key needed.

**Option 2: With Docker Compose (Full Stack)**
```bash
cd ~/projects/coding-agent-workflow
docker-compose up -d
```

This starts all services including:
- PostgreSQL (for n8n and ticketing)
- Keycloak (auth server)
- Ticketing API
- n8n (workflow automation)
- Codex Agent MCP server

**Option 3: Background with pm2 (Production)**
```bash
npm install -g pm2
pm2 start "node dist/index.js" --name "ticketing-mcp" --env-file .env
```

## What the Server Does

The MCP server exposes 8 tools:

1. **get_ticket** - Retrieve a specific ticket
2. **list_tickets** - List all tickets in a project
3. **search_tickets** - Search tickets by keyword
4. **add_comment** - Add a comment to a ticket
5. **update_ticket_state** - Change ticket status
6. **get_project_workflow** - Get available states for a project
7. **implement_ticket** - Spawn subagent to implement a feature
8. **list_projects** - Get all projects

## Testing the Server

Once running, the server listens on stdin/stdout for JSON-RPC requests.

### Test with MCP Inspector

```bash
# In another terminal
npx @modelcontextprotocol/inspector node ~/projects/coding-agent-workflow/codex-agent/dist/index.js
```

This launches a web UI at `http://localhost:5173` where you can:
- List all available tools
- Call tools interactively
- See request/response JSON
- Test the full workflow

### Test with curl (via SSH node)

```bash
cd ~/projects/coding-agent-workflow/codex-agent

# Test a simple ticket lookup
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list_projects",
    "arguments": {}
  }
}' | node dist/index.js
```

## Connection Details

### For Claude Code (Local MCP)

Claude Code can connect to the running server:

```
Settings → Add MCP Server:
Server: command
Command: node /path/to/dist/index.js
Environment:
  KEYCLOAK_USERNAME=AdminUser
  KEYCLOAK_PASSWORD=admin123
  ... (etc)
```

### For n8n Workflows (SSH Node)

Use the SSH node to invoke skills:

```bash
ssh user@host "cd ~/projects/coding-agent-workflow/codex-agent && bash scripts/skill-cli.sh \"implement PROJ-001\""
```

### For Docker Deployment

The docker-compose.yml includes the codex-agent service pre-configured:

```yaml
codex-agent:
  build: ./codex-agent
  environment:
    - KEYCLOAK_USERNAME=AdminUser
    - KEYCLOAK_PASSWORD=admin123
    - KEYCLOAK_BASE_URL=http://keycloak:8080
    - KEYCLOAK_REALM=ticketing
    - KEYCLOAK_CLIENT_ID=myclient
    - TICKETING_API_BASE_URL=http://ticketing-api:8080
```

## Troubleshooting

### Build Errors
```bash
# Clean and rebuild
rm -rf dist
npm run build
```

### Missing Dependencies
```bash
# Reinstall everything
rm -rf node_modules package-lock.json
npm install
npm run build
```

### Environment Variable Errors
```
Error: Missing KEYCLOAK_USERNAME and/or KEYCLOAK_PASSWORD

# Solution: Export environment variables before running
export KEYCLOAK_USERNAME=AdminUser
export KEYCLOAK_PASSWORD=admin123
node dist/index.js
```

### API Connection Errors
```
Error: Failed to authenticate with Keycloak

# Verify Keycloak is running
curl http://localhost:8081/auth/realms/ticketing

# Check credentials in realm.json
cat ../ticketing-system/keycloak/realm.json
```

### Subagent Timeout
```
Error: Subagent timeout after 30 minutes

# Increase timeout in environment
export SUBAGENT_TIMEOUT=3600000  # 60 minutes
```

## Development Workflow

1. **Make code changes** in `src/`
2. **Rebuild**: `npm run build`
3. **Test**: Use MCP Inspector or CLI
4. **Deploy**: Commit and push, or restart docker-compose

## How Subagent Spawning Works

The `implement_ticket` tool spawns a subagent by calling the existing Claude/Codex CLI on the server:

```
MCP Server (codex-agent)
    ↓
implement_ticket tool is called
    ↓
spawns: claude --read-file .claude-prompt.txt
    ↓
Uses existing Claude/Codex session on server
    ↓
Returns implementation results
```

**No API keys needed** - the Claude CLI authenticates using the existing session on the server.

### Prerequisites for Subagent Spawning

1. Claude/Codex CLI must be installed on the server
2. User must be authenticated with Claude
3. `claude` command must be available in PATH

Check if Claude CLI is available:
```bash
which claude
claude --version
```

## Next Steps

1. ✅ Build the TypeScript: `npm run build`
2. ✅ Start the server: `npm start` (no API key required)
3. Test with MCP Inspector
4. Connect from Claude Code
5. Deploy with Docker Compose
6. Invoke from n8n workflows via SSH
7. Verify Claude CLI is available for subagent spawning

For detailed usage, see:
- `README.md` - Architecture and API
- `IMPLEMENTATION_SKILL.md` - Feature implementation details
- `QUICKSTART.md` - Quick reference
- `/CLAUDE_CODE_SKILLS_GUIDE.md` - Skill usage
