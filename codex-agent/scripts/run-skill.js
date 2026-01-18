#!/usr/bin/env node

/**
 * Feature Implementation Skill - Direct CLI Invocation
 * Usage: node run-skill.js <skill-name> <ticket-id> [options]
 *
 * Examples:
 *   node run-skill.js implement_ticket PROJ-001
 *   node run-skill.js implement_ticket PROJ-001 --repo /path/to/repo
 *   node run-skill.js implement_ticket PROJ-001 --workspace /workspaces
 */

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const skillName = process.argv[2];
const ticketId = process.argv[3];

// Parse options
const options = {};
for (let i = 4; i < process.argv.length; i += 2) {
  const key = process.argv[i]?.replace(/^--/, '');
  const value = process.argv[i + 1];
  if (key && value) {
    options[key] = value;
  }
}

// Validate inputs
if (!skillName || !ticketId) {
  console.error('Usage: node run-skill.js <skill-name> <ticket-id> [options]');
  console.error('');
  console.error('Available skills:');
  console.error('  - implement_ticket');
  console.error('  - get_ticket');
  console.error('  - list_tickets');
  console.error('  - add_comment');
  console.error('  - update_ticket_state');
  console.error('');
  console.error('Examples:');
  console.error('  node run-skill.js implement_ticket PROJ-001');
  console.error('  node run-skill.js implement_ticket PROJ-001 --repo /path/to/repo');
  console.error('  node run-skill.js get_ticket PROJ-001');
  process.exit(1);
}

// Setup environment
const projectDir = path.dirname(path.dirname(__filename));
const distDir = path.join(projectDir, 'dist');

if (!fs.existsSync(distDir)) {
  console.error('Error: TypeScript not compiled. Run: npm run build');
  process.exit(1);
}

// Start MCP server as subprocess
const mcp = spawn('node', [path.join(distDir, 'index.js')], {
  cwd: projectDir,
  stdio: ['pipe', 'pipe', 'inherit'],
  env: {
    ...process.env,
    REPO_PATH: options.repo || process.env.REPO_PATH || '.',
    WORKSPACE_ROOT: options.workspace || process.env.WORKSPACE_ROOT || `${process.env.HOME}/worktrees`
  }
});

let resultReceived = false;
let timeoutHandle;

// Handle response from MCP server
mcp.stdout.on('data', (data) => {
  const lines = data.toString().split('\n');

  for (const line of lines) {
    if (!line.trim()) continue;

    try {
      const msg = JSON.parse(line);

      // Check if this is a tool result
      if (msg.result && msg.result.content) {
        resultReceived = true;
        clearTimeout(timeoutHandle);

        const content = msg.result.content[0];
        if (content.type === 'text') {
          console.log(content.text);
        } else {
          console.log(JSON.stringify(msg.result, null, 2));
        }

        mcp.kill();
        process.exit(0);
      }

      // Check for errors
      if (msg.error) {
        resultReceived = true;
        clearTimeout(timeoutHandle);
        console.error('Error:', msg.error.message);
        mcp.kill();
        process.exit(1);
      }
    } catch (e) {
      // Ignore parsing errors - might be partial data
    }
  }
});

// Send tool call request after server starts
setTimeout(() => {
  const request = {
    jsonrpc: '2.0',
    id: 1,
    method: 'tools/call',
    params: {
      name: skillName,
      arguments: {
        ticketId: ticketId,
        ...(options.repo && { repoPath: options.repo }),
        ...(options.workspace && { workspaceRoot: options.workspace })
      }
    }
  };

  mcp.stdin.write(JSON.stringify(request) + '\n');
}, 500);

// Timeout after 30 minutes for long implementations
timeoutHandle = setTimeout(() => {
  if (!resultReceived) {
    console.error('Timeout: Implementation did not complete within 30 minutes');
    mcp.kill();
    process.exit(1);
  }
}, 30 * 60 * 1000);

// Handle errors
mcp.on('error', (err) => {
  console.error('Failed to start MCP server:', err.message);
  process.exit(1);
});

mcp.on('exit', (code) => {
  if (code !== 0 && !resultReceived) {
    console.error('MCP server exited with code:', code);
    process.exit(code);
  }
});

// Handle interruption
process.on('SIGINT', () => {
  console.log('\nImplementation interrupted');
  mcp.kill();
  process.exit(130);
});
