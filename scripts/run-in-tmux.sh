#!/bin/bash
set -e

TICKET_ID=$1
PROJECT_ID=$2

if [ -z "$TICKET_ID" ]; then
  echo "Usage: $0 <TICKET_ID> [PROJECT_ID]"
  echo "Example: $0 PROJ-001"
  exit 1
fi

SESSION_NAME="implement-ticket-${TICKET_ID}"
LOG_FILE="/tmp/implement-${TICKET_ID}.log"

echo "[$(date)] Starting implementation for ticket: $TICKET_ID" > "$LOG_FILE"

# Option 1: Use local Go CLI binary (recommended - fastest)
tmux new-session -d -s "$SESSION_NAME" \
  "cd $(dirname $0)/.. && \
   KEYCLOAK_USERNAME=AdminUser KEYCLOAK_PASSWORD=admin123 \
   ./bin/implement-ticket --ticket '$TICKET_ID' 2>&1 | tee -a $LOG_FILE; \
   echo ''; \
   read -p 'Implementation complete. Press Enter to close...'"

# Option 2: Use Docker container (requires container to be running)
# tmux new-session -d -s "$SESSION_NAME" \
#   "docker exec codex-agent /app/bin/implement-ticket --ticket '$TICKET_ID' 2>&1 | tee -a $LOG_FILE; \
#    echo ''; \
#    read -p 'Implementation complete. Press Enter to close...'"

# Option 3: Use TypeScript MCP server via Docker
# tmux new-session -d -s "$SESSION_NAME" \
#   "docker exec codex-agent node /app/scripts/run-skill.js implement_ticket '$TICKET_ID' 2>&1 | tee -a $LOG_FILE; \
#    echo ''; \
#    read -p 'Implementation complete. Press Enter to close...'"

echo "✅ Started implementation in tmux session: $SESSION_NAME"
echo "📋 Ticket: $TICKET_ID"
echo "📄 Log: tail -f $LOG_FILE"
echo "🔗 Attach: tmux attach-session -t $SESSION_NAME"
echo ""
echo "💡 Tip: The implementation uses the 'claude' CLI authenticated in the container"
