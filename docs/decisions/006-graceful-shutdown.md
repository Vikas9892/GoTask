# ADR 006: Signal-Driven Graceful Shutdown

## Problem
Abruptly terminating an asynchronous worker process leaves in-flight jobs in an ambiguous or corrupted state, severs active database transactions, and causes client HTTP requests to abruptly drop with connection reset errors.

## Decision
Implement coordinated graceful shutdown for both API and Worker services using `os/signal` (`SIGINT`, `SIGTERM`), `context.Context` cancellation, HTTP server `Shutdown()`, and `sync.WaitGroup.Wait()`.

## Why
- **Clean In-Flight Work Completion**: Active HTTP requests and active worker job executions are given up to `SHUTDOWN_TIMEOUT` (e.g. 15s) to complete cleanly.
- **Zero Resource Leaks**: Database pools and open network sockets are flushed and closed properly before the process exits.
- **Docker / Kubernetes Compatibility**: Responds gracefully to container stop commands (`docker stop` sends `SIGTERM` followed by `SIGKILL` if timeout expires).

## Tradeoffs
- Process termination takes slightly longer (up to the duration of the longest in-flight job or `SHUTDOWN_TIMEOUT`). This is an essential production tradeoff for data consistency.
