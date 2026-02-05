#!/bin/bash
# E2E Test: Circuit Breaker - Verifies cross-instance circuit breaker works

set -e

GATEWAY_1="${GATEWAY_1:-http://localhost:5000}"
GATEWAY_2="${GATEWAY_2:-http://localhost:5001}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }

echo "=== E2E Test: Circuit Breaker ==="
echo "Gateway 1: $GATEWAY_1"
echo "Gateway 2: $GATEWAY_2"
echo ""

cd "$PROJECT_ROOT"

# Ensure keys exist
if [ ! -f "keys/private.pem" ]; then
    mkdir -p keys
    openssl genrsa -out keys/private.pem 2048 2>/dev/null
    openssl rsa -in keys/private.pem -pubout -out keys/public.pem 2>/dev/null
fi

TOKEN=$(go run scripts/generate_jwt.go -sub cb-test -client_id "cb-$(date +%s)" -exp 1h)

# Clear any existing circuit breaker state for service-a
echo "Clearing circuit breaker state..."
docker exec distributed-api-gateway-redis-1 redis-cli KEYS "circuit:/service-a:*" | xargs -r docker exec -i distributed-api-gateway-redis-1 redis-cli DEL 2>/dev/null || true
echo ""

# Test 1: Healthy backend
echo "Test 1: Healthy backend..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
if [ "$STATUS" = "200" ]; then
    pass "Healthy backend returns 200"
else
    fail "Healthy backend should return 200, got: $STATUS"
fi

# Test 2: Stop service and trigger failures
echo ""
echo "Test 2: Trigger circuit breaker..."

# Ensure service is running first, then stop it
docker start distributed-api-gateway-service-a-1 2>/dev/null || true
sleep 2
docker stop distributed-api-gateway-service-a-1 2>/dev/null || docker-compose stop service-a 2>/dev/null
sleep 3

echo "Sending requests to trigger failures (need 5+ failures)..."
CIRCUIT_OPENED=false
for i in $(seq 1 15); do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
    echo "  Request $i: $STATUS"
    if [ "$STATUS" = "503" ]; then
        CIRCUIT_OPENED=true
        pass "Circuit opened after $i requests"
        break
    fi
done

if [ "$CIRCUIT_OPENED" = "false" ]; then
    fail "Circuit did not open after 15 failures"
fi

# Test 3: Circuit should be OPEN
echo ""
echo "Test 3: Circuit OPEN..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
if [ "$STATUS" = "503" ]; then
    pass "Circuit is OPEN - returns 503"
else
    fail "Expected 503 (circuit open), got: $STATUS"
fi

# Test 4: Cross-instance circuit state shared
echo ""
echo "Test 4: Cross-instance state..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_2/service-a/hello")
if [ "$STATUS" = "503" ]; then
    pass "Circuit state shared - Gateway 2 returns 503"
else
    fail "Gateway 2 should return 503, got: $STATUS"
fi

# Test 5: Other services unaffected
echo ""
echo "Test 5: Other services unaffected..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-b/hello")
if [ "$STATUS" = "200" ]; then
    pass "Service B still works"
else
    fail "Service B should work, got: $STATUS"
fi

# Test 6: Recovery after cooldown
echo ""
echo "Test 6: Recovery..."
echo "Restarting service-a..."
docker start distributed-api-gateway-service-a-1 2>/dev/null || docker-compose start service-a 2>/dev/null

echo "Waiting for service-a to be healthy..."
for i in $(seq 1 10); do
    if curl -s "http://localhost:6000/health" > /dev/null 2>&1; then
        break
    fi
    sleep 1
done

echo "Waiting 30s for circuit breaker cooldown..."
sleep 30

STATUS=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$GATEWAY_1/service-a/hello")
if [ "$STATUS" = "200" ]; then
    pass "Circuit recovered - returns 200"
else
    fail "Circuit should recover, got: $STATUS"
fi

echo ""
echo -e "${GREEN}All circuit breaker tests passed!${NC}"
