#!/bin/bash

# Skill CLI - Parse Claude-style prompts and execute MCP skills
# Usage: claude -p "run Implementation skill on ticket: PROJ-001"
#        claude -p "get ticket PROJ-001"
#        claude -p "implement PROJ-001"

# Get the full prompt
PROMPT="$*"

# Debug logging
LOG_FILE="/tmp/skill-cli.log"
echo "[$(date)] Received prompt: $PROMPT" >> "$LOG_FILE"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Build TypeScript if needed
if [ ! -d "$PROJECT_DIR/dist" ]; then
  echo "Building TypeScript..."
  cd "$PROJECT_DIR"
  npm run build
fi

# Parse the prompt to extract skill and ticket info
# Patterns to match:
# - "run Implementation skill on ticket: PROJ-001"
# - "implement PROJ-001"
# - "get ticket PROJ-001"
# - "add comment to PROJ-001"
# - "update state of PROJ-001"

SKILL_NAME=""
TICKET_ID=""

# Check for "Implementation skill" or "implement_ticket"
if [[ "$PROMPT" =~ [Ii]mplementation.*skill|[Ii]mplement.*(ticket|on).* ]]; then
  SKILL_NAME="implement_ticket"

  # Extract ticket ID (PROJ-### or UUID pattern)
  if [[ "$PROMPT" =~ ([A-Z]{2,}-[0-9]{3}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}) ]]; then
    TICKET_ID="${BASH_REMATCH[1]}"
  fi
fi

# Check for "get ticket"
if [[ "$PROMPT" =~ [Gg]et.*(ticket|info) ]]; then
  SKILL_NAME="get_ticket"

  if [[ "$PROMPT" =~ ([A-Z]{2,}-[0-9]{3}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}) ]]; then
    TICKET_ID="${BASH_REMATCH[1]}"
  fi
fi

# Check for "list tickets"
if [[ "$PROMPT" =~ [Ll]ist.*(tickets|issues) ]]; then
  SKILL_NAME="list_tickets"

  # Extract project ID
  if [[ "$PROMPT" =~ [Pp]roject.*(in|:).*([0-9a-f-]+) ]]; then
    TICKET_ID="${BASH_REMATCH[2]}"
  fi
fi

# Check for "search tickets"
if [[ "$PROMPT" =~ [Ss]earch.*(tickets|for) ]]; then
  SKILL_NAME="search_tickets"

  # Extract search query
  if [[ "$PROMPT" =~ [Ff]or[:\"]?\s+([^\"]+) ]] || [[ "$PROMPT" =~ search[:\"]?\s+([^\"]+) ]]; then
    TICKET_ID="${BASH_REMATCH[1]}"
  fi
fi

# Check for "add comment"
if [[ "$PROMPT" =~ [Aa]dd.*comment ]]; then
  SKILL_NAME="add_comment"

  if [[ "$PROMPT" =~ ([A-Z]{2,}-[0-9]{3}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}) ]]; then
    TICKET_ID="${BASH_REMATCH[1]}"
  fi
fi

# Check for "update state" or "change state"
if [[ "$PROMPT" =~ [Cc]hange.*state|[Uu]pdate.*state ]]; then
  SKILL_NAME="update_ticket_state"

  if [[ "$PROMPT" =~ ([A-Z]{2,}-[0-9]{3}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}) ]]; then
    TICKET_ID="${BASH_REMATCH[1]}"
  fi
fi

# Check for "list projects"
if [[ "$PROMPT" =~ [Ll]ist.*projects ]]; then
  SKILL_NAME="list_projects"
fi

# If no skill matched, try to be helpful
if [ -z "$SKILL_NAME" ]; then
  cat << 'HELP'
I can help you with these skills:

FEATURE IMPLEMENTATION (Main):
  - "run Implementation skill on ticket: PROJ-001"
  - "implement PROJ-001"

TICKET OPERATIONS:
  - "get ticket PROJ-001"
  - "add comment to PROJ-001"
  - "update state of PROJ-001"
  - "list tickets in project UUID"
  - "search tickets for authentication"
  - "list projects"

WORKFLOW:
  - "get workflow states for project UUID"

Examples:
  claude -p "run Implementation skill on ticket: PROJ-001"
  claude -p "get ticket PROJ-001"
  claude -p "search tickets for authentication"
HELP
  exit 1
fi

echo "[$(date)] Executing skill: $SKILL_NAME with ticket: $TICKET_ID" >> "$LOG_FILE"

# Execute the skill
if [ "$SKILL_NAME" == "implement_ticket" ] && [ -n "$TICKET_ID" ]; then
  node "$SCRIPT_DIR/run-skill.js" "$SKILL_NAME" "$TICKET_ID"
elif [ "$SKILL_NAME" == "list_projects" ]; then
  node "$SCRIPT_DIR/run-skill.js" "$SKILL_NAME"
elif [ -n "$TICKET_ID" ]; then
  node "$SCRIPT_DIR/run-skill.js" "$SKILL_NAME" "$TICKET_ID"
else
  echo "Error: Could not parse ticket ID or project ID from prompt: $PROMPT"
  exit 1
fi
