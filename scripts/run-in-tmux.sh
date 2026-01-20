#!/bin/bash
TICKET_ID=$1
PROJECT_ID=$2

SESSION_NAME="implement-ticket"
LOG_FILE="/tmp/implement-${TICKET_ID}.log"

echo "[$(date)] Starting implementation for ticket: $TICKET_ID" > "$LOG_FILE"

tmux new-session -d -s "$SESSION_NAME" \
  "docker exec codex-agent node /app/scripts/run-skill.js implement_ticket '$TICKET_ID' 2>&1 | tee -a $LOG_FILE; \
   echo ''; \
   read -p 'Implementation complete. Press Enter to close...'"

echo "Started implementation in tmux session: $SESSION_NAME"
echo "Log: tail -f $LOG_FILE"
echo "Attach: tmux attach-session -t $SESSION_NAME"
