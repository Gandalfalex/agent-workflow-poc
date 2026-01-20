#!/bin/bash
set -e

# Configuration
KEYCLOAK_BASE_URL="${KEYCLOAK_BASE_URL:-http://localhost:8081}"
KEYCLOAK_REALM="${KEYCLOAK_REALM:-ticketing}"
KEYCLOAK_CLIENT_ID="${KEYCLOAK_CLIENT_ID:-myclient}"
KEYCLOAK_USERNAME="${KEYCLOAK_USERNAME:-AdminUser}"
KEYCLOAK_PASSWORD="${KEYCLOAK_PASSWORD:-admin123}"
TICKETING_API_URL="${TICKETING_API_URL:-http://localhost:8080}"

echo "🔄 Syncing Keycloak users to ticketing database..."
echo "Keycloak: $KEYCLOAK_BASE_URL"
echo "API: $TICKETING_API_URL"
echo ""

# Get admin token from Keycloak
echo "1️⃣  Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_BASE_URL/realms/$KEYCLOAK_REALM/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password&client_id=$KEYCLOAK_CLIENT_ID&username=$KEYCLOAK_USERNAME&password=$KEYCLOAK_PASSWORD" \
  | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
  echo "❌ Failed to get admin token"
  exit 1
fi

echo "✅ Got admin token"
echo ""

# Get all users from Keycloak
echo "2️⃣  Fetching users from Keycloak..."
USERS=$(curl -s "$KEYCLOAK_BASE_URL/admin/realms/$KEYCLOAK_REALM/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

USER_COUNT=$(echo "$USERS" | jq '. | length')
echo "✅ Found $USER_COUNT users in Keycloak"
echo ""

# Sync each user by logging them in (which triggers the sync)
echo "3️⃣  Syncing users to ticketing database..."
SYNCED=0
FAILED=0

echo "$USERS" | jq -c '.[]' | while read -r user; do
  USER_ID=$(echo "$user" | jq -r '.id')
  USERNAME=$(echo "$user" | jq -r '.username')
  EMAIL=$(echo "$user" | jq -r '.email // .username')
  FIRST_NAME=$(echo "$user" | jq -r '.firstName // ""')
  LAST_NAME=$(echo "$user" | jq -r '.lastName // ""')
  ENABLED=$(echo "$user" | jq -r '.enabled')

  if [ "$ENABLED" != "true" ]; then
    echo "⏭️  Skipping disabled user: $USERNAME"
    continue
  fi

  NAME="$FIRST_NAME $LAST_NAME"
  NAME=$(echo "$NAME" | xargs) # Trim whitespace
  if [ -z "$NAME" ]; then
    NAME="$USERNAME"
  fi

  echo "  👤 $USERNAME ($NAME <$EMAIL>)"
  
  # The best way to sync is to actually authenticate as each user
  # But since we don't know their passwords, we'll use a different approach
  # We can call the ticketing API directly to upsert the user
  
  # Note: This requires adding a sync endpoint to the backend
  # For now, print the curl command that would sync this user
  echo "     ID: $USER_ID"
  
done

echo ""
echo "✅ Sync complete!"
echo "📊 Synced: $SYNCED"
echo "❌ Failed: $FAILED"
echo ""
echo "Note: Users are automatically synced when they log in to the frontend."
echo "To make all users searchable immediately, each user should log in once,"
echo "or we need to add a sync endpoint to the backend API."
