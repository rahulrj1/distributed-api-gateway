#!/bin/bash
# E2E Test: Authentication - Verifies JWT validation works correctly

set -e

GATEWAY_URL="${GATEWAY_URL:-http://localhost:5000}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }

echo "=== E2E Test: Authentication ==="
echo "Gateway: $GATEWAY_URL"
echo ""

cd "$PROJECT_ROOT"

# Ensure keys exist
if [ ! -f "keys/private.pem" ]; then
    mkdir -p keys
    openssl genrsa -out keys/private.pem 2048 2>/dev/null
    openssl rsa -in keys/private.pem -pubout -out keys/public.pem 2>/dev/null
fi

# Generate tokens with unique client_id
CLIENT_ID="auth-$(date +%s)"
VALID_TOKEN=$(go run scripts/generate_jwt.go -sub testuser -client_id "$CLIENT_ID" -exp 1h)
EXPIRED_TOKEN=$(go run scripts/generate_jwt.go -sub testuser -client_id "$CLIENT_ID" -exp -1s)
echo "Tokens generated"
echo ""

# Test 1: Request without token should return 401
echo "Test 1: No token..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/service-a/hello")
if [ "$STATUS" = "401" ]; then
    pass "No token returns 401"
else
    fail "No token should return 401, got: $STATUS"
fi

# Test 2: Request with valid token should return 200
echo "Test 2: Valid token..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $VALID_TOKEN" "$GATEWAY_URL/service-a/hello")
if [ "$STATUS" = "200" ]; then
    pass "Valid token returns 200"
else
    fail "Valid token should return 200, got: $STATUS"
fi

# Test 3: Request with expired token should return 401
echo "Test 3: Expired token..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $EXPIRED_TOKEN" "$GATEWAY_URL/service-a/hello")
if [ "$STATUS" = "401" ]; then
    pass "Expired token returns 401"
else
    fail "Expired token should return 401, got: $STATUS"
fi

# Test 4: Request with invalid token should return 401
echo "Test 4: Invalid token..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer invalid.token.here" "$GATEWAY_URL/service-a/hello")
if [ "$STATUS" = "401" ]; then
    pass "Invalid token returns 401"
else
    fail "Invalid token should return 401, got: $STATUS"
fi

# Test 5: Health endpoint works without token
echo "Test 5: Health without token..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health")
if [ "$STATUS" = "200" ]; then
    pass "Health works without token"
else
    fail "Health should return 200, got: $STATUS"
fi

# Test 6: Metrics endpoint works without token
echo "Test 6: Metrics without token..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/metrics")
if [ "$STATUS" = "200" ]; then
    pass "Metrics works without token"
else
    fail "Metrics should return 200, got: $STATUS"
fi

echo ""
echo -e "${GREEN}All authentication tests passed!${NC}"
