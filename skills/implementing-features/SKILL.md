---
name: implementing-features
description: Autonomously implements features from tickets by creating isolated workspaces, running subagents, and updating ticket state. Use when you need to automatically implement a feature described in a ticket with full code generation, testing, and integration.
mcp_server: codex-agent
mcp_tool: implement_ticket
---

# Implementing Features

This skill automates the entire feature implementation process from ticket creation through code review readiness using the **Codex Agent MCP Server**.

## Overview

This skill uses the `implement_ticket` MCP tool from the Codex Agent server. When you provide a ticket ID or key, the tool will:

1. **Fetch ticket details** from the ticketing system (including comments for full context)
2. **Create an isolated git workspace** (worktree) for the feature
3. **Spawn a Claude subagent** to implement the feature
4. **Run tests** to verify implementation
5. **Update ticket state** to "In Review"
6. **Add implementation summary** as a ticket comment

## MCP Server Connection

**Server:** `codex-agent` (TypeScript/Node.js or Go CLI)  
**MCP Tool:** `implement_ticket`  
**Location:** `/codex-agent/`

The tool is available in two implementations:
- **TypeScript/Node.js:** Full MCP server with 7 tools (`src/tools/implementation.ts`)
- **Go CLI:** Standalone binary (`cmd/implement-ticket/main.go`)

## When to Use

Use this skill when you have:
- A feature ticket in the ticketing system
- A git repository set up with the codebase
- Need for autonomous implementation
- Want code ready for review within 5-30 minutes

## Usage

### Using the MCP Tool

**MCP Tool:** `implement_ticket`

**Input:**
```json
{
  "ticketId": "PROJ-001",
  "workspaceRoot": "/Users/you/worktrees",
  "autoCommit": true,
  "autoPush": false,
  "repoPath": "/path/to/repo"
}
```

**Parameters:**
- `ticketId` (required): Ticket UUID or key (e.g., "PROJ-001")
- `workspaceRoot` (optional): Root directory for worktrees (default: `~/worktrees`)
- `autoCommit` (optional): Auto-commit changes (default: `true`)
- `autoPush` (optional): Auto-push to remote (default: `false`)
- `repoPath` (optional): Path to repository for worktree creation

### Using the Go CLI

```bash
# Using ticket key
./bin/implement-ticket --ticket PROJ-001

# Using ticket UUID
./bin/implement-ticket --ticket 550e8400-e29b-41d4-a716-446655440000

# With custom paths
./bin/implement-ticket --ticket PROJ-001 \
  --repo /path/to/repo \
  --workspace /path/to/worktrees
```

## How It Works

The `implement_ticket` MCP tool orchestrates an 8-step workflow:

### Step 1: Ticket Resolution
- Accepts both ticket keys (PROJ-001) and UUIDs
- If ticket key provided, searches all projects to resolve to UUID
- Fetches full ticket details including title, description, priority
- Validates the ticket is a "feature" type (not a bug)

### Step 2: Fetch Comments
- Retrieves all ticket comments for discussion context
- Includes comment author, timestamp, and message
- Provides full context to the subagent

### Step 3: Workspace Creation
- Creates isolated git worktree using `scripts/create-worktree.sh`
- Branch format: `feature/PROJ-001`
- Location: `~/worktrees/PROJ-001` (configurable)
- Branch is isolated from main - safe for parallel implementations

### Step 4: Generate Prompt
- Uses template: `prompts/implement-feature.md`
- Injects ticket context (title, description, comments, story)
- Provides clear implementation guidelines
- Defines success criteria and output format

### Step 5: Spawn Subagent
- Calls `claude` CLI command (uses existing authenticated session)
- Passes workspace path and comprehensive prompt
- Timeout: 30 minutes (configurable via `SUBAGENT_TIMEOUT`)

### Step 6: Subagent Implementation
The Claude subagent autonomously:
- Reads existing code to understand patterns
- Implements feature according to specification
- Writes tests for new functionality
- Runs test suite to verify
- Commits changes with message: `PROJ-001: Feature description`
- Outputs JSON summary

### Step 7: Update Ticket State
- Fetches available workflow states for the project
- Finds "In Review" state (or similar: "review", "in_review")
- Updates ticket state via API (if `AUTO_UPDATE_STATE` not false)

### Step 8: Add Comment
- Formats implementation summary with files changed, test results, commit SHA
- Adds comment to ticket for team visibility
- Includes next steps for code review

## Output

The skill returns detailed information:

```json
{
  "success": true,
  "ticketKey": "PROJ-001",
  "workspacePath": "/workspaces/PROJ-001",
  "branch": "feature/PROJ-001",
  "summary": "Implemented user authentication with OAuth2",
  "filesChanged": ["src/auth/index.ts", "src/auth/oauth.ts", "tests/auth.test.ts"],
  "testsRun": true,
  "testsPassed": true,
  "commitSha": "abc123def456",
  "nextSteps": ["Merge after code review", "Deploy to staging"]
}
```

## Configuration

### Environment Variables

**Keycloak Authentication (Required):**
```bash
export KEYCLOAK_USERNAME=AdminUser
export KEYCLOAK_PASSWORD=admin123
```

**Keycloak Configuration (Optional):**
```bash
export KEYCLOAK_BASE_URL=http://localhost:8081    # Default shown
export KEYCLOAK_REALM=ticketing                   # Default shown
export KEYCLOAK_CLIENT_ID=myclient                # Default shown
```

**API Configuration:**
```bash
export TICKETING_API_BASE_URL=http://localhost:8080  # Default shown
```

**Workspace Configuration:**
```bash
export REPO_PATH=/path/to/repo              # Default: current directory
export WORKSPACE_ROOT=/path/to/worktrees    # Default: ~/worktrees
export SUBAGENT_TIMEOUT=30m                 # Default: 30 minutes
export AUTO_UPDATE_STATE=true               # Default: true
```

**Claude CLI:**
```bash
# The implementation uses the 'claude' CLI command
# Ensure it's installed and authenticated:
claude auth
```

## Example Workflow

### Scenario: Implement "Add Dark Mode Toggle"

**1. Create Ticket**
```
Title: Add dark mode toggle to settings
Type: Feature
Description: Add toggle button to switch between light/dark themes
Priority: Medium
→ Ticket PROJ-123 created
```

**2. Trigger Implementation**
```
Implement ticket PROJ-123
```

**3. Automated Process**
- Worktree created: `feature/PROJ-123`
- Subagent reads existing theme system
- Implements toggle component
- Adds theme CSS files
- Writes component tests
- Runs tests ✅ All pass
- Commits: "PROJ-123: Add dark mode toggle"

**4. Ticket Updated**
- State: "In Review"
- Comment: Implementation summary with files
- Branch: Ready for code review

**5. Developer Reviews**
- Checks out `feature/PROJ-123`
- Reviews code changes
- Runs locally and tests
- Merges to main

## Error Handling

If something goes wrong:

- **Ticket not found** - Check ticket ID format (PROJ-001 or full UUID)
- **Repository not found** - Set REPO_PATH or verify repo exists
- **Workspace creation failed** - Check git repository is valid
- **Subagent timeout** - Feature too complex; split into smaller tasks
- **Tests failed** - Subagent output includes test failures; can retry with modifications

All errors include detailed messages with suggested next steps.

## Advanced Usage

### Batch Implementation (Go CLI)

Implement multiple tickets:

```bash
for ticket in PROJ-001 PROJ-002 PROJ-003; do
  ./bin/implement-ticket --ticket $ticket
done
```

### Using Both Implementations

**TypeScript MCP Server:**
```bash
cd codex-agent
npm install && npm run build
KEYCLOAK_USERNAME=AdminUser npm start
# Use with MCP client (Claude Code, etc.)
```

**Go Standalone CLI:**
```bash
# Build
cd /path/to/project
go build -o bin/implement-ticket ./codex-agent/cmd/implement-ticket/

# Authenticate Claude CLI first
claude auth

# Run
KEYCLOAK_USERNAME=AdminUser \
KEYCLOAK_PASSWORD=admin123 \
./bin/implement-ticket --ticket PROJ-001
```

### Monitoring

Check implementation progress:

```bash
# View worktree
cd ~/worktrees/PROJ-001
git log --oneline -10

# Check test results
npm test

# Review changes
git diff main
```

## Limitations

- **Features only**: Use for new feature implementation, not bug fixes
- **30-minute timeout**: Complex features may need to be split
- **No remote push**: Changes stay local for review first
- **Requires valid ticket**: Must exist in ticketing system
- **Type-specific**: Only works with "feature" ticket type

## Integration

### From n8n Workflow

```bash
ssh user@host "cd ~/codex-agent && bash scripts/skill-cli.sh \"implement PROJ-001\""
```

### From GitHub Actions

```yaml
- name: Implement Feature
  run: ssh deploy@server "implement-features PROJ-001"
```

### From Cron Job

```bash
0 9 * * MON-FRI ssh user@host "implement-features DAILY-001"
```

## Troubleshooting

### TypeScript Build Failed
```bash
cd codex-agent
npm install
npm run build
```

### Go Build Failed
```bash
cd /path/to/project
go mod tidy
go build -o bin/implement-ticket ./codex-agent/cmd/implement-ticket/
```

### Authentication Errors
```bash
# Verify Keycloak is accessible
curl http://localhost:8081/realms/ticketing/.well-known/openid-configuration

# Check credentials
echo $KEYCLOAK_USERNAME
echo $KEYCLOAK_PASSWORD  # Will show if set
```

### Subagent Failures
```bash
# Verify Claude CLI is installed
claude --version

# Check authentication status
claude auth status

# Re-authenticate if needed
claude auth

# Test manual subagent spawn
cd ~/worktrees/PROJ-001
claude --print "Implement a simple test"
```

### Worktree Issues
```bash
# List existing worktrees
git worktree list

# Remove stale worktree
git worktree remove ~/worktrees/PROJ-001

# Cleanup
git worktree prune
```

## References

- [Codex Agent MCP Server](/codex-agent/README.md) - MCP server documentation
- [Ticketing System API](/codex-agent/src/api-client.ts) - API client implementation
- [Implementation Tool Source (TS)](/codex-agent/src/tools/implementation.ts) - TypeScript implementation
- [Implementation CLI Source (Go)](/codex-agent/cmd/implement-ticket/main.go) - Go implementation
- [Subagent Spawning](/codex-agent/src/utils/subagent.ts) - Subagent utility
- [Prompt Template](/codex-agent/prompts/implement-feature.md) - Feature implementation prompt

## Related Skills

- [Getting ticket information](/skills/ticket-management/SKILL.md)
- [Updating ticket states](/skills/ticket-management/SKILL.md)
- [Adding ticket comments](/skills/ticket-management/SKILL.md)

## Contact & Support

For issues or feature requests:
1. Check the logs: `/tmp/skill-cli.log`
2. Verify configuration: `REPO_PATH` and `WORKSPACE_ROOT`
3. Test SSH connectivity to remote server
4. Review ticket details in ticketing system
