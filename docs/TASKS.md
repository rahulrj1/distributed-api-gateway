# Task Breakdown

Breaking the API Gateway project into independent, PR-sized tasks.

---

## Task Overview

| # | Task | Dependencies | Estimated Effort |
|---|------|--------------|------------------|
| 1 | Service A (Python/Flask) | None | Small |
| 2 | Service B (Java/Spring) | None | Medium |
| 3 | Service C (Node/Express) | None | Small |
| 4 | Gateway Skeleton + Health Endpoints | None | Small |
| 5 | Gateway Routing + Proxy | Task 4 | Medium |
| 6 | Docker Compose (all services) | Tasks 1-5 | Small |
| 7 | JWT Authentication Middleware | Task 6 | Medium |
| 8 | Rate Limiting Middleware | Task 6 | Medium |
| 9 | Circuit Breaker Middleware | Task 6 | Medium |
| 10 | Observability (Metrics + Tracing) | Task 6 | Medium |
| 11 | Integration & E2E Tests | Tasks 7-10 | Medium |
| 12 | README + Final Polish | All | Small |

---

## Testing Strategy

Each task includes:
- **Unit tests**: Test individual functions in isolation
- **Integration tests**: Test with real dependencies (Redis, backends)
- **E2E tests**: Test full request flow through gateway

Test location: `gateway/tests/` and `scripts/e2e/`

---

## Task 1: Service A (Python/Flask)

**Goal**: Simple Python backend service

**Deliverables**:
- `services/service-a/app.py`
- `services/service-a/requirements.txt`
- `services/service-a/Dockerfile`

**Endpoints**:
- `GET /health` → `{"status": "healthy", "service": "service-a"}`
- `GET /hello` → `{"message": "Hello from Python Service A"}`
- `POST /echo` → Echo back request body, headers, method

**Acceptance Criteria**:
- Runs locally: `python app.py`
- Runs in Docker: `docker build && docker run`
- All 3 endpoints work

**Independent**: Yes, no dependencies on other tasks

---

## Task 2: Service B (Java/Spring Boot)

**Goal**: Simple Java backend service

**Deliverables**:
- `services/service-b/src/main/java/...`
- `services/service-b/pom.xml`
- `services/service-b/Dockerfile`

**Endpoints**:
- `GET /health` → `{"status": "healthy", "service": "service-b"}`
- `GET /hello` → `{"message": "Hello from Java Service B"}`
- `POST /echo` → Echo back request body, headers, method

**Acceptance Criteria**:
- Runs locally: `mvn spring-boot:run`
- Runs in Docker: `docker build && docker run`
- All 3 endpoints work

**Independent**: Yes, no dependencies on other tasks

---

## Task 3: Service C (Node.js/Express)

**Goal**: Simple Node.js backend service

**Deliverables**:
- `services/service-c/index.js`
- `services/service-c/package.json`
- `services/service-c/Dockerfile`

**Endpoints**:
- `GET /health` → `{"status": "healthy", "service": "service-c"}`
- `GET /hello` → `{"message": "Hello from Node Service C"}`
- `POST /echo` → Echo back request body, headers, method

**Acceptance Criteria**:
- Runs locally: `npm start`
- Runs in Docker: `docker build && docker run`
- All 3 endpoints work

**Independent**: Yes, no dependencies on other tasks

---

## Task 4: Gateway Skeleton + Health Endpoints

**Goal**: Basic Go HTTP server with health checks

**Deliverables**:
- `gateway/main.go`
- `gateway/go.mod`
- `gateway/Dockerfile`
- `gateway/config/config.go`
- `gateway/handler/health.go`

**Endpoints**:
- `GET /health` → `{"status": "healthy"}`
- `GET /metrics` → Placeholder (empty for now)

**Acceptance Criteria**:
- Runs locally: `go run main.go`
- Runs in Docker
- Health endpoint responds

**Independent**: Yes, no dependencies on other tasks

---

## Task 5: Gateway Routing + Proxy

**Goal**: Path-based routing to backend services

**Deliverables**:
- `gateway/config/routes.yaml`
- `gateway/proxy/forwarder.go`
- `gateway/handler/proxy.go`
- Unit tests

**Functionality**:
- Load routes from YAML config
- Match request path to route
- Strip prefix if configured
- Forward request to backend
- Stream response back
- Add `X-Request-ID`, `X-Forwarded-For` headers

**Acceptance Criteria**:
- `GET /service-a/hello` → forwards to Service A, returns response
- `GET /service-b/hello` → forwards to Service B, returns response
- `GET /service-c/hello` → forwards to Service C, returns response
- Unknown path → 404

**Dependencies**: Task 4 (gateway skeleton)

---

## Task 6: Docker Compose

**Goal**: Run all services together

**Deliverables**:
- `docker-compose.yml`
- `prometheus/prometheus.yml`

**Services**:
- gateway-1 (port 5000)
- gateway-2 (port 5001)
- service-a (port 6000)
- service-b (port 6001)
- service-c (port 6002)
- redis (port 6379)
- prometheus (port 9090)

**Acceptance Criteria**:
- `docker compose up` starts all services
- Health checks pass
- Requests through gateway reach backends

**Dependencies**: Tasks 1-5 (all services exist)

---

## Task 7: JWT Authentication Middleware

**Goal**: Validate JWT tokens at gateway

**Deliverables**:
- `gateway/middleware/auth.go`
- `gateway/pkg/jwt/validator.go`
- `scripts/generate_jwt.go`
- `keys/` directory setup instructions
- Unit tests

**Functionality**:
- Extract token from `Authorization: Bearer <token>`
- Validate RS256 signature with public key
- Check expiration
- Extract `sub` → `X-User-ID` header
- Extract `client_id` → context for rate limiting
- Skip auth for `/health`, `/metrics`

**Acceptance Criteria**:
- Valid token → request forwarded with `X-User-ID`
- Missing token → 401
- Expired token → 401
- Invalid signature → 401
- `/health` works without token

**Dependencies**: Task 6 (Docker Compose)

---

## Task 8: Rate Limiting Middleware

**Goal**: Redis-backed sliding window rate limiting

**Deliverables**:
- `gateway/middleware/ratelimit.go`
- `gateway/pkg/ratelimit/slidingwindow.go`
- `gateway/pkg/redis/client.go`
- Unit tests + Integration tests

**Functionality**:
- Sliding window counter algorithm
- Lua script for atomic check-and-increment
- Rate limit by `client_id` (from JWT), fallback to IP
- Return 429 with `Retry-After` header when exceeded
- Fail-open if Redis unavailable

**Acceptance Criteria**:
- Under limit → requests allowed
- Over limit → 429 with `Retry-After`
- Redis down → requests allowed (fail-open)
- **Cross-instance**: Limit shared across gateway instances

**Dependencies**: Task 6 (Docker Compose with Redis)

---

## Task 9: Circuit Breaker Middleware

**Goal**: Protect backends from cascade failures

**Deliverables**:
- `gateway/middleware/circuitbreaker.go`
- `gateway/pkg/circuitbreaker/breaker.go`
- Unit tests + Integration tests

**Functionality**:
- Track failures per service in Redis
- CLOSED → OPEN when threshold met
- OPEN → reject with 503
- OPEN → HALF-OPEN after cooldown
- HALF-OPEN → CLOSED after 2 successes
- Fail-open if Redis unavailable

**Acceptance Criteria**:
- Healthy backend → requests forwarded
- Backend failing → circuit opens after threshold
- Circuit open → instant 503
- Backend recovers → circuit closes
- Redis down → requests allowed (fail-open)
- **Cross-instance**: Circuit state shared across gateway instances

**Dependencies**: Task 6 (Docker Compose with Redis)

---

## Task 10: Observability

**Goal**: Prometheus metrics + OpenTelemetry tracing

**Deliverables**:
- `gateway/observability/metrics.go`
- `gateway/observability/tracing.go`
- `gateway/middleware/metrics.go`
- Updated `prometheus/prometheus.yml`

**Metrics**:
- `gateway_requests_total` (counter)
- `gateway_request_duration_seconds` (histogram)
- `gateway_rate_limit_rejections_total` (counter)
- `gateway_circuit_breaker_state` (gauge)

**Tracing**:
- Span for each request
- Propagate trace context to backends

**Acceptance Criteria**:
- `/metrics` returns Prometheus format
- Prometheus scrapes gateway metrics
- Traces appear (stdout or collector)

**Dependencies**: Task 5 (routing works)

---

## Task 11: Integration & E2E Tests

**Goal**: Comprehensive test suite for the full system

**Deliverables**:
- `scripts/e2e/test_routing.sh`
- `scripts/e2e/test_auth.sh`
- `scripts/e2e/test_ratelimit.sh`
- `scripts/e2e/test_circuitbreaker.sh`
- `scripts/e2e/run_all.sh`

**E2E Test: Routing**
```bash
# Test all three backend services are reachable
curl http://localhost:5000/service-a/hello  # → Python response
curl http://localhost:5000/service-b/hello  # → Java response
curl http://localhost:5000/service-c/hello  # → Node response
curl http://localhost:5000/unknown/path     # → 404
```

**E2E Test: Authentication**
```bash
# Without token
curl http://localhost:5000/service-a/hello  # → 401

# With valid token
curl -H "Authorization: Bearer $TOKEN" http://localhost:5000/service-a/hello  # → 200

# With expired token
curl -H "Authorization: Bearer $EXPIRED" http://localhost:5000/service-a/hello  # → 401
```

**E2E Test: Rate Limiting (Cross-Instance)**
```bash
# Send 6 requests to each instance (limit=10)
for i in {1..6}; do curl -H "Auth..." http://localhost:5000/service-a/hello; done
for i in {1..6}; do curl -H "Auth..." http://localhost:5001/service-a/hello; done
# Expect: 10 succeed, 2 get 429
```

**E2E Test: Circuit Breaker (Cross-Instance)**
```bash
# Stop Service A
docker stop service-a

# Trigger failures on instance 1
for i in {1..5}; do curl -H "Auth..." http://localhost:5000/service-a/hello; done

# Verify circuit open on instance 2
curl -H "Auth..." http://localhost:5001/service-a/hello  # → 503 instant

# Restart Service A
docker start service-a

# Wait for cooldown (30s), verify recovery
sleep 35
curl -H "Auth..." http://localhost:5000/service-a/hello  # → 200
```

**Acceptance Criteria**:
- All E2E tests pass
- Tests run in CI (optional)
- Tests are documented in README

**Dependencies**: Tasks 7-10 complete

---

## Task 12: README + Final Polish

**Goal**: Documentation and cleanup

**Deliverables**:
- `README.md` with:
  - Architecture diagram
  - How to run
  - How to test (unit, integration, E2E)
  - API examples with curl
  - Design decisions summary
  - Troubleshooting section

**Cleanup**:
- Remove debug logs
- Consistent error messages
- Code comments where needed

**Dependencies**: All other tasks complete

---

## Execution Order

**Phase 1 (Parallel)**: Tasks 1, 2, 3, 4
- All backend services + gateway skeleton
- No dependencies, can be done simultaneously

**Phase 2**: Task 5, then Task 6
- Routing + Docker Compose
- Need skeleton first, then connect everything

**Phase 3 (Parallel)**: Tasks 7, 8, 9, 10
- All middleware (with unit tests)
- Can be done in parallel once Docker Compose works

**Phase 4**: Task 11
- Integration & E2E tests
- Proves the "distributed" aspect works

**Phase 5**: Task 12
- Final documentation

---

## Suggested PR Order (if doing sequentially)

1. PR #1: Service A (Python)
2. PR #2: Service B (Java)
3. PR #3: Service C (Node)
4. PR #4: Gateway Skeleton
5. PR #5: Gateway Routing + Proxy
6. PR #6: Docker Compose
7. PR #7: JWT Auth + Unit Tests
8. PR #8: Rate Limiting + Unit Tests + Cross-Instance Tests
9. PR #9: Circuit Breaker + Unit Tests + Cross-Instance Tests
10. PR #10: Observability
11. PR #11: E2E Test Suite
12. PR #12: README + Polish

---

## Test Summary

| Test Type | Location | Purpose |
|-----------|----------|---------|
| Unit Tests | `*_test.go` files | Test individual functions in isolation |
| Integration Tests | `*_test.go` with build tag | Test with real dependencies (Redis) |
| E2E Tests | `scripts/e2e/` | Test full system end-to-end |

**Integration Tests** (run with `go test -tags=integration`):
- Rate limiter with real Redis
- Circuit breaker with real Redis
- JWT validator with real keys

**E2E Tests** (proves "distributed"):
- Rate Limiting: Shared counter across gateway instances
- Circuit Breaker: Shared circuit state across gateway instances

---

*Last Updated: 2026-02-01*
