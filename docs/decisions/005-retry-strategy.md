# ADR 005: Exponential Backoff and Attempt Tracking

## Problem
Transient failures (e.g. temporary network blips or external service outages) will cause jobs to fail if not retried. However, retrying immediately (busy spinning) hammers failing dependencies and wastes CPU cycles.

## Decision
Implement configurable retries up to `MAX_RETRIES` with exponential backoff (`1s`, `2s`, `4s`, `8s`, up to a maximum cap of `30s`), recording each execution attempt in a dedicated `job_attempts` relational table.

## Why
- **Thundering Herd Avoidance**: Exponential backoff gives transient external service outages time to recover before subsequent retries.
- **Context-Aware Waiting**: Uses `time.NewTimer` with `select` listening to `ctx.Done()`, so shutdown signals interrupt retry delays immediately without waiting for timers to expire.
- **Auditability**: The `job_attempts` table records full timestamps, error messages, and status for every execution attempt.

## Tradeoffs
- Re-enqueuing directly into the in-process channel means in-memory retry order is maintained without an external distributed scheduler.
