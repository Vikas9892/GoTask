# ADR 001: Separate API and Worker Microservices

## Problem
In background job systems, handling user-facing HTTP traffic has vastly different resource characteristics, failure domains, and scaling patterns than executing asynchronous background jobs (which may consume significant CPU, memory, or external network I/O).

## Decision
Separate GoTask into two distinct Go application binaries and Docker containers:
1. **API Service**: Accepts HTTP requests, validates payloads, writes jobs to PostgreSQL, and queries job attempt history.
2. **Worker Service**: Polls jobs from PostgreSQL into its bounded in-memory Go channel queue, executes jobs via worker goroutines, and updates state.

## Why
- **Fault Isolation**: A spike in background job processing or a catastrophic worker crash never impacts the availability or latency of the user-facing REST API.
- **Independent Scalability**: In real-world deployments, worker pods can be scaled independently of API replicas based on job queue backlog.
- **Clean Separation of Concerns**: Web routing and HTTP concerns are completely decoupled from worker pool lifecycle, timeouts, and retry logic.

## Tradeoffs
- Go channels cannot be shared directly across separate OS processes. GoTask resolves this by having PostgreSQL serve as the durable inbox table, while the Worker Service runs an internal dispatcher that feeds its in-process bounded Go channel queue.
