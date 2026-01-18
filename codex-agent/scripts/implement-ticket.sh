#!/bin/bash

# Feature Implementation Skill - SSH Entry Point
# Usage: ./implement-ticket.sh PROJ-001 [--repo /path/to/repo] [--workspace /path/to/workspaces]
#
# This script triggers the feature implementation skill by invoking the MCP tool
# via the stdio transport.

set -e

TICKET_ID="${1:-}"
REPO_PATH="${REPO_PATH:-.}"
WORKSPACE_ROOT="${WORKSPACE_ROOT:-.}/worktrees"

# Parse arguments
while [[ $# -gt 1 ]]; do
  case "$2" in
    --repo)
      REPO_PATH="$3"
      shift 2
      ;;
    --workspace)
      WORKSPACE_ROOT="$3"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

# Validate inputs
if [ -z "$TICKET_ID" ]; then
  echo "Error: Ticket ID is required"
  echo "Usage: ./implement-ticket.sh <TICKET_ID> [--repo /path/to/repo] [--workspace /path/to/workspaces]"
  echo "Example: ./implement-ticket.sh PROJ-001"
  echo "Example: ./implement-ticket.sh PROJ-001 --repo /path/to/repo --workspace /workspaces"
  exit 1
fi

# Find the node executable
NODE_BIN=$(which node)
if [ -z "$NODE_BIN" ]; then
  echo "Error: Node.js is required but not installed"
  exit 1
fi

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$PROJECT_DIR/dist"

# Check if compiled code exists
if [ ! -d "$DIST_DIR" ]; then
  echo "Error: TypeScript not compiled. Run: npm run build"
  exit 1
fi

# Create a temporary Node script to invoke the MCP tool
TEMP_INVOKER=$(mktemp)
trap "rm -f $TEMP_INVOKER" EXIT

cat > "$TEMP_INVOKER" << 'INVOKER_EOF'
const readline = require('readline');

// Read from stdin and parse JSONRPC responses
const inputLines = [];

process.stdin.on('data', (data) => {
  const lines = data.toString().split('\n');
  lines.forEach(line => {
    if (line.trim()) {
      inputLines.push(JSON.parse(line));
    }
  });
});

process.stdin.on('end', () => {
  // Process all accumulated responses
  const results = inputLines.filter(msg => msg.result || msg.error);
  if (results.length > 0) {
    const lastResult = results[results.length - 1];
    if (lastResult.result) {
      console.log(lastResult.result.content[0].text);
    } else if (lastResult.error) {
      console.error('Error:', lastResult.error.message);
      process.exit(1);
    }
  }
});

// Send the tool call request
const request = {
  jsonrpc: '2.0',
  id: 1,
  method: 'tools/call',
  params: {
    name: 'implement_ticket',
    arguments: {
      ticketId: process.argv[2],
      repoPath: process.argv[3],
      workspaceRoot: process.argv[4]
    }
  }
};

console.log(JSON.stringify(request));
INVOKER_EOF

# Create Node.js wrapper script that starts the MCP server and calls the tool
TEMP_WRAPPER=$(mktemp)
trap "rm -f $TEMP_WRAPPER" EXIT

cat > "$TEMP_WRAPPER" << 'WRAPPER_EOF'
const { spawn } = require('child_process');
const path = require('path');

const projectDir = process.argv[2];
const ticketId = process.argv[3];
const repoPath = process.argv[4];
const workspaceRoot = process.argv[5];

// Start the MCP server as a subprocess
const mcp = spawn('node', [path.join(projectDir, 'dist', 'index.js')], {
  cwd: projectDir,
  stdio: ['pipe', 'pipe', 'pipe'],
  env: process.env
});

let buffer = '';
let toolResult = null;

// Collect responses from MCP server
mcp.stdout.on('data', (data) => {
  buffer += data.toString();

  // Try to parse complete JSON messages
  const lines = buffer.split('\n');
  buffer = lines[lines.length - 1]; // Keep incomplete line

  for (let i = 0; i < lines.length - 1; i++) {
    const line = lines[i].trim();
    if (line) {
      try {
        const msg = JSON.parse(line);

        // Look for tool result
        if (msg.result && msg.result.content) {
          toolResult = msg.result.content[0].text;
        }
      } catch (e) {
        // Ignore parsing errors
      }
    }
  }
});

mcp.stderr.on('data', (data) => {
  // Log debug info but don't fail
  // console.error('[MCP]', data.toString());
});

// Send request to call the tool
setTimeout(() => {
  const request = {
    jsonrpc: '2.0',
    id: 1,
    method: 'tools/call',
    params: {
      name: 'implement_ticket',
      arguments: {
        ticketId: ticketId,
        repoPath: repoPath || '.',
        workspaceRoot: workspaceRoot || process.env.WORKSPACE_ROOT || `${process.env.HOME}/worktrees`
      }
    }
  };

  mcp.stdin.write(JSON.stringify(request) + '\n');
}, 100);

// Wait for result
setTimeout(() => {
  if (toolResult) {
    console.log(toolResult);
    mcp.kill();
    process.exit(0);
  } else {
    console.error('No result received from MCP server');
    mcp.kill();
    process.exit(1);
  }
}, 60000); // 60 second timeout

mcp.on('error', (err) => {
  console.error('Failed to start MCP server:', err.message);
  process.exit(1);
});

mcp.on('exit', (code) => {
  if (code !== 0 && !toolResult) {
    console.error('MCP server exited with code:', code);
    process.exit(code);
  }
});
WRAPPER_EOF

# Run the wrapper
"$NODE_BIN" "$TEMP_WRAPPER" "$PROJECT_DIR" "$TICKET_ID" "$REPO_PATH" "$WORKSPACE_ROOT"
