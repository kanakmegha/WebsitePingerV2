# Pinger (GO) - Production-Grade Website Monitoring SaaS Platform

Pinger is a high-performance, multi-tenant website monitoring SaaS platform
designed to monitor HTTP Uptime, SSL Certificates, WHOIS Expiry, and DNS Records
(A, AAAA, MX, TXT, SPF, DMARC) across 100K+ monitors with strict request rate
limiting and distributed request coalescing.

## Key System Highlights

1. **No Data Caching**: API directly queries PostgreSQL for check audit history.
2. **Distributed Request Coalescing**: Go `singleflight` combined with Redis
   Pub/Sub locks ensures that even under massive tenant overlap (e.g. 50 tenants
   monitoring `google.com`), **only ONE outbound network query** is executed per
   cycle.
3. **Mandatory Rate Limiting**: Outbound checks per domain strictly observe
   minimum intervals:
   - **HTTP**: `>= 30s`
   - **DNS**: `>= 5 min`
   - **SSL**: `>= 1 hour`
   - **WHOIS**: `>= 24 hours`
4. **Idempotent Alerting Engine**: State-transition alert trigger with
   deduplication locks preventing alert storms.

---

## Tech Stack

- **Backend**: Go (1.22+)
- **ORM**: Ent (`entgo.io/ent`)
- **Database**: PostgreSQL
- **Queue & Distributed Locking**: Redis 7
- **Routing**: `go-chi/chi/v5`
- **Containerization**: Docker & Docker Compose

---

## Folder Structure

```
Pinger(GO)/
├── cmd/
│   ├── api/          # REST API server
│   ├── scheduler/    # Ticker enqueuing domain check jobs
│   └── worker/       # Scalable worker pool
├── internal/
│   ├── alert/        # Idempotent alerting engine
│   ├── checker/      # HTTP, SSL, WHOIS, DNS network inspection
│   ├── coalescer/    # Distributed singleflight + Redis PubSub coalescer
│   ├── ent/          # Generated Ent ORM client & schemas
│   ├── handler/      # REST API handlers
│   ├── middleware/   # JWT Auth & Multi-Tenant context middleware
│   ├── queue/        # Redis queue producer & consumer
│   └── ratelimit/    # Minimum fetch interval enforcer
├── deployments/      # Dockerfiles & docker-compose.yml
└── README.md
```

---

## Running Locally with Docker Compose

```bash
cd deployments
docker-compose up --build -d
```

Services started:

- `api`: `http://localhost:4002`
- `postgres`: `localhost:5432`
- `redis`: `localhost:6379`
- `scheduler`: Background ticker service
- `worker`: Concurrent worker pool (50 goroutines)

---

## API Documentation

### 1. Register User & Tenant

```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@org.com",
  "password": "securepassword"
}
```

### 2. Login

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@org.com",
  "password": "securepassword"
}
```

### 3. Create Monitor

```http
POST /api/monitors
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "name": "Production Google Monitor",
  "url": "https://google.com",
  "type": "http",
  "interval_seconds": 60,
  "timeout_seconds": 10
}
```

### 4. Fetch Check History (Direct PostgreSQL Query - No Cache)

```http
GET /api/monitors/{monitor_id}/checks
Authorization: Bearer <JWT_TOKEN>
```
