# Low-Level Design: Distributed API Gateway

---

## 1. Project Structure

```
API_Gateway/
├── docker-compose.yml
├── README.md
├── keys/                        # RSA key pair (gitignored)
│   ├── private.pem
│   └── public.pem
│
├── gateway/                     # Go
│   ├── main.go
│   ├── go.mod
│   ├── Dockerfile
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── proxy/
│   └── pkg/
│
├── services/
│   ├── service-a/              # Python/Flask
│   │   ├── app.py
│   │   ├── requirements.txt
│   │   └── Dockerfile
│   ├── service-b/              # Java/Spring Boot
│   │   ├── src/
│   │   ├── pom.xml
│   │   └── Dockerfile
│   └── service-c/              # Node.js/Express
│       ├── index.js
│       ├── package.json
│       └── Dockerfile
│
├── prometheus/
│
└── scripts/
    └── generate_jwt.go
```

---

## 2. Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 5000 | Gateway listen port |
| `REDIS_ADDR` | localhost:6379 | Redis address |
| `REDIS_PASSWORD` | (empty) | Redis password |
| `JWT_PUBLIC_KEY_PATH` | (required) | Path to public.pem |
| `JWT_ISSUER` | (empty) | Expected issuer claim |
| `RATE_LIMIT_WINDOW` | 60s | Rate limit window |
| `RATE_LIMIT_DEFAULT` | 100 | Requests per window |
| `CIRCUIT_WINDOW` | 60s | Failure tracking window |
| `CIRCUIT_MIN_FAILURES` | 5 | Min failures to open |
| `CIRCUIT_FAILURE_THRESHOLD` | 0.5 | Failure rate to open |
| `CIRCUIT_COOLDOWN` | 30s | Time before half-open |
| `CIRCUIT_SUCCESS_THRESHOLD` | 2 | Successes to close |

### Routes File (YAML)

```yaml
routes:
  - path_prefix: "/service-a"
    target: "http://service-a:6000"
    strip_prefix: true
    timeout: 5s

  - path_prefix: "/service-b"
    target: "http://service-b:6001"
    strip_prefix: true
    timeout: 5s

  - path_prefix: "/service-c"
    target: "http://service-c:6002"
    strip_prefix: true
    timeout: 5s
```

---

## 3. JWT Validation

**Algorithm**: RS256

**Steps**:
1. Extract token from `Authorization: Bearer <token>`
2. Decode header, verify `alg` is RS256
3. Verify signature using public key
4. Check `exp` claim (reject if expired)
5. Check `iss` claim if configured
6. Extract `sub` → `X-User-ID`, `client_id` → rate limit key

**Skip auth for**: `/health`, `/metrics`

---

## 4. Rate Limiting

**Algorithm**: Sliding Window Counter

**Redis Keys**:
```
ratelimit:{client_id}:{window_start_timestamp}
TTL: 120 seconds (2× window)
```

**Logic**:
1. Calculate current window start: `floor(now / 60) * 60`
2. Get count from current window
3. Get count from previous window
4. Calculate weighted count:
   ```
   elapsed = now - window_start
   weight = 1 - (elapsed / window_size)
   count = (prev_count × weight) + current_count
   ```
5. If `count >= limit` → reject with 429
6. Else → increment current window, allow request

**Atomicity**: Use Lua script to make check-and-increment atomic.

---

## 5. Circuit Breaker

**Redis Keys**:

State:
```
circuit:{service}:state
Value: {state, opened_at, half_open_successes}
```

Window counters:
```
circuit:{service}:window:{timestamp}
Value: {total, failures}
TTL: 120 seconds
```

**State Machine Logic**:

| Current State | Condition | Action |
|---------------|-----------|--------|
| CLOSED | Request arrives | Allow, track result |
| CLOSED | failures ≥ 5 AND rate ≥ 50% | → OPEN |
| OPEN | Request arrives | Reject 503 |
| OPEN | 30s elapsed | → HALF-OPEN |
| HALF-OPEN | Request arrives | Allow (probe) |
| HALF-OPEN | Success | Increment success counter |
| HALF-OPEN | 2 consecutive successes | → CLOSED |
| HALF-OPEN | Any failure | → OPEN (reset timer) |

---

## 6. Proxy

**Request forwarding**:
1. Match route by path prefix
2. Strip prefix if configured
3. Copy method, headers, query params, body
4. Add headers: `X-Forwarded-For`, `X-User-ID`, `X-Client-ID`, `X-Request-ID`
5. Remove `Authorization` header
6. Forward with configured timeout
7. Stream response back to client

---

## 7. Error Response Format

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {}
  },
  "request_id": "uuid"
}
```

`details` is optional, used for extra context (e.g., `retry_after` for 429).

---

## 8. Key Generation

```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem
```

---

## 9. Test Scenarios

### Authentication

| Input | Expected |
|-------|----------|
| Valid token | 200, forwarded |
| Missing header | 401 |
| Expired token | 401 |
| Invalid signature | 401 |

### Rate Limiting

| Input | Expected |
|-------|----------|
| 50 requests | All 200 |
| 101 requests | 100×200, 1×429 |
| Redis down | All allowed (fail-open) |

### Circuit Breaker

| Input | Expected |
|-------|----------|
| 5 failures / 10 total | Circuit opens |
| 5 failures / 100 total | Stays closed (5% rate) |
| Request while open | 503 |
| After 30s cooldown | Half-open |
| 2 successes in half-open | Closes |
| Failure in half-open | Re-opens |

---

*Last Updated: 2026-02-01*
