// Package models defines the core data structures and database schema
// for the job queue system.
//
// The package provides:
//   - Database tables and relationships for job management
//   - Enum types for job status and priority
//   - Event logging for job lifecycle tracking
//   - Database functions and triggers for automatic job processing
//
// Core Models:
//   - Job: Represents a work unit with status, priority, retry logic, and timestamps
//   - JobEvent: Audit trail of all state changes and processing events
//
// Database Features:
//   - Automatic timestamp management (created_at, updated_at)
//   - Comprehensive indexing for query performance
//   - Job state transition auditing
//   - Built-in functions for job processing:
//   - get_next_job(): Claim next available jobs
//   - retry_failed_jobs(): Automatically retry failed jobs
//   - update_updated_at(): Maintain modification timestamps
//   - log_job_event(): Track all job state changes
//
// Usage:
// Typically used by repository and service layers to interact with
// the job queue persistence layer. The SQL schema is designed for:
//   - High throughput job processing
//   - Reliable state tracking
//   - Operational visibility
//   - Horizontal scaling
//
// The schema supports:
//   - Priority-based job scheduling
//   - Configurable retry logic
//   - Worker assignment tracking
//   - Result and error storage
//   - Scheduled job execution
package models
