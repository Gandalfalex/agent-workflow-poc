# Claude Code Skills Guide

Complete guide for using the autonomous feature implementation system through Claude Code native skills.

## Overview

Three Claude Code skills are available for ticket automation:

1. **implementing-features** - Autonomously implement features
2. **managing-tickets** - Retrieve and update ticket information
3. **git-workspace-bootstrap** - Create isolated git workspaces

## Quick Start

### From Claude Code

Simply ask Claude Code to perform actions:

```
Implement ticket PROJ-001
```

Claude automatically:
1. Recognizes the `implementing-features` skill
2. Loads full skill documentation
3. Executes the implementation
4. Returns results

## Skills

### 1. Implementing Features

**Description:** Autonomously implements features from tickets by creating isolated workspaces, running subagents, and updating ticket state.

**Usage:**
```
Implement ticket PROJ-001
Implement ticket PROJ-001 using /custom/repo
Implement feature PROJ-123
```

**What it does:**
- ✓ Fetches ticket details and comments from ticketing system
- ✓ Creates isolated git worktree (feature/PROJ-001)
- ✓ Spawns Claude subagent to implement feature
- ✓ Runs tests to verify implementation
- ✓ Updates ticket state to "In Review"
- ✓ Adds implementation summary comment
- ✓ Returns detailed results

**Output:**
```json
{
  "success": true,
  "ticketKey": "PROJ-001",
  "workspacePath": "/workspaces/PROJ-001",
  "branch": "feature/PROJ-001",
  "summary": "Implemented feature...",
  "filesChanged": ["file1.ts", "file2.ts"],
  "testsRun": true,
  "testsPassed": true,
  "commitSha": "abc123def456"
}
```

**Example Workflow:**
1. Create feature ticket in ticketing system
2. Ask Claude: "Implement ticket PROJ-001"
3. Claude implements feature autonomously
4. Review code in feature branch
5. Merge to main when ready

---

### 2. Managing Tickets

**Description:** Retrieves ticket information, updates ticket states, adds comments, and searches tickets.

**Usage:**
```
Get ticket PROJ-001
List tickets in project [project-id]
Search tickets for authentication
Add comment to PROJ-001: Implementation complete
Update PROJ-001 state to In Review
```

**Available Operations:**
- ✓ Get ticket details with all comments
- ✓ List tickets with filtering
- ✓ Search across projects
- ✓ Add comments/updates
- ✓ Update ticket states

**Examples:**

```
Get ticket PROJ-001
```
Returns full ticket details including all comments.

```
Search tickets for authentication
```
Finds all tickets mentioning authentication across projects.

```
Add comment to PROJ-001: Implementation complete, ready for review
```
Adds status update visible to entire team.

```
Update PROJ-001 state to In Review
```
Moves ticket through workflow.

---

### 3. Git Workspace Bootstrap

**Description:** Create a new git workspace and branch inside an existing repo using git worktrees.

**Usage:**
```
Bootstrap a workspace for PROJ-001
Create worktree for feature/PROJ-001
Start workspace for idea/new-feature
```

**What it does:**
- ✓ Creates isolated git worktree
- ✓ Creates new feature branch
- ✓ Provides clean working directory
- ✓ No effect on main checkout

---

## Complete Workflows

### Workflow 1: Feature Implementation

```
User: "Implement ticket PROJ-001"

Claude Code:
  1. Scans /skills directory
  2. Finds "implementing-features" skill
  3. Loads SKILL.md documentation
  4. Executes implementation:
     - Fetch ticket details
     - Create worktree
     - Spawn subagent
     - Run tests
     - Update ticket
  5. Returns results to user

Result: Feature implemented and ready for review!
```

### Workflow 2: Monitor Multiple Tickets

```
User: "Show me high-priority tickets and their status"

Claude Code:
  1. Recognizes "managing-tickets" skill
  2. Executes:
     - List high-priority tickets
     - Get details for each
     - Show current state
  3. Presents summary

User: "Implement the top 3"

Claude Code:
  1. Identifies implementing-features skill
  2. Implements each ticket in sequence
  3. Returns results for all 3
```

### Workflow 3: Batch Implementation

```
User: "Get all unstarted feature tickets"

Claude Code:
  Uses managing-tickets skill:
  - Searches for "status:todo type:feature"
  - Returns list of 5 tickets

User: "Implement the first three"

Claude Code:
  Uses implementing-features skill:
  - Implements PROJ-001
  - Implements PROJ-002
  - Implements PROJ-003
  - Returns results for each
```

---

## Integration with n8n

The skills are also accessible through n8n via SSH:

```bash
ssh user@host "cd ~/codex-agent && bash scripts/skill-cli.sh \"implement PROJ-001\""
```

### n8n SSH Node Setup

1. Add SSH node to workflow
2. Configure SSH credentials
3. Set command:
```bash
cd ~/codex-agent && bash scripts/skill-cli.sh "implement {{ $json.ticketId }}"
```

---

## Configuration

### Environment Variables

Set before using skills:

```bash
export REPO_PATH=/path/to/repo              # Repository for worktrees
export WORKSPACE_ROOT=/path/to/workspaces   # Where to create worktrees
export SUBAGENT_TIMEOUT=1800000             # Timeout in ms (30 min default)
```

### For Docker/Container Deployment

Set in docker-compose.yml or container environment:

```yaml
environment:
  - REPO_PATH=/repo
  - WORKSPACE_ROOT=/workspaces
  - SUBAGENT_TIMEOUT=1800000
```

---

## File Structure

```
/skills
├── implementing-features/
│   └── SKILL.md                    # Autonomous implementation
├── managing-tickets/
│   └── SKILL.md                    # Ticket management
└── git-workspace-bootstrap/
    └── SKILL.md                    # Workspace creation

/codex-agent
├── src/
│   ├── tools/
│   │   ├── implementation.ts       # Feature implementation tool
│   │   ├── tickets.ts             # Ticket retrieval tool
│   │   ├── comments.ts            # Comment tool
│   │   ├── workflow.ts            # Workflow state tool
│   │   ├── auth.ts                # Keycloak auth
│   │   └── api-client.ts          # Ticketing API client
│   └── index.ts                   # MCP server
│
└── scripts/
    ├── skill-cli.sh               # Natural language parser
    ├── run-skill.js               # Node.js invoker
    ├── implement                  # Bash wrapper
    ├── create-worktree.sh         # Git worktree creation
    └── implement-ticket.sh        # Alternative wrapper
```

---

## How Skills Execute

### Behind the Scenes

1. **Claude Code** detects relevant skill from user prompt
2. **Skill metadata** (name + description) is in system prompt
3. **Full SKILL.md** is loaded when relevant
4. **Claude understands** the instructions
5. **Execution flow:**
   - Calls scripts/skill-cli.sh with natural language prompt
   - skill-cli.sh parses prompt and identifies skill
   - Invokes scripts/run-skill.js with skill name and parameters
   - run-skill.js spawns MCP server process
   - MCP server executes the actual tool
   - Returns JSON result
6. **Result** is presented to user

### Technology Stack

- **Claude Code Skills** - Skill documentation and discovery
- **MCP (Model Context Protocol)** - Tool execution layer
- **Keycloak** - Authentication
- **Git worktrees** - Isolated workspaces
- **Anthropic SDK** - Subagent spawning
- **Node.js** - Runtime environment
- **Ticketing System API** - Data backend

---

## Examples

### Example 1: Implement Feature

```
Claude Code User: "Implement ticket PROJ-123"

Claude Code processes:
  1. Recognize: "implementing-features" skill
  2. Load: Full SKILL.md
  3. Execute: Implementation process
  4. Return: Detailed results

Output:
  ✓ Worktree created: feature/PROJ-123
  ✓ Subagent implemented feature
  ✓ Tests passed
  ✓ Ticket updated to "In Review"
  ✓ Comment added with summary
```

### Example 2: Search and Update

```
Claude Code User: "Find all authentication tickets and add a comment that implementation starts tomorrow"

Claude Code processes:
  1. Use "managing-tickets" skill to search
  2. Find tickets matching "authentication"
  3. For each: Add comment "Implementation starts tomorrow"
  4. Return count and summary

Output:
  ✓ Found 3 tickets
  ✓ Added comment to PROJ-001
  ✓ Added comment to PROJ-050
  ✓ Added comment to PROJ-075
```

### Example 3: Batch Implementation

```
Claude Code User: "Get high-priority feature tickets and implement the first two"

Claude Code processes:
  1. Search for high-priority features
  2. Implement PROJ-200
  3. Implement PROJ-201
  4. Return both results

Output:
  ✓ PROJ-200: Implemented, tests passed
  ✓ PROJ-201: Implemented, tests passed
```

---

## Troubleshooting

### Skill Not Triggered

If Claude doesn't recognize the skill:
- Be explicit: "Use the implementing-features skill..."
- Include key terms: "implement", "feature", "ticket"
- Check skill description matches your intent

### Execution Failed

Check:
1. Ticket exists in ticketing system: `Get ticket PROJ-001`
2. Repository path configured: `echo $REPO_PATH`
3. TypeScript compiled: `ls -la codex-agent/dist/`
4. SSH access (if remote): `ssh user@host "echo ok"`

### Results Not As Expected

1. Review SKILL.md documentation for usage details
2. Check ticket details: `Get ticket PROJ-XXX`
3. Verify ticketing system is running
4. Check logs: `/tmp/skill-cli.log`

---

## Advanced

### Custom Repository

Implement feature in different repository:

```
Implement ticket PROJ-001 using /path/to/other/repo
```

### Monitor Progress

```
cd ~/worktrees/PROJ-001
git log --oneline -10
git diff main
npm test
```

### Manual Completion

If subagent times out:
```bash
cd ~/worktrees/PROJ-001
# Make manual changes
git add .
git commit -m "PROJ-001: Manual implementation"

# Then update ticket manually
Get ticket PROJ-001
Update PROJ-001 state to In Review
Add comment to PROJ-001: Manual implementation completed
```

---

## Best Practices

1. **Start Simple** - Test with small features first
2. **Monitor First** - Get ticket info before implementing
3. **Review Code** - Always review generated code before merging
4. **Use Prompts** - Provide clear, specific prompts to Claude
5. **Check Logs** - Monitor logs during execution
6. **Test Locally** - Test implementations locally before merging

---

## Support

### Documentation

- [Implementing Features](/skills/implementing-features/SKILL.md)
- [Managing Tickets](/skills/managing-tickets/SKILL.md)
- [Git Workspace Bootstrap](/skills/git-workspace-bootstrap/SKILL.md)
- [MCP Tools Reference](/codex-agent/IMPLEMENTATION_SKILL.md)
- [n8n Integration](/N8N_SKILL_INTEGRATION.md)

### Resources

- Ticketing System: http://localhost:5173
- API: http://localhost:8080
- Keycloak: http://localhost:8081

### Issues

For problems:
1. Check `/tmp/skill-cli.log` for detailed logs
2. Verify configuration and environment variables
3. Test ticket retrieval: `Get ticket PROJ-001`
4. Review SKILL.md documentation
