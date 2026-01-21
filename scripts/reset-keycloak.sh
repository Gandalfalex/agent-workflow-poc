#!/bin/bash
set -e

echo "🔄 Resetting Keycloak with fresh configuration..."
echo ""

# Stop containers
echo "1️⃣  Stopping Keycloak containers..."
docker compose stop keycloak keycloak-postgres
echo "✅ Containers stopped"
echo ""

# Remove containers
echo "2️⃣  Removing containers..."
docker compose rm -f keycloak keycloak-postgres
echo "✅ Containers removed"
echo ""

# Find and remove the volume
echo "3️⃣  Removing Keycloak database volume..."
VOLUME_NAME=$(docker volume ls --format '{{.Name}}' | grep keycloak_pgdata || echo "")
if [ -n "$VOLUME_NAME" ]; then
  echo "Found volume: $VOLUME_NAME"
  docker volume rm "$VOLUME_NAME"
  echo "✅ Volume removed"
else
  echo "⚠️  No keycloak_pgdata volume found (might already be deleted)"
fi
echo ""

# Recreate containers
echo "4️⃣  Starting fresh Keycloak..."
docker compose up -d keycloak
echo "✅ Keycloak starting..."
echo ""

# Wait for Keycloak to be ready
echo "5️⃣  Waiting for Keycloak to be ready (this may take 20-30 seconds)..."
READY=false
for i in {1..30}; do
  if curl -s http://localhost:8081/realms/ticketing > /dev/null 2>&1; then
    READY=true
    break
  fi
  echo -n "."
  sleep 2
done
echo ""

if [ "$READY" = true ]; then
  echo "✅ Keycloak is ready!"
else
  echo "❌ Keycloak did not become ready in time"
  echo "Check logs with: docker logs keycloak"
  exit 1
fi
echo ""

echo "✨ Keycloak has been reset with fresh configuration!"
echo ""
echo "Next steps:"
echo "1. Restart ticketing-api: docker compose restart ticketing-api"
echo "2. Run sync script: ./scripts/sync-keycloak-users.sh"
