# ADR 004: Worker Pool Concurrency Model

## Problem
Spawning an unconstrained number of goroutines (one per incoming job) can lead to resource exhaustion (out-of-memory errors, database connection pool exhaustion, or CPU thread starvation) under high traffic.

## Decision
Implement a fixed-size, configurable worker pool (`WORKER_COUNT`) where a defined number of worker goroutines consume from a single shared receive channel (`<-chan *model.Job`).

## Why
- **Predictable Resource Consumption**: CPU, RAM, and database connections remain bounded and stable even under heavy job spikes.
- **Natural Load Balancing**: Go's runtime scheduler automatically distributes jobs from the channel buffer to whichever worker goroutine becomes idle first.
- **Clean Lifecycle Control**: `sync.WaitGroup` and `context.Context` allow deterministic startup, execution, and shutdown.

## Tradeoffs
- If all worker goroutines are executing long-running jobs, newly queued jobs must wait in the channel buffer. This is mitigated by job timeouts (`JOB_TIMEOUT`) and configurable worker counts.
