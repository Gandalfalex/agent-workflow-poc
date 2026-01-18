---
name: implementing-features
description: Autonomously implements features from tickets by creating isolated workspaces, running subagents, and updating ticket state. Use when you need to automatically implement a feature described in a ticket with full code generation, testing, and integration.
---

# Implementing Features

This skill automates the entire feature implementation process from ticket creation through code review readiness.

## Overview

When you provide a ticket ID or key, this skill will:

1. **Fetch ticket details** from the ticketing system
2. **Create an isolated git workspace** (worktree) for the feature
3. **Spawn a Claude subagent** to implement the feature
4. **Run tests** to verify implementation
5. **Update ticket state** to "In Review"
6. **Add implementation summary** as a ticket comment

## When to Use

Use this skill when you have:
- A feature ticket in the ticketing system
- A git repository set up with the codebase
- Need for autonomous implementation
- Want code ready for review within 5-30 minutes

## Usage

### Basic Implementation

Provide the ticket ID or key:

```
Implement ticket PROJ-001
```

or

```
Run the implementing-features skill on ticket: PROJ-001
```

### With Custom Repository

If your code is in a different repository:

```
Implement ticket PROJ-001 using /path/to/custom/repo
```

## How It Works

### Step 1: Ticket Resolution
The skill resolves the ticket by:
- Accepting both ticket keys (PROJ-001) and UUIDs
- Fetching full ticket details including title, description, priority
- Retrieving all comments for context
- Validating the ticket is a "feature" type

### Step 2: Workspace Creation
- Creates isolated git worktree: `feature/PROJ-001`
- Locations: `~/worktrees/PROJ-001` (configurable)
- Branch is isolated from main branch
- Safe for parallel implementations

### Step 3: Context Generation
- Generates implementation prompt with ticket details
- Includes all comments and discussion
- Provides project patterns and conventions
- Defines success criteria

### Step 4: Subagent Implementation
A Claude subagent then:
- Reads existing code to understand patterns
- Implements feature according to specification
- Writes tests for new functionality
- Runs test suite to verify
- Commits changes with clear commit messages

### Step 5: Ticket Update
- Changes ticket state to "In Review"
- Adds comment with implementation summary
- Includes files changed, test results, commit SHA
- Provides next steps for code review

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

The skill uses environment variables (can be set before running):

```bash
export REPO_PATH=/path/to/repo              # Default: current directory
export WORKSPACE_ROOT=/path/to/workspaces   # Default: ~/worktrees
export SUBAGENT_TIMEOUT=1800000             # Timeout in ms (30 min default)
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

### Batch Implementation

Implement multiple tickets:

```bash
for ticket in PROJ-001 PROJ-002 PROJ-003; do
  implement-features $ticket
done
```

### Custom Prompting

Modify implementation behavior by providing additional context:

```
Implement ticket PROJ-001 using React hooks only, skip database changes
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

### Build Failed
```bash
cd ~/codex-agent && npm run build
```

### TypeScript Not Compiled
```bash
ls -la dist/
npm run build  # if dist doesn't exist
```

### SSH Connection Issues
```bash
ssh -v user@host "echo 'Connected'"
```

### Check Logs
```bash
tail -f /tmp/skill-cli.log
```

## References

- [Ticketing System API](/codex-agent/README.md) - Available API endpoints
- [MCP Tools](/codex-agent/IMPLEMENTATION_SKILL.md) - Underlying tools
- [SSH Integration](/SSH_SKILL_INVOCATION.md) - SSH command details
- [n8n Integration](/N8N_SKILL_INTEGRATION.md) - n8n workflow setup

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
