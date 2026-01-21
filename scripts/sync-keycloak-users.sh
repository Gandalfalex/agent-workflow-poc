#!/bin/bash
set -e

# Configuration
TICKETING_API_URL="${TICKETING_API_URL:-http://localhost:8080}"
ADMIN_USERNAME="${ADMIN_USERNAME:-AdminUser}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123}"

echo "🔄 Syncing Keycloak users to ticketing database..."
echo "API: $TICKETING_API_URL"
echo ""

# Login to get session cookie
echo "1️⃣  Logging in as $ADMIN_USERNAME..."
LOGIN_RESPONSE=$(curl -s -c /tmp/ticketing_cookies.txt -X POST "$TICKETING_API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"identifier\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}")

if echo "$LOGIN_RESPONSE" | jq -e '.user' > /dev/null 2>&1; then
  echo "✅ Logged in successfully"
else
  echo "❌ Login failed:"
  echo "$LOGIN_RESPONSE" | jq .
  exit 1
fi
echo ""

# Call the sync endpoint
echo "2️⃣  Calling sync endpoint..."
SYNC_RESPONSE=$(curl -s -b /tmp/ticketing_cookies.txt -X POST "$TICKETING_API_URL/admin/sync-users")

if echo "$SYNC_RESPONSE" | jq -e '.synced' > /dev/null 2>&1; then
  SYNCED=$(echo "$SYNC_RESPONSE" | jq -r '.synced')
  TOTAL=$(echo "$SYNC_RESPONSE" | jq -r '.total')
  echo "✅ Sync complete!"
  echo "📊 Synced: $SYNCED / $TOTAL users"
else
  echo "❌ Sync failed:"
  echo "$SYNC_RESPONSE" | jq .
  rm -f /tmp/ticketing_cookies.txt
  exit 1
fi
echo ""

# Verify users are searchable
echo "3️⃣  Verifying users..."
USERS_RESPONSE=$(curl -s -b /tmp/ticketing_cookies.txt "$TICKETING_API_URL/users")
USER_COUNT=$(echo "$USERS_RESPONSE" | jq '.items | length')
echo "✅ Found $USER_COUNT users in database"
echo ""
echo "$USERS_RESPONSE" | jq -r '.items[] | "  👤 \(.name) <\(.email)>"'

# Cleanup
rm -f /tmp/ticketing_cookies.txt

echo ""
echo "✨ All done! Users are now searchable."
