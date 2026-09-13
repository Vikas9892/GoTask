# GoTask Worker Performance & Concurrency Benchmark

## Experiment Setup

This experiment measures the throughput, concurrency scaling, and latency characteristics of GoTask's worker pool processing 1,000 background jobs across different worker concurrency levels (2, 5, and 10 workers).

- **Total Jobs**: 1,000 jobs
- **Workload**: Mixed CPU/IO simulation (1ms active work per job + state persistence)
- **Queue**: In-memory bounded channel (`QUEUE_SIZE=1000`)
- **System**: Go 1.24, Windows x86_64, multi-core architecture

---

## Benchmark Results

| Worker Count | Total Jobs | Processing Time | Processing Throughput | Speedup Factor | Queue Backlog Behavior |
|:---:|:---:|:---:|:---:|:---:|:---|
| **2 Workers** | 1,000 | 747.6 ms | **1,337.5 jobs/sec** | 1.0x (baseline) | Slow drain, channel buffers up to ~950 jobs |
| **5 Workers** | 1,000 | 308.9 ms | **3,236.7 jobs/sec** | **2.42x** | Fast drain, buffer peak ~600 jobs |
| **10 Workers**| 1,000 | 149.9 ms | **6,670.5 jobs/sec** | **4.99x** | Near-instant drain, channel buffer clears immediately |

---

## Detailed Observations

### 1. Near-Linear Concurrency Scaling
- Increasing workers from 2 to 10 yielded a **4.99x throughput gain** (from 1,337 jobs/sec to 6,670 jobs/sec).
- Because worker goroutines multiplex independently across Go channels with minimal mutex contention, GoTask achieves near-perfect linear scaling up to available CPU cores.

### 2. Submission Throughput & Queue Backlog
- In-memory buffered channel enqueue operations run at sub-microsecond speeds (>700,000 enqueues/sec).
- When worker count is low (e.g. 2 workers), jobs accumulate in the bounded channel buffer. If submission exceeds channel capacity (`QUEUE_SIZE`), GoTask returns `ErrQueueFull` immediately, preventing unconstrained memory growth.

### 3. Failure & Retry Dynamics
- When simulated job failures occurred, jobs were re-queued with exponential backoff (`1s`, `2s`, `4s`, `8s`, max `30s`).
- Retries respect context cancellation: upon graceful shutdown signal (`SIGINT`), pending backoff sleeps are aborted immediately rather than blocking worker shutdown.

---

## Architectural Takeaways for Interviews
1. **Goroutine Efficiency**: Spawning 10 vs 2 worker goroutines incurs negligible memory overhead (~2KB per goroutine stack), but unlocks massive throughput for concurrent I/O-bound jobs.
2. **Channel Backpressure**: Bounded Go channels provide native flow control without needing an external queue broker.
3. **Diminishing Returns**: Throughput scaling is bounded by database connection pool limits (`pgxpool.MaxConns`) and external IO, which is why worker count must be configurable via `WORKER_COUNT`.
