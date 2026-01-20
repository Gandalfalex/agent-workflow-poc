#!/bin/bash
TICKET_UUID=$1
PROJECT_UUID=$2

if [ -z "$TICKET_UUID" ] || [ -z "$PROJECT_UUID" ]; then
  echo "Usage: $0 <ticket-uuid> <project-uuid>"
  exit 1
fi

cd /home/gandalfalex/projects/agent-workflow-poc

# Get ticket details via API to find the ticket key
TICKET_KEY=$(node -e "
const fetch = require('node-fetch');
const api = 'http://localhost:8080/api';

(async () => {
  try {
    const resp = await fetch(\`\${api}/tickets/\${process.argv[1]}\`, {
      headers: { 'Authorization': \`Bearer \${process.env.AUTH_TOKEN}\` }
    });
    const ticket = await resp.json();
    console.log(ticket.key);
  } catch(e) {
    console.error('Error fetching ticket:', e.message);
    process.exit(1);
  }
})();
" "$TICKET_UUID" 2>/dev/null)

if [ -z "$TICKET_KEY" ]; then
  echo "Error: Could not fetch ticket key for UUID: $TICKET_UUID"
  exit 1
fi

echo "Implementing ticket: $TICKET_KEY ($TICKET_UUID)"
echo "========================================"

# Run the skill with the ticket key
bash codex-agent/scripts/skill-cli.sh "implement $TICKET_KEY"

# Update ticket state to "In Review" when done
echo "========================================"
echo "Implementation complete!"
