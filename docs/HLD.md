# High-Level Design: Distributed API Gateway

## 1. Overview

A production-grade API Gateway built with Go that provides:
- Path-based routing to backend services
- JWT authentication at the edge
- Distributed rate limiting (Redis-backed)
- Circuit breaker for resilience
- Prometheus metrics & OpenTelemetry tracing

---

## 2. Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Router | chi or net/http |
| Rate Limit Store | Redis |
| Metrics | Prometheus |
| Tracing | OpenTelemetry |
| Containers | Docker + Docker Compose |

---

## 3. Architecture

```
┌──────────┐         ┌──────────────────────────────────────────────────┐
│          │         │              API GATEWAY CLUSTER                 │
│  Client  │────────▶│                                                  │
│          │         │   ┌─────────────────┐   ┌─────────────────┐      │
└──────────┘         │   │ Gateway :5000   │   │ Gateway :5001   │      │
                     │   │                 │   │                 │      │
                     │   │ 1. Route Match  │   │ 1. Route Match  │      │
                     │   │ 2. JWT Auth     │   │ 2. JWT Auth     │      │
                     │   │ 3. Rate Limit   │   │ 3. Rate Limit   │      │
                     │   │ 4. Circuit Break│   │ 4. Circuit Break│      │
                     │   │ 5. Proxy        │   │ 5. Proxy        │      │
                     │   └────────┬────────┘   └────────┬────────┘      │
                     │            │                     │               │
                     └────────────┼─────────────────────┼───────────────┘
                                  │                     │
                                  └──────────┬──────────┘
                                             │
                                             ▼
                                    ┌────────────────┐
                                    │     Redis      │
                                    │     :6379      │
                                    │                │
                                    │ • Rate Limits  │
                                    │ • Circuit State│
                                    └────────────────┘
                                             │
                      ┌──────────────────────┴──────────────────────┐
                      │                                             │
                      ▼                                             ▼
             ┌────────────────┐                            ┌────────────────┐
             │  Service A     │                            │  Service B     │
             │    :6000       │                            │    :6001       │
             └────────────────┘                            └────────────────┘
```

**Components:**

| Component | Port | Instances |
|-----------|------|-----------|
| API Gateway | 5000, 5001 | 2 |
| Service A | 6000 | 1 |
| Service B | 6001 | 1 |
| Redis | 6379 | 1 |
| Prometheus | 9090 | 1 |

---

## 4. Request Pipeline

```
Request ──▶ Route Match ──▶ JWT Auth ──▶ Rate Limit ──▶ Circuit Breaker ──▶ Proxy ──▶ Response
               │               │             │               │               │
               ▼               ▼             ▼               ▼               ▼
              404             401           429             503            502/504
```

Each stage can short-circuit and return immediately.

---

## 5. Routing

| Path Prefix | Target | Strip Prefix |
|-------------|--------|--------------|
| `/service-a` | `http://service-a:6000` | Yes |
| `/service-b` | `http://service-b:6001` | Yes |

Example: `GET /service-a/users` → Backend receives `GET /users`

---

## 6. Authentication

- **Algorithm**: RS256 (asymmetric key pair)
- **Token location**: `Authorization: Bearer <token>`
- **Claims used**: `sub` (user ID), `client_id` (for rate limiting), `exp`

**Key Distribution:**

| Component | Has | Can Do |
|-----------|-----|--------|
| Token Generator (CLI script) | Private key | Sign tokens |
| API Gateway | Public key | Verify tokens (cannot forge) |

```
┌────────────────────┐                      ┌────────────────────┐
│  Token Generator   │                      │    API Gateway     │
│  (scripts/)        │                      │                    │
│                    │                      │                    │
│  PRIVATE KEY       │                      │  PUBLIC KEY        │
│  • Signs tokens    │───── token ─────────▶│  • Verifies tokens │
│                    │                      │  • Cannot sign     │
└────────────────────┘                      └────────────────────┘
```

Bypass auth: `/health`, `/metrics`

---

## 7. Rate Limiting

| Setting | Value |
|---------|-------|
| Algorithm | Sliding window counter |
| Window | 60 seconds |
| Default limit | 100 req/min |
| Identity | JWT `client_id`, fallback to IP |
| Storage | Redis |

**Redis failure**: Fail-open (allow all requests)

---

## 8. Circuit Breaker

| Setting | Value |
|---------|-------|
| Storage | Redis (shared across instances) |
| Granularity | Per service |
| Failure types | 5xx, timeout, connection error |
| Window | 60 seconds |
| Open condition | ≥5 failures AND >50% failure rate |
| Cooldown | 30 seconds |
| Close condition | 2 consecutive successes |

### State Machine

```
                       ≥5 failures AND >50% rate
              ┌───────────────────────────────────────────────────────────┐
              │                                                           │
              ▼                                                           │
        ┌──────────┐    30s cooldown    ┌───────────┐  2 consecutive  ┌──────────┐
        │          │      expires       │           │    successes    │          │
        │   OPEN   │ ─────────────────▶ │ HALF-OPEN │ ──────────────▶ │  CLOSED  │
        │          │                    │           │                 │          │
        └──────────┘                    └───────────┘                 └──────────┘
              ▲                               │
              │                               │
              │           any failure         │
              └───────────────────────────────┘
```

**Transitions:**
- **CLOSED → OPEN**: When ≥5 failures AND >50% failure rate in 60s window
- **OPEN → HALF-OPEN**: After 30s cooldown expires
- **HALF-OPEN → CLOSED**: After 2 consecutive successful requests
- **HALF-OPEN → OPEN**: On any failure (resets cooldown timer)

**Redis failure**: Bypass circuit breaker (allow requests)

---

## 9. Timeouts

| Timeout | Value |
|---------|-------|
| Backend connect | 1 second |
| Backend read | 5 seconds |
| Graceful shutdown | 30 seconds |

**Retries**: None (client's responsibility)

---

## 10. Observability

**Metrics (Prometheus)**:
- `gateway_requests_total` — request count by method, status, service
- `gateway_request_duration_seconds` — latency histogram
- `gateway_rate_limit_rejections_total` — rate limit 429s
- `gateway_circuit_breaker_state` — circuit state gauge

**Tracing**: OpenTelemetry spans propagated to backends

**Logging**: Structured JSON with request_id correlation

---

## 11. Error Responses

| Code | Status | Cause |
|------|--------|-------|
| `NOT_FOUND` | 404 | Route not matched |
| `UNAUTHORIZED` | 401 | JWT invalid |
| `RATE_LIMIT_EXCEEDED` | 429 | Over quota |
| `CIRCUIT_OPEN` | 503 | Backend unhealthy |
| `GATEWAY_TIMEOUT` | 504 | Backend timeout |
| `BAD_GATEWAY` | 502 | Backend unreachable |
| `INTERNAL_ERROR` | 500 | Unexpected gateway error |

---

## 12. Deployment

All components run via Docker Compose:

```yaml
services:
  gateway-1:    # Port 5000
  gateway-2:    # Port 5001  
  service-a:    # Port 6000
  service-b:    # Port 6001
  redis:        # Port 6379
  prometheus:   # Port 9090
```

Health checks: `/health` (liveness), `/health/ready` (includes Redis)

---

*Last Updated: 2026-02-01*
