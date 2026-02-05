# Distributed API Gateway

![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?logo=redis&logoColor=white)
![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?logo=prometheus&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green.svg)

A high-performance API gateway built in Go for distributed deployments. Features JWT authentication, Redis-backed rate limiting, circuit breaker pattern, and Prometheus observability — all with shared state across multiple instances.

📖 **[High-Level Design](docs/HLD.md)** · **[Low-Level Design](docs/LLD.md)** · **[Task Breakdown](docs/TASKS.md)**

## Features

| Feature | Description |
|---------|-------------|
| **Path-based Routing** | Route requests to backends based on URL prefix |
| **JWT Authentication** | RS256 signature validation with expiry checks |
| **Rate Limiting** | Sliding window algorithm, shared via Redis |
| **Circuit Breaker** | Protects backends from cascade failures |
| **Prometheus Metrics** | Request counts, latency, circuit breaker state |
| **Pipeline Visualizer** | Real-time UI showing request flow through middleware |

## Quick Start

```bash
# Generate JWT keys
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -in keys/private.pem -pubout -out keys/public.pem

# Start everything
docker compose up -d

# Test it
TOKEN=$(go run scripts/generate_jwt.go -sub testuser -client_id app -exp 1h)
curl -H "Authorization: Bearer $TOKEN" http://localhost:5000/service-a/hello

# Open Pipeline Visualizer
open http://localhost:3000
```

## Testing

```bash
# Unit tests
cd gateway && go test ./... -v

# E2E tests (requires running containers)
./scripts/e2e/run_all.sh
```

## License

This project is licensed under the [MIT License](LICENSE).

---

<div align="center">
  <b>Built with ❤️ by Rahul Kumar</b><br>
  <a href="https://rahulrj1.github.io/portfolio-revamped/">Portfolio</a> • <a href="https://github.com/rahulrj1">GitHub</a> • <a href="https://www.linkedin.com/in/rahul-rj/">LinkedIn</a><br>
  <br>
  <small>Assisted by <a href="https://cursor.com">Cursor AI</a> 🤖</small>
</div>
