# ADR 003: Use PostgreSQL as Durable Source of Truth

## Problem
In-memory channels are transient and lost when processes crash or restart. The system needs ACID-compliant durable storage to track job definitions, statuses, retry counts, and execution attempt logs across process restarts.

## Decision
Use PostgreSQL with `pgxpool` as the single durable source of truth for all job data (`jobs` and `job_attempts` tables).

## Why
- **ACID Guarantees**: State transitions (`pending` -> `processing` -> `completed`/`failed`) are persisted with transaction safety.
- **Relational Integrity**: Foreign key constraints tie attempt logs directly to their parent jobs (`ON DELETE CASCADE`).
- **Standardized Client**: `jackc/pgx/v5` provides high-performance connection pooling with native context cancellation support.
- **Crash Resilience**: Allows the Worker Service to recover cleanly on startup without losing any user jobs.

## Tradeoffs
- Database updates require network roundtrips for state transitions. This is kept performant using focused queries, indexed columns (`status`, `created_at`, `job_id`), and connection pooling with 25 connections.
