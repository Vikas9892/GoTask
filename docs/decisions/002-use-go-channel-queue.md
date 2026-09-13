# ADR 002: Use In-Process Bounded Go Channels as Job Queue

## Problem
Many projects immediately introduce external message brokers (Redis, RabbitMQ, Kafka, SQS) which introduce significant operational overhead, external connection states, serialization complexity, and third-party dependencies, obscuring core Go concurrency concepts.

## Decision
Use bounded buffered Go channels (`chan *model.Job`) inside the Worker Service as the internal dispatch queue, with PostgreSQL as the durable backing store.

## Why
- **Core Go Concurrency**: Provides a real-world demonstration of goroutines, buffered channels, `select`, bounded queue flow control, and channel closure.
- **High Throughput & Low Latency**: In-memory channel dispatch operates at microsecond latencies (>700,000 ops/sec) with zero network overhead.
- **Built-In Backpressure**: When the bounded buffer fills, `Enqueue` returns `ErrQueueFull` immediately instead of consuming unlimited memory.
- **Zero Third-Party Queue Infrastructure**: No Redis, RabbitMQ, or Kafka required.

## Tradeoffs
- In-memory channels do not survive process restarts or power failure. This is mitigated by GoTask's startup recovery mechanism, which reloads pending and orphaned jobs from PostgreSQL upon service boot.
