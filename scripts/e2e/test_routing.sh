#!/bin/bash
# E2E Test: Routing - Verifies requests reach correct backend services

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

echo "=== E2E Test: Routing ==="
echo "Gateway: $GATEWAY_URL"
echo ""

cd "$PROJECT_ROOT"

# Ensure keys exist
if [ ! -f "keys/private.pem" ]; then
    mkdir -p keys
    openssl genrsa -out keys/private.pem 2048 2>/dev/null
    openssl rsa -in keys/private.pem -pubout -out keys/public.pem 2>/dev/null
fi

# Always generate fresh token with unique client_id
TOKEN=$(go run scripts/generate_jwt.go -sub routing-test -client_id "routing-$(date +%s)" -exp 1h)

# Test 1: Service A (Python)
echo "Test 1: Service A routing..."
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/service-a/hello")
if echo "$RESPONSE" | grep -q "Python"; then
    pass "Service A reachable - got Python response"
else
    fail "Service A not reachable - got: $RESPONSE"
fi

# Test 2: Service B (Java)
echo "Test 2: Service B routing..."
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/service-b/hello")
if echo "$RESPONSE" | grep -q "Java"; then
    pass "Service B reachable - got Java response"
else
    fail "Service B not reachable - got: $RESPONSE"
fi

# Test 3: Service C (Node)
echo "Test 3: Service C routing..."
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/service-c/hello")
if echo "$RESPONSE" | grep -q "Node"; then
    pass "Service C reachable - got Node response"
else
    fail "Service C not reachable - got: $RESPONSE"
fi

# Test 4: Unknown path returns 404
echo "Test 4: Unknown path..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_URL/unknown/path")
if [ "$STATUS" = "404" ]; then
    pass "Unknown path returns 404"
else
    fail "Unknown path should return 404, got: $STATUS"
fi

# Test 5: Health endpoint (no auth required)
echo "Test 5: Health endpoint..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health")
if [ "$STATUS" = "200" ]; then
    pass "Health endpoint works"
else
    fail "Health endpoint should return 200, got: $STATUS"
fi

echo ""
echo -e "${GREEN}All routing tests passed!${NC}"
