# Production Deployment with Shadowing Proxy

This setup uses the **shadowing reverse proxy** as the main entry point. The proxy can mirror traffic to new workspace deployments for comparison and safe feature rollout.

## Architecture

```
                    ┌──────────────────────────────────┐
                    │   Shadowing Proxy (Port 80)      │
                    │   Admin UI (Port 8081)           │
                    │   Traffic Mirroring & Comparison │
                    └──────────────┬───────────────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
    ┌───▼────────────┐    ┌───────▼─────────┐    ┌──────────▼────┐
    │   Keycloak     │    │ Ticketing API   │    │      n8n      │
    │   /auth/*      │    │ (Production)    │    │ (Production)  │
    │   /realms/*    │    │ /api/ticketing/ │    │   /n8n/       │
    └────────────────┘    └─────────────────┘    └───────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │  New Feature Branch         │
                    │  Ticketing API (Shadow)     │
                    │  (from worktree)            │
                    │  Compared & Mirrored        │
                    └─────────────────────────────┘
```

## Key Features

- **Production Primary**: Main `ticketing-api-prod` service
- **Shadow Deployments**: New versions from git worktrees
- **Traffic Mirroring**: Automatically duplicate requests to shadow services
- **Response Comparison**: Shadowing DB stores comparison results
- **Admin UI**: Manage routes and view results at port 8081

## Quick Start

### 1. Build and Start Production

```bash
cd /Users/ich/projects/coding-agent-workflow

# Build shadowing and ticketing
docker compose -f docker-compose.prod.yaml build

# Start all services
docker compose -f docker-compose.prod.yaml up -d

# Verify
docker compose -f docker-compose.prod.yaml ps
```

### 2. Access Services

- **Main Entry**: http://localhost/ (shadowing proxy)
- **Ticketing**: http://localhost/api/ticketing/ (routed through shadowing)
- **n8n**: http://localhost/n8n/
- **Keycloak Admin**: http://localhost/auth/admin
- **Shadowing Admin UI**: http://localhost:8081

### 3. Sync Users

```bash
./scripts/sync-keycloak-users.sh localhost
```

## Deploying Shadow/Feature Versions

The shadowing proxy can mirror traffic to multiple versions for comparison. Here's how to deploy a new feature version:

### 1. Create a Git Worktree for Your Feature

```bash
# From coding-agent-workflow directory
git worktree add worktrees/feature-name feature-branch

# Or from a local branch
git worktree add worktrees/feature-name -b feature-name origin/main
cd worktrees/feature-name
# Make your changes
git commit -am "your changes"
```

### 2. Uncomment Shadow Service in docker-compose.prod.yaml

The compose file has example shadow services commented out. Uncomment `ticketing-api-shadow`:

```yaml
  ticketing-api-shadow:
    build:
      context: ./worktrees/feature-name/ticketing-system
      dockerfile: backend/Dockerfile
    environment:
      PORT: "8080"
      DATABASE_URL: postgres://ticketing:ticketing@ticketing-postgres-shadow:5432/ticketing?sslmode=disable
      # ... rest of config
    labels:
      - "traefik.http.services.ticketing-shadow.loadbalancer.server.port=8080"

  ticketing-postgres-shadow:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: ticketing
      POSTGRES_PASSWORD: ticketing
      POSTGRES_DB: ticketing
    volumes:
      - ticketing-shadow-data:/var/lib/postgresql/data
```

### 3. Rebuild and Deploy

```bash
# Rebuild with new shadow service
docker compose -f docker-compose.prod.yaml build ticketing-api-shadow

# Start shadow service
docker compose -f docker-compose.prod.yaml up -d ticketing-api-shadow

# Verify it's running
docker compose -f docker-compose.prod.yaml ps | grep shadow
```

### 4. Configure Shadowing in Admin UI

1. Access http://localhost:8081
2. Create a new route:
   - **Primary target**: `http://ticketing-api-prod:8080`
   - **Shadow target**: `http://ticketing-api-shadow:8080`
3. Enable shadowing/comparison
4. Start making requests - shadowing will mirror them and compare responses

### 5. Review Comparison Results

In the shadowing admin UI:
- View side-by-side response comparisons
- Check for differences in:
  - HTTP status codes
  - Response headers
  - Response body
  - Latency
- Identify issues before rolling out

### 6. Promote Shadow to Production

Once testing is complete:

```bash
# Merge feature branch to main
git checkout main
git merge feature-name
git push origin main

# Rebuild production service
docker compose -f docker-compose.prod.yaml build ticketing-api-prod

# Restart production
docker compose -f docker-compose.prod.yaml up -d ticketing-api-prod

# Disable shadow or remove from compose
# docker compose -f docker-compose.prod.yaml down ticketing-api-shadow
```

## Service Discovery

The shadowing proxy uses Docker labels for service discovery:

```yaml
labels:
  - "traefik.http.routers.<name>.rule=PathPrefix(`/path`)"
  - "traefik.http.services.<name>.loadbalancer.server.port=<port>"
```

Services on the `prod-network` are automatically discovered. The proxy reads labels and creates routes.

## Workflow: Skills/Agents Creating Workspaces

When agents create feature workspaces:

```
1. Agent creates git worktree: worktrees/feature-name/
2. Agent modifies code
3. Agent commits changes to worktree branch
4. Admin uncomments shadow service in docker-compose
5. Admin runs: docker compose build && docker compose up -d
6. Shadowing automatically discovers the new service
7. Traffic is mirrored to both prod and shadow
8. Comparison results appear in shadowing admin UI
9. After review, promote to production or discard
```

## Database Management

### Multiple Databases for Shadowing

Each deployment (prod, shadow) has separate databases:

```
Production:
  - ticketing-postgres-prod
  - n8n-postgres (shared)
  - keycloak-postgres (shared)

Shadow (feature branches):
  - ticketing-postgres-shadow
  - n8n-postgres-shadow (if needed)
```

### Backup Production Databases

```bash
mkdir -p backups
DATE=$(date +%Y-%m-%d_%H-%M-%S)

docker exec coding-agent-workflow-ticketing-postgres-prod-1 \
  pg_dump -U ticketing ticketing > backups/ticketing-prod-$DATE.sql

docker exec coding-agent-workflow-n8n-postgres-1 \
  pg_dump -U n8n n8n > backups/n8n-prod-$DATE.sql

docker exec coding-agent-workflow-keycloak-postgres-1 \
  pg_dump -U keycloak keycloak > backups/keycloak-prod-$DATE.sql
```

### Reset Shadow Database

```bash
# If shadow DB gets corrupted, reset it
docker compose -f docker-compose.prod.yaml down ticketing-postgres-shadow
docker volume rm coding-agent-workflow_ticketing-shadow-data
docker compose -f docker-compose.prod.yaml up -d ticketing-postgres-shadow
```

## Orange Pi Deployment

### 1. Initial Setup

```bash
ssh gandalfalex@192.168.178.70

cd ~/projects/coding-agent-workflow
git pull origin main

# Build all images (takes 10-15 minutes)
docker compose -f docker-compose.prod.yaml build

# Start production
docker compose -f docker-compose.prod.yaml up -d
```

### 2. Access on Network

- **Main**: http://192.168.178.70/
- **Shadowing Admin**: http://192.168.178.70:8081

### 3. Deploy New Features

```bash
# On your local machine, push feature branch
git push origin feature-name

# On server, create worktree and deploy
ssh gandalfalex@192.168.178.70
cd ~/projects/coding-agent-workflow
git fetch origin
git worktree add worktrees/feature-name origin/feature-name

# Uncomment shadow service in docker-compose.prod.yaml
# Then:
docker compose -f docker-compose.prod.yaml build ticketing-api-shadow
docker compose -f docker-compose.prod.yaml up -d ticketing-api-shadow

# Access shadowing admin UI at http://192.168.178.70:8081
# Configure mirroring and test
```

## Managing Shadow Deployments

### List Active Shadows

```bash
docker compose -f docker-compose.prod.yaml ps | grep shadow
```

### View Shadow Logs

```bash
docker compose -f docker-compose.prod.yaml logs -f ticketing-api-shadow
```

### Stop Shadow Service

```bash
docker compose -f docker-compose.prod.yaml stop ticketing-api-shadow
```

### Remove Shadow Service

```bash
# Stop and remove
docker compose -f docker-compose.prod.yaml down ticketing-api-shadow

# Clean up volume
docker volume rm coding-agent-workflow_ticketing-shadow-data

# Remove from compose file
vim docker-compose.prod.yaml  # Remove shadow services
```

## Monitoring & Troubleshooting

### Check Shadowing Proxy Health

```bash
curl http://localhost:8081/api/status
```

### View Service Discovery

```bash
# Check what services shadowing discovered
docker logs coding-agent-workflow-shadowing-1 | grep -i "discovered"
```

### Verify Shadow Mirroring

```bash
# Make a request to ticketing
curl http://localhost/api/ticketing/users

# Check shadowing DB for comparison
docker exec coding-agent-workflow-shadowing-db-1 \
  psql -U shadowing -d shadowing -c "SELECT * FROM comparisons ORDER BY created_at DESC LIMIT 1;"
```

### Debug Route Not Found

```bash
# Check shadowing logs
docker logs coding-agent-workflow-shadowing-1

# Verify service labels
docker inspect coding-agent-workflow-ticketing-api-prod-1 | grep -A 5 traefik

# Check network connectivity
docker network inspect prod-network
```

## Workflow Examples

### Example 1: Testing a Bug Fix

```bash
# 1. Create fix branch
git checkout -b fix/bug-123
# ... make changes ...
git commit -am "fix: bug 123"
git push origin fix/bug-123

# 2. On server: Deploy as shadow
git worktree add worktrees/fix-bug-123 origin/fix/bug-123

# 3. Edit docker-compose.prod.yaml - uncomment shadow service
# 4. Deploy
docker compose -f docker-compose.prod.yaml build ticketing-api-shadow
docker compose -f docker-compose.prod.yaml up -d ticketing-api-shadow

# 5. Test via shadowing admin UI (http://localhost:8081)
# 6. If good: merge to main
git checkout main && git merge fix/bug-123 && git push origin main

# 7. Promote shadow to prod
docker compose -f docker-compose.prod.yaml build ticketing-api-prod
docker compose -f docker-compose.prod.yaml up -d ticketing-api-prod
```

### Example 2: Testing New Feature

```bash
# 1. Feature branch with agent-generated code
git checkout -b feature/dashboard-redesign
# ... agent creates code ...
git push origin feature/dashboard-redesign

# 2. Deploy as shadow
git worktree add worktrees/feature-dashboard origin/feature/dashboard-redesign

# 3. Uncomment + deploy shadow
# 4. Route 10% of traffic to shadow (configure in shadowing UI)
# 5. Monitor response comparisons
# 6. Gradually increase shadow traffic to 50%
# 7. Full rollover when confident
```

## Performance Tuning

### Shadowing Database

For high-traffic scenarios:

```yaml
shadowing-db:
  environment:
    POSTGRES_INITDB_ARGS: "-c max_connections=200 -c shared_buffers=256MB -c effective_cache_size=1GB"
```

### Disable Comparison Logging (if high traffic)

In shadowing admin UI, disable storing all comparisons and only keep samples.

## Security

- Shadowing admin UI (8081) should be restricted to internal networks
- Shadow deployments can access same databases - ensure proper isolation
- Use strong passwords for all databases
- Don't expose shadowing admin UI to public internet

## Next Steps

1. Deploy production version: `docker compose -f docker-compose.prod.yaml up -d`
2. Create first feature branch with agent
3. Deploy as shadow (uncomment + build)
4. Test via shadowing admin UI
5. Promote to production when ready

The shadowing approach enables safe feature testing with live traffic comparison!
