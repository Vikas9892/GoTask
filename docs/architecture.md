# GoTask Architecture

## Overview

GoTask is an asynchronous job processing system demonstrating core Go concurrency and clean backend architecture without heavy external messaging brokers (such as Redis, RabbitMQ, or Kafka).

```
                      +-------------------+
                      |   React Frontend  |
                      +---------+---------+
                                |
                                | HTTP REST
                                v
                      +-------------------+
                      |    API Service    |
                      |    (net/http)     |
                      +---------+---------+
                                |
                                | INSERT jobs (status: pending)
                                v
                      +-------------------+
                      |    PostgreSQL     | <----------------+
                      | (Durable Storage) |                  |
                      +---------+---------+                  |
                                |                            |
                                | Poll / Recovery            | UPDATE status
                                v                            | Record attempts
                      +-------------------+                  |
                      |  Worker Service   |                  |
                      |  - Poller Loop    |                  |
                      |  - chan *Job      |                  |
                      |  - Worker Pool    | -----------------+
                      +-------------------+
```

## The Inter-Process Queue Problem & Clean Solution

### The Challenge
Go channels are purely in-memory data structures residing in the virtual memory of a single OS process. They cannot be directly shared across different OS processes (API Service and Worker Service) without IPC or external brokers.

### The Solution
1. **Durable Persistence (API Service)**:
   - When a client sends `POST /api/jobs`, the API service validates the payload and writes the job to PostgreSQL with `status = 'pending'`.
   - The API immediately responds with HTTP 201 Created and the created Job object.

2. **Work Queue & Dispatch (Worker Service)**:
   - The Worker Service owns the in-process **bounded Go channel queue** (`Queue` with `chan *model.Job`) and the **worker pool** (`N` goroutines).
   - An internal dispatcher in the Worker Service fetches pending jobs from PostgreSQL and pushes them into the bounded Go channel.
   - If the bounded channel is full, backpressure occurs naturally; jobs remain pending in PostgreSQL until workers consume channel items.

3. **Crash Recovery**:
   - If the Worker Service stops or crashes, all jobs remain safely persisted in PostgreSQL.
   - On startup, the Worker Service scans for orphaned `processing` jobs and all `pending` jobs, reloading them into the channel buffer.

4. **Concurrency Inside the Worker**:
   - Worker pool goroutines consume from the `<-chan *model.Job`.
   - Each worker updates PostgreSQL status to `processing`, creates a `job_attempts` record, executes the appropriate handler (`email`, `webhook`, `report`), and transitions the job to `completed` or `failed`.
