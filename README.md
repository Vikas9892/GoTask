# GoTask

A lightweight concurrent background-job processing system built with **Go**, **PostgreSQL**, **React**, **Vite**, and **Tailwind CSS**.

GoTask is intentionally engineered to demonstrate **core Go backend architecture and concurrency concepts** (goroutines, bounded channels, `select`, `context.Context`, `sync.WaitGroup`, and graceful shutdown) without relying on heavy external message brokers such as Redis, RabbitMQ, or Kafka.

---

## Architecture Diagram

```mermaid
flowchart TD
    subgraph Frontend ["Frontend (React + Vite + Tailwind CSS)"]
        UI["Web Dashboard"]
    end

    subgraph API ["API Service (Go net/http)"]
        Router["HTTP Router & Middleware"]
        APISvc["Job Service"]
        APIRepo["Postgres Job Repo"]
    end

    subgraph DB ["PostgreSQL (Durable Storage)"]
        JobsTable[("jobs Table\n(status: pending, processing, completed, failed)")]
        AttemptsTable[("job_attempts Table\n(execution logs & errors)")]
    end

    subgraph Worker ["Worker Service (Go Runtime)"]
        Dispatcher["Startup Recovery & Poller"]
        Queue["Bounded Go Channel\nchan *model.Job (Buffer: 100)"]
        Pool["Worker Pool\n(N Goroutines)"]
        Executors["Job Executors\n(email, webhook, report)"]
    end

    UI -->|REST API (JSON)| Router
    Router --> APISvc
    APISvc --> APIRepo
    APIRepo -->|INSERT / SELECT| JobsTable

    JobsTable -.->|Startup Recovery / Poll Pending| Dispatcher
    Dispatcher -->|Enqueue| Queue
    Queue -->|Consume Work| Pool
    Pool --> Executors
    Pool -->|UPDATE Status / Record Attempts| DB
```

---

## Core System Architecture

GoTask is structured as two decoupled Go application processes:

1. **API Service** (`services/api`):
   - Handles client HTTP REST traffic using Go standard library `net/http`.
   - Validates incoming payloads and persists jobs to PostgreSQL with initial status `pending`.
   - Exposes job attempt history, health checks, and Prometheus metrics.

2. **Worker Service** (`services/worker`):
   - Manages the in-process **bounded Go channel queue** and a configurable **worker pool** of goroutines.
   - At startup, recovers any pending or abandoned jobs from PostgreSQL.
   - Dispatches jobs across worker goroutines, updates state in PostgreSQL, handles exponential retry backoffs, and logs execution attempts.

3. **PostgreSQL as the Durable Source of Truth**:
   - The Go channel is solely an **in-process concurrency dispatch buffer**; it is not durable.
   - PostgreSQL provides ACID durability for all job states and attempt logs across restarts or node reboots.

---

## Key Go Concepts in Action

| Go Concept | Implementation in GoTask |
|---|---|
| **Goroutines** | Fixed worker pool (`WORKER_COUNT`) concurrently pulling and executing background jobs. |
| **Buffered Channels** | Bounded in-process job queue (`make(chan *model.Job, capacity)`) providing native flow control and backpressure. |
| **`select` & Non-Blocking Sends** | Queue enqueue returns `ErrQueueFull` immediately when buffer capacity is reached rather than blocking indefinitely. |
| **`context.Context`** | Propagates request cancellation, enforces per-job timeouts (`JOB_TIMEOUT`), and interrupts backoff sleep during shutdown. |
| **`sync.WaitGroup`** | Coordinates worker pool lifecycle, ensuring in-flight jobs complete before process exit. |
| **Graceful Shutdown** | Intercepts `SIGINT`/`SIGTERM`, closes the channel, allows workers to finish, closes `pgxpool`, and exits cleanly. |

---

## Job State Machine

```
         ┌───────────────┐
         │    Pending    │ ◄─── (Retry / Startup Recovery)
         └───────┬───────┘
                 │ (Worker picks up job)
                 ▼
         ┌───────────────┐
         │  Processing   │
         └──┬─────────┬──┘
            │         │
(Success)   │         │ (Failure)
            ▼         ▼
┌─────────────┐     ┌───────────────────────┐
│  Completed  │     │ Attempts < MaxAttempts?│
└─────────────┘     └───────┬───────────────┘
                            │
               ┌────────────┴────────────┐
          Yes  │                         │ No
               ▼                         ▼
        ┌─────────────┐           ┌─────────────┐
        │   Pending   │           │   Failed    │
        │  (Backoff)  │           └─────────────┘
        └─────────────┘
```

---

## REST API Specification

| Method | Path | Description | Status Code |
|---|---|---|---|
| `POST` | `/api/jobs` | Submit a new background job (`email`, `webhook`, `report`) | `201 Created` |
| `GET` | `/api/jobs` | List recent jobs with pagination (`?limit=20&offset=0`) | `200 OK` |
| `GET` | `/api/jobs/{id}` | Get details and status for a specific job | `200 OK` / `404 Not Found` |
| `GET` | `/api/jobs/{id}/attempts`| Get execution attempt history and errors for a job | `200 OK` / `404 Not Found` |
| `DELETE` | `/api/jobs/{id}` | Delete a job and its attempt history | `204 No Content` |
| `GET` | `/health` | Liveness check (process running) | `200 OK` |
| `GET` | `/ready` | Readiness check (verifies PostgreSQL connectivity) | `200 OK` / `503 Service Unavailable` |
| `GET` | `/metrics` | Prometheus metrics scrape endpoint | `200 OK` |

### Sample Job Submission Payload

```json
POST /api/jobs
Content-Type: application/json

{
  "type": "email",
  "payload": {
    "to": "user@example.com",
    "subject": "Weekly Digest"
  }
}
```

---

## Configuration Reference

Configuration is managed centrally via environment variables with validated defaults:

| Variable | Service | Default | Description |
|---|---|---|---|
| `PORT` | API | `8080` | Port for API HTTP server |
| `WORKER_PORT` | Worker | `8081` | Port for Worker health & metrics server |
| `DATABASE_URL` | API & Worker | `postgres://postgres:postgres@localhost:5432/gotask?sslmode=disable` | PostgreSQL connection string |
| `QUEUE_SIZE` | API & Worker | `100` | Maximum capacity of the bounded Go channel queue |
| `WORKER_COUNT` | Worker | `5` | Number of concurrent worker goroutines |
| `MAX_RETRIES` | Worker | `3` | Maximum retry attempts before permanent failure |
| `JOB_TIMEOUT` | Worker | `30s` | Maximum execution duration before timeout cancellation |
| `SHUTDOWN_TIMEOUT`| API & Worker | `15s` | Maximum duration allowed for graceful shutdown |

---

## Quickstart & Local Development

### 1. Run with Docker Compose (Recommended)

Start the complete stack (API, Worker, PostgreSQL, and React Frontend):

```bash
docker compose up --build
```

- **Frontend Dashboard**: `http://localhost:3000`
- **API Service**: `http://localhost:8080`
- **Worker Metrics/Health**: `http://localhost:8081`
- **PostgreSQL**: `localhost:5432`

### 2. Run Locally from Source

#### Prerequisites
- Go 1.22+
- Node.js 20+
- PostgreSQL running locally with database `gotask`

#### Run Migrations
```bash
psql -U postgres -d gotask -f migrations/001_create_jobs.sql
psql -U postgres -d gotask -f migrations/002_create_job_attempts.sql
```

#### Run Services
```bash
# Terminal 1: API Service
go run ./services/api

# Terminal 2: Worker Service
go run ./services/worker

# Terminal 3: React Frontend
cd frontend
npm install
npm run dev
```

---

## Testing & Quality Assurance

```bash
# Format code
go fmt ./...

# Vet codebase
go vet ./...

# Run all unit and integration tests
go test -v ./...

# Run tests with race detector (requires 64-bit GCC/MinGW)
go test -race ./...
```

---

## Performance & Load Benchmarks

Tested with 1,000 background jobs processed through the bounded queue:

| Worker Count | Total Jobs | Processing Duration | Throughput | Speedup |
|:---:|:---:|:---:|:---:|:---:|
| **2 Workers** | 1,000 | 747.6 ms | 1,337.5 jobs/sec | 1.0x (baseline) |
| **5 Workers** | 1,000 | 308.9 ms | 3,236.7 jobs/sec | **2.42x** |
| **10 Workers** | 1,000 | 149.9 ms | 6,670.5 jobs/sec | **4.99x** |

Detailed benchmark methodology and queue backlog measurements are documented in [`docs/performance.md`](docs/performance.md).

---

## Architecture Decision Records (ADRs)

Key architectural decisions are documented under [`docs/decisions/`](docs/decisions/):
- [`001-use-microservices.md`](docs/decisions/001-use-microservices.md): Decoupling HTTP traffic from job execution.
- [`002-use-go-channel-queue.md`](docs/decisions/002-use-go-channel-queue.md): Using Go channels as bounded dispatch queues.
- [`003-use-postgresql.md`](docs/decisions/003-use-postgresql.md): ACID state transitions and attempt history.
- [`004-worker-pool.md`](docs/decisions/004-worker-pool.md): Goroutine resource containment.
- [`005-retry-strategy.md`](docs/decisions/005-retry-strategy.md): Exponential backoff and context-aware cancellation.
- [`006-graceful-shutdown.md`](docs/decisions/006-graceful-shutdown.md): Signal handling and clean in-flight job drain.

---

## Design Tradeoffs & Limitations

1. **In-Memory Channel Queue**: Bounded channels provide low-latency work distribution, but are transient. Startup recovery from PostgreSQL ensures jobs are never lost.
2. **Polling Frequency**: In multi-process mode without external pub/sub, the Worker polls PostgreSQL periodically (2s interval) for newly submitted jobs.
3. **Single Consumer Scaling**: If scaling to multiple independent worker nodes, jobs should use `SELECT ... FOR UPDATE SKIP LOCKED` to prevent duplicate job pickups across different worker processes.
