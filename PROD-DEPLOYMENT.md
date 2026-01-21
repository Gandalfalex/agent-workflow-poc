# Production Deployment Guide

This guide explains how to deploy the entire system (Traefik, Ticketing, n8n, Keycloak) using the production docker-compose.

## Architecture

```
                    ┌──────────────────────┐
                    │  Traefik (Port 80)   │
                    │  Dashboard (8080)    │
                    └──────────┬───────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
    ┌───▼────────┐    ┌───────▼────────┐    ┌───────▼──────┐
    │ Keycloak   │    │  Ticketing API │    │      n8n     │
    │ /auth/*    │    │  /ticketing/   │    │    /n8n/     │
    │ /realms/*  │    │  / (root)      │    │              │
    └────────────┘    └────────────────┘    └──────────────┘
```

## Services & Ports

| Service | Port | URL | Purpose |
|---------|------|-----|---------|
| Traefik | 80 | http://localhost/ | Main reverse proxy |
| Traefik Dashboard | 8080 | http://localhost:8080/dashboard/ | Monitor & debug |
| Keycloak | (via Traefik) | http://localhost/auth/ | Authentication |
| Ticketing | (via Traefik) | http://localhost/ticketing/ | Ticketing system |
| Ticketing (Root) | (via Traefik) | http://localhost/ | Fallback to ticketing |
| n8n | (via Traefik) | http://localhost/n8n/ | Workflow automation |

## Quick Start

### Local Development

```bash
cd /Users/ich/projects/coding-agent-workflow

# Build the ticketing system (if not already built)
docker compose -f docker-compose.prod.yaml build ticketing-api

# Start all services
docker compose -f docker-compose.prod.yaml up -d

# Check status
docker compose -f docker-compose.prod.yaml ps

# View logs
docker compose -f docker-compose.prod.yaml logs -f
```

### Access Services

- **Ticketing System**: http://localhost/
- **Ticketing Dashboard**: http://localhost/ticketing/
- **n8n**: http://localhost/n8n/
- **Keycloak Admin**: http://localhost/auth/admin
  - Username: `admin`
  - Password: `admin`
- **Traefik Dashboard**: http://localhost:8080/dashboard/

### Initialize Users

```bash
# Sync Keycloak users to ticketing database
./scripts/sync-keycloak-users.sh
```

## Deployment on Orange Pi (192.168.178.70)

### 1. Prepare Server

```bash
ssh 192.168.178.70

# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker (if not already installed)
curl -sSL https://get.docker.com | sh
sudo usermod -aG docker $USER
newgrp docker

# Verify Docker
docker --version
docker compose --version
```

### 2. Clone/Pull Project

```bash
# Clone or update the repository
cd ~
if [ ! -d projects ]; then mkdir -p projects; fi
cd projects

# Clone if not already cloned
if [ ! -d coding-agent-workflow ]; then
  git clone https://github.com/Gandalfalex/agent-workflow-poc.git coding-agent-workflow
fi

cd coding-agent-workflow
git pull origin main
```

### 3. Build and Start

```bash
# Build ticketing system (this takes 5-10 minutes)
docker compose -f docker-compose.prod.yaml build ticketing-api

# Start all services
docker compose -f docker-compose.prod.yaml up -d

# Check status
docker compose -f docker-compose.prod.yaml ps

# View logs
docker compose -f docker-compose.prod.yaml logs -f
```

### 4. Post-Deployment

```bash
# Wait for Keycloak to be ready (30-60 seconds)
sleep 60

# Sync users to ticketing database
./scripts/sync-keycloak-users.sh 192.168.178.70
```

### 5. Access Services

Access from any device on the network:
- **Ticketing**: http://192.168.178.70/
- **n8n**: http://192.168.178.70/n8n/
- **Keycloak Admin**: http://192.168.178.70/auth/admin
- **Traefik Dashboard**: http://192.168.178.70:8080/dashboard/

## Environment Variables

All services use environment variables from `docker-compose.prod.yaml`:

### Keycloak
- `KEYCLOAK_ADMIN`: Admin username (default: admin)
- `KEYCLOAK_ADMIN_PASSWORD`: Admin password (default: admin)

### Ticketing API
- `KEYCLOAK_BASE_URL`: Keycloak server (default: http://keycloak:8080)
- `KEYCLOAK_REALM`: Keycloak realm (default: ticketing)
- `KEYCLOAK_ADMIN_USER`: Admin user for sync (default: AdminUser)
- `KEYCLOAK_ADMIN_PASSWORD`: Admin password (default: admin123)

### n8n
- `N8N_PATH`: Base path for n8n (default: /n8n/)
- `WEBHOOK_URL`: Public webhook URL (default: http://localhost/n8n/)
- `GENERIC_TIMEZONE`: Timezone (default: Europe/Berlin)

## Managing Services

### View Status

```bash
docker compose -f docker-compose.prod.yaml ps
```

### View Logs

```bash
# All services
docker compose -f docker-compose.prod.yaml logs -f

# Specific service
docker compose -f docker-compose.prod.yaml logs -f ticketing-api
docker compose -f docker-compose.prod.yaml logs -f n8n
docker compose -f docker-compose.prod.yaml logs -f keycloak
```

### Restart Service

```bash
docker compose -f docker-compose.prod.yaml restart ticketing-api
```

### Stop All Services

```bash
docker compose -f docker-compose.prod.yaml down
```

### Stop and Remove Volumes (CAUTION - deletes data)

```bash
docker compose -f docker-compose.prod.yaml down -v
```

## Database Backups

### Backup All Databases

```bash
mkdir -p backups
DATE=$(date +%Y-%m-%d_%H-%M-%S)

# Keycloak
docker exec coding-agent-workflow-keycloak-postgres-1 \
  pg_dump -U keycloak keycloak > backups/keycloak-$DATE.sql

# Ticketing
docker exec coding-agent-workflow-ticketing-postgres-1 \
  pg_dump -U ticketing ticketing > backups/ticketing-$DATE.sql

# n8n
docker exec coding-agent-workflow-n8n-postgres-1 \
  pg_dump -U n8n n8n > backups/n8n-$DATE.sql

echo "Backups created in backups/"
ls -lh backups/
```

### Restore Database

```bash
# Ticketing
docker exec -i coding-agent-workflow-ticketing-postgres-1 \
  psql -U ticketing ticketing < backups/ticketing-2026-01-21.sql

# Or for all
for db in keycloak ticketing n8n; do
  docker exec -i coding-agent-workflow-${db}-postgres-1 \
    psql -U $db $db < backups/$db-*.sql
done
```

## Updating Services

### Update Ticketing System

```bash
# Pull latest code
git pull origin main

# Rebuild ticketing image
docker compose -f docker-compose.prod.yaml build ticketing-api

# Restart service
docker compose -f docker-compose.prod.yaml up -d ticketing-api

# Run migrations (if needed)
docker compose -f docker-compose.prod.yaml exec ticketing-api \
  /app/backend/cmd/migrate/migrate -database "postgres://ticketing:ticketing@ticketing-postgres:5432/ticketing?sslmode=disable" -path /app/backend/migrations up
```

### Update n8n

```bash
# Pull latest n8n image
docker compose -f docker-compose.prod.yaml pull n8n

# Restart
docker compose -f docker-compose.prod.yaml up -d n8n
```

### Update Keycloak

```bash
# Pull latest Keycloak image
docker compose -f docker-compose.prod.yaml pull keycloak

# Restart
docker compose -f docker-compose.prod.yaml up -d keycloak
```

## Monitoring

### Check Traefik Routes

1. Open http://localhost:8080/dashboard/
2. View HTTP routers, services, and middlewares
3. Check service health and response times

### Monitor Service Health

```bash
# Ticketing API
curl http://localhost/health

# Keycloak
curl http://localhost/realms/ticketing | jq .realm

# n8n (internal only)
docker exec coding-agent-workflow-n8n-1 curl http://localhost:5678/healthz
```

### Check Docker Resource Usage

```bash
docker stats coding-agent-workflow-*

# Or get detailed info
docker ps --format "table {{.Names}}\t{{.Size}}\t{{.Status}}"
```

## Troubleshooting

### Services Won't Start

```bash
# Check logs
docker compose -f docker-compose.prod.yaml logs

# Restart all
docker compose -f docker-compose.prod.yaml restart

# Check Docker daemon
docker ps
```

### Port 80 Already in Use

```bash
# Find what's using port 80
sudo lsof -i :80

# Stop conflicting service
sudo systemctl stop <service>

# Or use different port (edit docker-compose.prod.yaml)
```

### Traefik Routes Not Working

1. Check service labels: `docker inspect <container> | grep traefik`
2. Verify service is running: `docker ps | grep <service>`
3. Check Traefik logs: `docker logs <traefik-container>`
4. Restart Traefik: `docker compose -f docker-compose.prod.yaml restart traefik`

### n8n Path Issues

```bash
# Verify n8n is accessible directly
docker exec coding-agent-workflow-n8n-1 curl http://localhost:5678/

# Check if path redirect works
curl -L http://localhost/n8n/
```

### Keycloak 401 Errors

1. Verify users are synced: `curl http://localhost/api/ticketing/users | jq .`
2. Run sync manually: `./scripts/sync-keycloak-users.sh localhost`
3. Check Keycloak logs: `docker logs coding-agent-workflow-keycloak-1`

### Database Connection Issues

```bash
# Test database connection
docker exec coding-agent-workflow-ticketing-postgres-1 \
  psql -U ticketing -d ticketing -c "SELECT version();"

# Check database size
docker exec coding-agent-workflow-ticketing-postgres-1 \
  psql -U ticketing -d ticketing -c "SELECT pg_size_pretty(pg_database_size('ticketing'));"
```

## Security Hardening

### Change Default Passwords

Edit `docker-compose.prod.yaml` and change:

```yaml
KEYCLOAK_ADMIN_PASSWORD: [CHANGE_ME]
POSTGRES_PASSWORD: [CHANGE_ME]
```

### Enable HTTPS

Add to Traefik command:

```yaml
traefik:
  command:
    - "--entrypoints.websecure.address=:443"
    - "--certificatesresolvers.letsencrypt.acme.email=your-email@example.com"
    - "--certificatesresolvers.letsencrypt.acme.storage=/letsencrypt/acme.json"
    - "--certificatesresolvers.letsencrypt.acme.httpchallenge.entrypoint=web"
  volumes:
    - letsencrypt-data:/letsencrypt
```

### Restrict CORS

Change in ticketing-api:
```yaml
CORS_ALLOWED_ORIGINS: "https://your-domain.com,https://www.your-domain.com"
```

### Secure PostgreSQL

```bash
# Use strong passwords in docker-compose.prod.yaml
POSTGRES_PASSWORD: [use strong password]

# Connect only from containers (default in docker-compose)
# No external port exposure
```

## Performance Tuning

### Increase PostgreSQL Performance

Add to postgres services:

```yaml
environment:
  POSTGRES_INITDB_ARGS: "-c max_connections=200 -c shared_buffers=256MB"
```

### Increase n8n Concurrency

```yaml
n8n:
  environment:
    N8N_EXECUTIONS_DATA_MAX_AGE: 336
    N8N_EXECUTIONS_DATA_PRUNE: "true"
```

## System Requirements

**Minimum**:
- CPU: 2 cores
- RAM: 4 GB
- Storage: 20 GB (for data)
- Bandwidth: 1 Mbps

**Recommended (Production)**:
- CPU: 4+ cores
- RAM: 8+ GB
- Storage: 100+ GB
- Bandwidth: 10+ Mbps
- Backup storage: Equal to database size

## Support

For issues or questions:
1. Check logs: `docker compose -f docker-compose.prod.yaml logs -f`
2. Review Traefik dashboard: http://localhost:8080/dashboard/
3. Check service health endpoints
4. Consult service documentation:
   - Ticketing: See ticketing-system/README.md
   - n8n: https://docs.n8n.io/
   - Keycloak: https://www.keycloak.org/documentation
