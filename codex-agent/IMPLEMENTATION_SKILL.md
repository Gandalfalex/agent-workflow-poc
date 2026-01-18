# Feature Implementation Skill

## Overview

The `implement_ticket` MCP tool spawns Claude subagents to automatically implement features from your ticketing system. This enables a complete autonomous workflow:

1. **Get ticket details** from the ticketing system
2. **Create isolated git worktree** for the feature branch
3. **Spawn Claude subagent** with full context and guidance
4. **Implement the feature** autonomously using available tools
5. **Run tests** and verify implementation
6. **Update ticket state** to "In Review"
7. **Add comment** with implementation summary

## Quick Start

### Prerequisites

- Running ticketing system (http://ticketing-api:8080)
- Running Keycloak (for authentication)
- `ANTHROPIC_API_KEY` environment variable set
- Git repository available at the configured `REPO_PATH`

### Basic Usage

From Claude Code or any MCP client:

```
Call MCP Tool: implement_ticket
Input:
{
  "ticketId": "PROJ-001"
}
```

Or with options:

```
{
  "ticketId": "PROJ-001",
  "workspaceRoot": "/tmp/workspaces",
  "repoPath": "/path/to/repo",
  "autoCommit": true,
  "autoPush": false
}
```

### Expected Output

```json
{
  "success": true,
  "ticketKey": "PROJ-001",
  "workspacePath": "/tmp/workspaces/PROJ-001",
  "branch": "feature/PROJ-001",
  "summary": "Implemented user profile page with authentication",
  "filesChanged": ["src/pages/Profile.tsx", "src/tests/profile.test.ts"],
  "testsRun": true,
  "testsPassed": true,
  "commitSha": "abc123def456",
  "nextSteps": ["Merge into develop after review"]
}
```

## Input Schema

```typescript
{
  ticketId: string,           // Required: UUID or ticket key (e.g., "PROJ-001")
  workspaceRoot?: string,     // Optional: Root for worktrees (default: ~/worktrees)
  repoPath?: string,          // Optional: Repository path (default: REPO_PATH env var)
  autoCommit?: boolean,       // Optional: Auto-commit changes (default: true)
  autoPush?: boolean          // Optional: Auto-push to remote (default: false)
}
```

## How It Works

### Step-by-Step Flow

1. **Ticket Resolution**
   - Accepts both UUID and ticket key formats (e.g., "PROJ-001")
   - Resolves to full ticket object with all details
   - Validates ticket type is "feature" (not bug)

2. **Fetch Context**
   - Gets ticket title, description, priority
   - Fetches all comments for discussion context
   - Retrieves linked story/epic if applicable

3. **Create Worktree**
   - Runs `create-worktree.sh` script
   - Creates isolated git working directory
   - Branch name: `feature/{TICKET_KEY}` (e.g., `feature/PROJ-001`)
   - Location: `{WORKSPACE_ROOT}/{TICKET_KEY}`

4. **Generate Prompt**
   - Uses `prompts/implement-feature.md` template
   - Injects ticket context into template
   - Provides clear implementation guidelines
   - Defines success criteria

5. **Spawn Subagent**
   - Calls Anthropic API with Claude model
   - Passes workspace path and generated prompt
   - Subagent has full context for implementation
   - 30-minute timeout for long-running tasks

6. **Subagent Execution**
   - Claude reads existing code patterns
   - Implements feature according to spec
   - Writes tests for new functionality
   - Runs test suite
   - Commits changes locally
   - Outputs JSON summary when done

7. **Update Ticket**
   - Changes ticket state to "In Review" (if available state exists)
   - Adds formatted comment with implementation details
   - Includes files changed, test results, commit SHA
   - Links to worktree path for manual review

## Architecture

### Core Components

#### 1. Orchestration Tool (`src/tools/implementation.ts`)
- Main entry point for MCP tool
- Coordinates all workflow steps
- Error handling and validation
- Returns structured result

#### 2. Worktree Creation (`scripts/create-worktree.sh`)
- Bash script for git worktree management
- Creates isolated workspace per ticket
- Handles branch creation
- Returns JSON output for programmatic use

#### 3. Subagent Spawner (`src/utils/subagent.ts`)
- Manages Claude subagent lifecycle
- Calls Anthropic SDK with structured prompt
- Parses JSON response
- Handles errors and timeouts

#### 4. Template System (`src/utils/template.ts`)
- Simple `{{variable}}` placeholder replacement
- Injects ticket context into prompts
- Validates all placeholders replaced

#### 5. Prompt Template (`prompts/implement-feature.md`)
- Markdown template with placeholders
- Clear task description
- Implementation guidelines
- Success criteria
- JSON output format specification

### Directory Structure

```
codex-agent/
├── src/
│   ├── tools/
│   │   ├── implementation.ts      # Main MCP tool
│   │   ├── tickets.ts
│   │   ├── comments.ts
│   │   └── workflow.ts
│   ├── utils/
│   │   ├── subagent.ts            # Subagent spawner
│   │   └── template.ts            # Template rendering
│   ├── index.ts                   # MCP server
│   ├── auth.ts
│   └── api-client.ts
├── scripts/
│   └── create-worktree.sh         # Worktree creation
├── prompts/
│   └── implement-feature.md       # Feature implementation template
├── dist/                          # Compiled JavaScript
├── package.json
├── tsconfig.json
├── Dockerfile
└── README.md
```

## Environment Configuration

### Required Variables

```bash
ANTHROPIC_API_KEY=sk-...          # Anthropic API key for Claude
KEYCLOAK_USERNAME=AdminUser       # Keycloak user
KEYCLOAK_PASSWORD=admin123        # Keycloak password
```

### Optional Variables

```bash
WORKSPACE_ROOT=/workspaces        # Where to create worktrees (default: ~/worktrees)
REPO_PATH=/repo                   # Repository for worktree source (default: .)
SUBAGENT_TIMEOUT=1800000          # Timeout in ms (default: 30 min)
AUTO_UPDATE_STATE=true            # Update ticket state after completion
KEYCLOAK_BASE_URL=http://keycloak:8080
KEYCLOAK_REALM=ticketing
KEYCLOAK_CLIENT_ID=myclient
TICKETING_API_BASE_URL=http://ticketing-api:8080
```

### Docker Compose Configuration

```yaml
codex-agent:
  environment:
    - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    - WORKSPACE_ROOT=/workspaces
    - REPO_PATH=/repo
    - SUBAGENT_TIMEOUT=1800000
    - AUTO_UPDATE_STATE=true
  volumes:
    - ./ticketing-system:/repo:ro
    - ./worktrees:/workspaces
```

## Ticket State Transitions

### Before Implementation
- Ticket in any open state (e.g., "Todo", "In Progress")
- Visible to development team

### After Successful Implementation
- Ticket state changed to **"In Review"**
- Comment added with implementation details
- Code available in git branch for review
- Ready for code review and testing

### Failure Handling
- Ticket state unchanged if subagent fails
- Comment added with error details
- Worktree preserved for manual investigation
- Error message includes failure reason

## Example Workflow

### Scenario: Implement "Add Dark Mode Toggle"

1. **Create Ticket**
   ```
   Title: Add dark mode toggle to settings
   Type: Feature
   Description: Add a toggle button in user settings to switch between light and dark themes
   Priority: Medium
   ```
   → Ticket PROJ-123 created

2. **Call implement_ticket**
   ```
   {
     "ticketId": "PROJ-123"
   }
   ```

3. **System Creates Worktree**
   ```
   Branch: feature/PROJ-123
   Path: /workspaces/PROJ-123
   ```

4. **Subagent Implementation**
   - Reads existing theme system
   - Adds toggle component in settings
   - Implements light/dark CSS themes
   - Writes component tests
   - Runs full test suite
   - Commits changes with message: "PROJ-123: Add dark mode toggle"

5. **Ticket Updated**
   - State: Changed to "In Review"
   - Comment: Implementation summary with files changed
   - Branch: `feature/PROJ-123` ready for review

6. **Developer Reviews**
   - Checks out feature branch
   - Reviews code changes
   - Runs locally and tests
   - Merges to main branch

## Error Handling

### Common Errors and Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| "Ticket not found" | Invalid ticket ID | Verify ticket exists and use correct format |
| "Ticket is a bug, not a feature" | Wrong ticket type | Only features are implemented autonomously |
| "Repository path does not exist" | REPO_PATH not configured | Set REPO_PATH env var or pass repoPath param |
| "ANTHROPIC_API_KEY not set" | Missing API key | Set ANTHROPIC_API_KEY environment variable |
| "Worktree creation failed" | Git error | Check git repository is valid and accessible |
| "Subagent failed" | Claude execution error | Check prompt template and workspace permissions |

## Security Considerations

### Workspace Isolation
- Each ticket gets separate worktree
- Changes isolated from main branch
- No cross-contamination between implementations

### Git Safety
- Never pushes to remote by default
- Always creates feature branches
- Requires explicit `autoPush: true` to push
- Safe for review before merge

### API Security
- Anthropic API key stored in environment only
- Keycloak authentication for ticketing API
- Bearer token auto-refresh
- HMAC-signed webhook events (if enabled)

### Code Execution
- Subagent runs in isolated workspace
- Configurable timeout (default 30 min)
- Can kill long-running agents
- Logs captured for debugging

## Testing

### Local Testing

```bash
# Start MCP server
npm run dev

# In another terminal, test with MCP Inspector
# Connect to stdio and call:
# Tool: implement_ticket
# Input: { "ticketId": "PROJ-001" }
```

### Integration Testing

```bash
# 1. Create test ticket in ticketing system
# 2. Note the ticket ID/key

# 3. Call via MCP
{
  "ticketId": "TEST-001"
}

# 4. Monitor:
# - Worktree created: ls -la worktrees/TEST-001
# - Branch created: git branch -a
# - Ticket state changed: Check ticketing UI
# - Comment added: View ticket comments
```

### End-to-End Testing

```bash
# Full workflow test
1. Create feature ticket
2. Call implement_ticket with ticket ID
3. Wait for completion (1-30 minutes depending on complexity)
4. Verify:
   - Worktree exists at correct path
   - Branch created with correct name
   - Code implemented and tests pass
   - Ticket state changed to "In Review"
   - Comment with summary added
   - No uncommitted changes in worktree
```

## Performance Characteristics

### Timing
- **Worktree creation**: < 1 second
- **Prompt generation**: < 1 second
- **Subagent execution**: 1-30 minutes (typical: 5-15 min)
- **Ticket update**: < 1 second
- **Total**: ~5-30 minutes depending on feature complexity

### Resource Usage
- **Disk**: ~100MB per worktree (increases with codebase size)
- **Memory**: ~500MB for MCP server
- **API Calls**: ~10-20 API calls per implementation
- **Claude Tokens**: ~4k-8k tokens per feature

## Limitations and Future Improvements

### Current Limitations
- Only implements features, not bug fixes
- Requires valid Anthropic API key
- 30-minute timeout (configurable but long tasks may fail)
- No multi-ticket dependencies
- Manual merge required after implementation

### Future Enhancements
- [ ] Support bug fix workflows with different prompting
- [ ] Automatic merge for simple features
- [ ] Cross-ticket dependency handling
- [ ] Integration with CI/CD pipelines
- [ ] Performance optimization (caching, parallel execution)
- [ ] Custom prompts per project
- [ ] Integration with code review tools

## Troubleshooting

### Subagent Not Starting

```bash
# Check ANTHROPIC_API_KEY
echo $ANTHROPIC_API_KEY

# Check logs
docker-compose logs codex-agent
```

### Worktree Creation Failed

```bash
# Check git repository
git -C /repo rev-parse --show-toplevel

# Check permissions
ls -la /repo

# Test script manually
bash scripts/create-worktree.sh TEST-001 /repo
```

### Ticket State Not Updated

```bash
# Check AUTO_UPDATE_STATE
echo $AUTO_UPDATE_STATE

# Verify workflow states
# Call: get_project_workflow with projectId
```

## Support and Debugging

### Enable Verbose Logging

The tool logs to stderr by default:

```bash
# View logs
docker-compose logs -f codex-agent

# Look for [Implement] and [Subagent] prefixes
```

### Inspect Worktree

```bash
# Check created worktree
cd worktrees/PROJ-001
git status
git log --oneline

# Review changes
git diff main
```

### Manual Completion

If the subagent fails, you can manually:

```bash
cd worktrees/PROJ-001
# Make your changes
git add .
git commit -m "PROJ-001: Your message"

# Then manually update ticket
# State → "In Review"
# Add comment with details
```

## Related Tools

- `get_ticket` - Fetch ticket details
- `list_tickets` - List project tickets
- `add_comment` - Add comments to tickets
- `update_ticket_state` - Change ticket state manually
- `get_project_workflow` - View available states

## See Also

- [README.md](./README.md) - General MCP server documentation
- [QUICKSTART.md](./QUICKSTART.md) - Quick start guide
- `/codex-agent/prompts/implement-feature.md` - Feature implementation prompt
