# Webhook to SSH Skill Execution Workflow

## Overview

This n8n workflow provides a universal architecture for:
1. **Receiving webhook events** with explicit/raw data
2. **Normalizing data** into skill-compatible formats
3. **Executing skills** via SSH on a remote server
4. **Logging results** back to the ticketing system

```
Webhook → Extract Data → Normalize → SSH Execute → Log Results
```

## Architecture

### Three-Layer Design

```
┌─────────────────────────────────────────────────────────┐
│ LAYER 1: Webhook Input (Explicit/Raw Data)             │
│ - Any external system can POST data                     │
│ - No transformation required at source                  │
│ - Flexible data structure                              │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ LAYER 2: Normalization (Transform to Skill Format)     │
│ - Extract common fields from raw payload                │
│ - Route to appropriate skill handler                    │
│ - Build skill-specific commands                         │
│ - Set environment variables                             │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│ LAYER 3: SSH Execution (Remote Skill Agent)            │
│ - Execute skill via SSH on server                      │
│ - Capture stdout/stderr                                │
│ - Handle errors and timeouts                           │
│ - Return structured results                            │
└─────────────────────────────────────────────────────────┘
```

## Workflow Nodes

### 1. Webhook Trigger
**Type:** Webhook  
**ID:** `webhook_trigger`

Receives incoming HTTP POST requests with event data.

**URL Pattern:**
```
POST http://localhost:5678/webhook/webhook-skill-execution
```

**Accepts any JSON payload**, examples:
```json
{
  "event_type": "feature_implementation",
  "ticket_id": "PROJ-001",
  "data": {
    "workspace_root": "/workspaces",
    "repo_path": "/repo"
  }
}
```

### 2. Extract Webhook Data
**Type:** Assign (Set Variables)  
**ID:** `extract_webhook_data`

Normalizes webhook payload by extracting standard fields:
- `rawPayload` - Complete original payload
- `eventType` - Event type (feature_implementation, ticket_query, etc.)
- `ticketId` - Ticket identifier
- `action` - Requested action/command
- `data` - Additional data object
- `timestamp` - Event timestamp for tracking

**Maps these fields regardless of source naming conventions:**
- `ticket_id` / `ticketId` / `id` → `ticketId`
- `event_type` / `type` → `eventType`
- `action` / `command` → `action`

### 3. Check Event Type
**Type:** If Condition  
**ID:** `check_event_type`

Routes events to appropriate normalization handler based on `eventType`:
- **True branch:** `feature_implementation` → Feature implementation skill
- **False branch:** Other types → Ticket query or fallback

### 4. Normalization Nodes (Dynamic)

#### 4a. Normalize: Feature Implementation
**Type:** Assign  
**ID:** `normalize_feature_implementation`  
**Handles:** `feature_implementation` events

Transforms to feature implementation skill format:
```javascript
{
  skillCommand: "implement PROJ-001",
  skillName: "implementing-features",
  workspaceRoot: "/workspaces",
  repoPath: "/repo"
}
```

**Environment Variables:**
```bash
WORKSPACE_ROOT=/workspaces
REPO_PATH=/repo
```

#### 4b. Normalize: Ticket Query
**Type:** Assign  
**ID:** `normalize_ticket_query`  
**Handles:** `ticket_query` events

Transforms to ticket management skill format:
```javascript
{
  skillCommand: "get ticket PROJ-001",
  skillName: "managing-tickets"
}
```

#### 4c. Normalize: Fallback
**Type:** Assign  
**ID:** `normalize_fallback`  
**Handles:** Unknown event types

Passes through raw action:
```javascript
{
  skillCommand: "<raw action>",
  skillName: "custom"
}
```

### 5. Log Webhook Received
**Type:** HTTP Request  
**ID:** `log_webhook_received`  
**Endpoint:** `POST /webhooks/events`

Logs receipt of webhook before processing:
```json
{
  "event_type": "feature_implementation",
  "timestamp": 1705597800000,
  "status": "received"
}
```

### 6. Execute SSH Skill
**Type:** SSH  
**ID:** `execute_ssh_skill`  
**Requires:** SSH credentials configured

Executes the normalized skill command on remote server:

**Command Template:**
```bash
cd /home/user/codex-agent && \
  WORKSPACE_ROOT=/workspaces \
  REPO_PATH=/repo \
  bash scripts/skill-cli.sh "implement PROJ-001"
```

**SSH Details:**
- **Host:** Configured in credentials
- **User:** SSH credentials user
- **Port:** 22 (configurable)
- **Auth:** Private key or password

### 7. Process SSH Output
**Type:** Assign  
**ID:** `process_ssh_output`

Processes SSH response:
- Extracts stdout/stderr
- Calculates execution time
- Determines success/failure status
- Structures output for logging

### 8. Logging Nodes

#### 8a. Log Execution Success
**Type:** HTTP Request  
**ID:** `log_execution_success`  
**Endpoint:** `POST /webhooks/events`

Logs successful execution:
```json
{
  "event_type": "feature_implementation",
  "status": "success",
  "execution_time_ms": 45000,
  "output": { "...": "skill output" }
}
```

#### 8b. Log Execution Error
**Type:** HTTP Request  
**ID:** `log_execution_error`  
**Endpoint:** `POST /webhooks/events`

Logs execution errors:
```json
{
  "event_type": "feature_implementation",
  "status": "error",
  "error": "SSH connection timeout"
}
```

## Setup Instructions

### 1. Import Workflow

#### Via n8n UI:
1. Go to n8n Dashboard: `http://localhost:5678`
2. Click **"+" → "Import from file"**
3. Select `n8n-webhook-to-ssh-workflow.json`
4. Click **"Import"**

#### Via API:
```bash
curl -X POST http://localhost:5678/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d @n8n-webhook-to-ssh-workflow.json
```

### 2. Configure SSH Credentials

1. **Settings → Credentials**
2. **New → SSH**
3. **Name:** `ssh_credentials`
4. **Hostname:** Your server (e.g., `192.168.1.100` or `server.com`)
5. **Username:** SSH user (e.g., `deploy`)
6. **Authentication Type:**
   - **Private Key:** Paste your private key
   - **Password:** Enter SSH password
7. **Port:** 22 (default)
8. **Save**

### 3. Verify Webhook Endpoint

The webhook URL will be:
```
POST http://your-n8n-instance:5678/webhook/webhook-skill-execution
```

### 4. Test Workflow

Send a test webhook:

```bash
curl -X POST http://localhost:5678/webhook/webhook-skill-execution \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "feature_implementation",
    "ticket_id": "PROJ-001",
    "data": {
      "workspace_root": "/workspaces",
      "repo_path": "/repo"
    }
  }'
```

## Usage Examples

### Example 1: Implement Feature via Webhook

**Webhook Payload:**
```json
{
  "event_type": "feature_implementation",
  "ticket_id": "DARK-MODE-001",
  "data": {
    "workspace_root": "/home/deploy/workspaces",
    "repo_path": "/home/deploy/repos/app"
  }
}
```

**SSH Command Executed:**
```bash
cd /home/user/codex-agent && \
  WORKSPACE_ROOT=/home/deploy/workspaces \
  REPO_PATH=/home/deploy/repos/app \
  bash scripts/skill-cli.sh "implement DARK-MODE-001"
```

**Execution Flow:**
1. Webhook received with feature implementation request
2. Data extracted: ticket_id=DARK-MODE-001, event_type=feature_implementation
3. Normalized to skill command: "implement DARK-MODE-001"
4. SSH executes on remote server
5. Results logged back

### Example 2: Query Ticket via Webhook

**Webhook Payload:**
```json
{
  "event_type": "ticket_query",
  "ticketId": "API-AUTH-001"
}
```

**SSH Command Executed:**
```bash
cd /home/user/codex-agent && \
  bash scripts/skill-cli.sh "get ticket API-AUTH-001"
```

### Example 3: Custom Action via Webhook

**Webhook Payload:**
```json
{
  "event_type": "custom",
  "action": "list projects"
}
```

**SSH Command Executed:**
```bash
cd /home/user/codex-agent && \
  bash scripts/skill-cli.sh "list projects"
```

## Webhook Payload Specifications

### Common Fields

```json
{
  "event_type": "feature_implementation|ticket_query|custom",
  "ticket_id": "PROJ-001",
  "action": "custom command (optional)",
  "data": {
    "workspace_root": "/path/to/workspaces (optional)",
    "repo_path": "/path/to/repo (optional)"
  },
  "metadata": {
    "source": "github|jira|slack|etc",
    "user": "user@example.com",
    "timestamp": "2024-01-18T10:30:00Z"
  }
}
```

### Field Mapping

The workflow automatically maps these field names:

| Standard | Alternatives | Extracted As |
|----------|--------------|--------------|
| `event_type` | `type`, `event` | `eventType` |
| `ticket_id` | `ticketId`, `id` | `ticketId` |
| `action` | `command`, `skill_action` | `action` |
| `data` | `payload`, `params` | `data` |

### Environment Variables from Data

You can pass environment variables in the `data` object:

```json
{
  "event_type": "feature_implementation",
  "ticket_id": "PROJ-001",
  "data": {
    "workspace_root": "/custom/workspaces",
    "repo_path": "/custom/repo",
    "subagent_timeout": "1800000"
  }
}
```

These become shell environment variables in the SSH command.

## Error Handling

### SSH Connection Fails
- **Log:** Execution error logged to ticketing system
- **Retry:** Configure n8n retry policy on SSH node
- **Fallback:** Optional webhook notification to alerting system

### Skill Execution Timeout
- **Timeout:** Default 30 minutes (configurable via SUBAGENT_TIMEOUT)
- **Result:** Error logged, execution marked as failed
- **Recovery:** Manual retry or split into smaller tasks

### Webhook Authentication (Optional)
To secure the webhook, use n8n's webhook authentication:

1. **Workflow settings**
2. **Enable webhook authentication**
3. **Add authentication method**
4. Include auth in webhook URL:
```
POST http://localhost:5678/webhook/webhook-skill-execution?apiKey=YOUR_API_KEY
```

## Monitoring & Logging

### View Execution Logs
1. **Workflows → Select workflow**
2. **View → Execution history**
3. Click on execution to see details

### Check SSH Output
Each execution shows:
- **SSH stdout:** Skill execution output
- **SSH stderr:** Any error messages
- **Exit code:** Success (0) or failure (non-zero)

### Ticketing System Events
The workflow logs to `/webhooks/events` endpoint:

```bash
# View recent events
curl http://localhost:8080/webhooks/events?limit=20

# Filter by status
curl http://localhost:8080/webhooks/events?status=success
curl http://localhost:8080/webhooks/events?status=error
```

## Integration Examples

### GitHub Actions

Trigger from GitHub workflows:

```yaml
- name: Implement Feature
  run: |
    curl -X POST \
      -H "Content-Type: application/json" \
      -d '{
        "event_type": "feature_implementation",
        "ticket_id": "'"${{ github.event.pull_request.title }}"'",
        "data": {
          "repo_path": "/repo"
        }
      }' \
      http://n8n-server:5678/webhook/webhook-skill-execution
```

### GitLab CI/CD

```yaml
implement_feature:
  script:
    - |
      curl -X POST \
        -H "Content-Type: application/json" \
        -d "{
          \"event_type\": \"feature_implementation\",
          \"ticket_id\": \"$CI_MERGE_REQUEST_TITLE\",
          \"metadata\": {
            \"source\": \"gitlab\",
            \"user\": \"$GITLAB_USER_LOGIN\"
          }
        }" \
        http://n8n-server:5678/webhook/webhook-skill-execution
```

### Slack Bot

```python
@app.message("implement (.+)")
def implement_feature(message, say):
    ticket_id = message["text"].split()[-1]
    
    response = requests.post(
        "http://n8n-server:5678/webhook/webhook-skill-execution",
        json={
            "event_type": "feature_implementation",
            "ticket_id": ticket_id,
            "metadata": {
                "source": "slack",
                "user": message["user"]
            }
        }
    )
    
    if response.ok:
        say(f"✅ Implementing {ticket_id}...")
    else:
        say(f"❌ Failed to start implementation")
```

### Zapier Integration

1. **Create Zap**
2. **Trigger:** Any Zapier event (form submission, email, etc.)
3. **Action:** Webhooks → POST
4. **URL:** `http://n8n-server:5678/webhook/webhook-skill-execution`
5. **Payload:**
   ```json
   {
     "event_type": "feature_implementation",
     "ticket_id": "{{ ticketId }}",
     "data": {
       "repo_path": "/repo"
     }
   }
   ```

## Extending the Workflow

### Add New Skill Type

1. **Add condition** in "Check Event Type" node:
   ```
   IF eventType == "my_skill"
   ```

2. **Create normalization node:**
   - Type: Assign
   - Map raw fields to skill command format

3. **Connect to SSH node**

4. **Update documentation**

### Custom Data Extraction

Modify "Extract Webhook Data" node to handle your source format:

```javascript
{
  name: "projectName",
  value: "={{ $json.project?.name || $json.project_name }}"
}
```

### Add Pre-processing Steps

Insert nodes before SSH execution:
- Validate ticket exists
- Check permissions
- Get additional context
- Transform data further

## Troubleshooting

### Webhook Not Triggering

1. Check webhook URL is correct
2. Verify workflow is active (toggle in top-right)
3. Test with curl/Postman
4. Check n8n logs: `docker logs n8n`

### SSH Connection Refused

1. Verify SSH credentials are correct
2. Check server is running SSH service: `ssh user@host echo test`
3. Verify firewall allows port 22
4. Check user can run required commands: `ssh user@host bash scripts/skill-cli.sh`

### Skill Command Not Found

1. Verify script path: `/home/user/codex-agent/scripts/skill-cli.sh`
2. Check script is executable: `chmod +x scripts/skill-cli.sh`
3. Verify skill exists in skills directory
4. Check environment variables set correctly

### Slow Execution

1. SSH command execution slower than expected?
2. Check remote server load: `ssh user@host uptime`
3. Verify skill isn't timing out
4. Consider increasing SUBAGENT_TIMEOUT

## Security Considerations

1. **SSH Credentials:** Store in n8n credential vault, never expose
2. **Webhook URL:** Share only with trusted systems
3. **Payload Validation:** Add schema validation in custom nodes
4. **Access Control:** Limit SSH user permissions to required commands
5. **Audit Logging:** All executions logged to ticketing system

## References

- [Skills Documentation](/skills/implementing-features/SKILL.md)
- [SSH Skill Invocation](/SSH_SKILL_INVOCATION.md)
- [n8n SSH Node Docs](https://docs.n8n.io/integrations/builtin/execute/ssh/)
- [n8n Webhook Node Docs](https://docs.n8n.io/integrations/builtin/trigger-nodes/n8ntrigger-webhook/)
- [Ticketing System API](/README.md)

---

**Webhook → Normalize → Execute → Log = Universal Skill Orchestration**
