# MCP & Go Implementation Update - Changelog

## Summary

Updated the Codex Agent implementation to fix the Go CLI and ensure consistency between TypeScript and Go implementations. Updated all skills documentation to reference the new MCP server structure.

## Changes Made

### 1. Go Implementation (`codex-agent/cmd/implement-ticket/main.go`)

**Complete rewrite to match TypeScript functionality:**

- ✅ **Keycloak OAuth2 Authentication**
  - Implemented password grant flow
  - Added token refresh with 30-second buffer
  - Automatic token caching and reuse

- ✅ **Full API Client**
  - `GetTicket`, `ListProjects`, `ListTickets`, `GetComments`
  - `GetWorkflow`, `UpdateTicket`, `AddComment`
  - Bearer token authentication on all requests

- ✅ **Ticket Key Resolution**
  - Resolves both UUIDs and ticket keys (e.g., "PROJ-001")
  - Searches all projects when ticket key is provided

- ✅ **Comment Fetching**
  - Retrieves all ticket comments for full context
  - Passes to subagent for informed implementation

- ✅ **Enhanced Prompt Generation**
  - Includes ticket details, comments, story context
  - Rich implementation instructions
  - Clear output format expectations

- ✅ **Improved Subagent JSON Parsing**
  - Handles Claude's JSON output format
  - Extracts embedded JSON from text responses
  - Fallback to raw output as summary

- ✅ **Comprehensive Help**
  - Documents all environment variables
  - Usage examples

### 2. Go Module Structure

- ✅ Moved `go.mod` to project root (`/`)
- ✅ Removed nested `codex-agent/cmd/implement-ticket/go.mod`
- ✅ Updated `codex-agent/scripts/build-go.sh` to build from root
- ✅ Verified compilation: `go build -o bin/implement-ticket ./codex-agent/cmd/implement-ticket/`

### 3. TypeScript Implementation (`codex-agent/src/tools/implementation.ts`)

- ✅ **Duration String Support**
  - Added `parseTimeout()` function
  - Supports both milliseconds (backward compatible) and duration strings
  - Valid formats: `1800000` (ms), `30m`, `1h`, `90s`, `500ms`
  - Defaults to `30m` if invalid format

### 4. Docker Compose (`docker-compose.yml`)

**Environment Variables Updated:**

```yaml
# Before:
- SUBAGENT_TIMEOUT=1800000  # milliseconds

# After:
- SUBAGENT_TIMEOUT=30m      # duration string
```

- ✅ Added comments to group environment variables
- ✅ Reorganized for better readability
- ✅ Both TypeScript and Go implementations now support this format

### 5. Skills Documentation

**`skills/managing-tickets/SKILL.md`:**
- ✅ Added MCP server metadata (`mcp_server: codex-agent`)
- ✅ Listed all 6 MCP tools in frontmatter
- ✅ Updated all examples to show MCP tool usage with JSON input/output
- ✅ Added MCP server connection details
- ✅ Documented Keycloak authentication configuration
- ✅ Updated integration examples

**`skills/implementing-features/SKILL.md`:**
- ✅ Added MCP tool metadata (`mcp_tool: implement_ticket`)
- ✅ Documented both TypeScript and Go implementations
- ✅ Added 8-step workflow description
- ✅ Added JSON input examples for MCP tool
- ✅ Added Go CLI usage examples
- ✅ Updated configuration with all environment variables
- ✅ Enhanced troubleshooting section
- ✅ Updated references to source code

**`skills/git-workspace-bootstrap/SKILL.md`:**
- ✅ No changes needed (standalone bash script)

### 6. Main README (`README.md`)

- ✅ Updated architecture diagram ("Codex Agent MCP Server")
- ✅ Clarified 7 MCP tools (corrected from 8)
- ✅ Added TypeScript and Go implementation details
- ✅ Updated skills section with MCP tool documentation
- ✅ Enhanced configuration section
- ✅ Added build instructions for both implementations
- ✅ Added testing instructions
- ✅ Updated documentation references

## Environment Variable Reference

### Required

```bash
# Keycloak Authentication
KEYCLOAK_USERNAME=AdminUser
KEYCLOAK_PASSWORD=admin123

# Claude CLI must be installed and authenticated
# No API key needed - uses existing session
claude auth
```

### Optional (with defaults)

```bash
# Keycloak Configuration
KEYCLOAK_BASE_URL=http://localhost:8081
KEYCLOAK_REALM=ticketing
KEYCLOAK_CLIENT_ID=myclient

# API Configuration
TICKETING_API_BASE_URL=http://localhost:8080

# Workspace Configuration
WORKSPACE_ROOT=~/worktrees
REPO_PATH=.
SUBAGENT_TIMEOUT=30m          # Supports: 30m, 1h, 1800000 (ms), etc.
AUTO_UPDATE_STATE=true
```

## Duration Format Support

The `SUBAGENT_TIMEOUT` environment variable now supports both formats:

| Format | Example | Description |
|--------|---------|-------------|
| Milliseconds | `1800000` | Backward compatible |
| Duration String | `30m` | Human-readable (Go-style) |
| | `1h` | 1 hour |
| | `90s` | 90 seconds |
| | `500ms` | 500 milliseconds |

Both TypeScript and Go implementations parse these formats identically.

## Testing

### Build & Test TypeScript

```bash
cd codex-agent
npm install
npm run build

# Test MCP server
KEYCLOAK_USERNAME=AdminUser \
KEYCLOAK_PASSWORD=admin123 \
npm start
```

### Build & Test Go CLI

```bash
# Build
cd /path/to/project
go build -o bin/implement-ticket ./codex-agent/cmd/implement-ticket/

# Or use script
bash codex-agent/scripts/build-go.sh

# Authenticate Claude CLI
claude auth

# Test
KEYCLOAK_USERNAME=AdminUser \
KEYCLOAK_PASSWORD=admin123 \
./bin/implement-ticket --ticket PROJ-001
```

### Docker Compose

```bash
# Ensure Claude CLI is authenticated on the host
# (if mounting credentials) or inside the container
claude auth

# Start services
docker-compose up -d

# View logs
docker-compose logs -f codex-agent
```

## Breaking Changes

None - all changes are backward compatible:
- TypeScript still accepts milliseconds for `SUBAGENT_TIMEOUT`
- Go implementation accepts both formats
- Docker Compose uses duration strings but TypeScript parses them correctly

## Migration Guide

If you were using `SUBAGENT_TIMEOUT` with milliseconds:

### No action needed
Both implementations support the old format. However, for consistency with Go-style durations:

```bash
# Old (still works)
export SUBAGENT_TIMEOUT=1800000

# New (recommended)
export SUBAGENT_TIMEOUT=30m
```

## Files Modified

```
codex-agent/cmd/implement-ticket/main.go       # Complete rewrite
codex-agent/src/tools/implementation.ts        # Added parseTimeout()
codex-agent/scripts/build-go.sh                # Updated build path
docker-compose.yml                             # Updated env vars
skills/managing-tickets/SKILL.md               # MCP tool documentation
skills/implementing-features/SKILL.md          # MCP tool documentation
README.md                                      # Updated references
go.mod                                         # Moved to root
```

## Files Added

```
CHANGELOG-MCP-UPDATE.md                        # This file
```

## Files Removed

```
codex-agent/cmd/implement-ticket/go.mod        # Consolidated to root
```

## Next Steps

1. ✅ All code is compiled and working
2. ✅ Documentation is updated
3. ✅ Both implementations (TypeScript & Go) are feature-complete
4. ⏭️ Test with real tickets
5. ⏭️ Deploy to production

## Verification Checklist

- [x] Go implementation compiles
- [x] TypeScript implementation compiles
- [x] Duration parsing works in both implementations
- [x] Docker Compose configuration updated
- [x] All skills documentation updated
- [x] Main README updated
- [x] Build scripts updated
- [x] No breaking changes introduced

---

## Update: Removed ANTHROPIC_API_KEY Requirement

**Date:** Follow-up fix

### Changes

Both implementations (TypeScript and Go) use the `claude` CLI command instead of direct API calls. This means:

✅ **No API Key Required**
- Removed all `ANTHROPIC_API_KEY` environment variable requirements
- Uses existing authenticated `claude` CLI session
- No additional API costs beyond the Claude Code/CLI subscription

✅ **Prerequisites**
```bash
# Install Claude CLI (if not already installed)
# https://claude.ai/cli

# Authenticate once
claude auth

# Check authentication status
claude auth status
```

✅ **How It Works**

**TypeScript Implementation:**
```typescript
// Uses claude CLI command
const { stdout, stderr } = await exec(
  `cd "${workspacePath}" && claude --read-file .claude-prompt.txt --output-format text`,
  { timeout: timeout }
);
```

**Go Implementation:**
```go
// Uses claude CLI command
cmd := exec.CommandContext(ctx, "claude", "--dangerously-skip-permissions", 
  "--output-format", "json", "--print", prompt)
```

### Files Updated

- ✅ `docker-compose.yml` - Removed ANTHROPIC_API_KEY env var
- ✅ `README.md` - Updated all references to use `claude auth`
- ✅ `skills/implementing-features/SKILL.md` - Updated configuration section
- ✅ `codex-agent/IMPLEMENTATION_SKILL.md` - Updated prerequisites and troubleshooting
- ✅ `CHANGELOG-MCP-UPDATE.md` - Updated environment variable reference

### Benefits

1. **Simpler Setup**: No need to manage API keys
2. **Single Authentication**: Use existing Claude CLI session
3. **No Extra Costs**: Included in Claude Code/CLI subscription
4. **Better Security**: No API keys to store or rotate

### Migration

If you previously set `ANTHROPIC_API_KEY`:

```bash
# Old (no longer needed)
export ANTHROPIC_API_KEY=sk-ant-...

# New (just authenticate once)
claude auth
```

The environment variable is simply ignored if set - no breaking changes.
