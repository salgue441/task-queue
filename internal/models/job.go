package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusRetrying  JobStatus = "retrying"
	JobStatusDead      JobStatus = "dead"
)

// JobPriority represents the priority level of a job
type JobPriority int

const (
	JobPriorityLow      JobPriority = 0
	JobPriorityNormal   JobPriority = 1
	JobPriorityHigh     JobPriority = 2
	JobPriorityCritical JobPriority = 3
)

// Job represents a task in the queue system
type Job struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	Type        string          `json:"type" db:"type" validate:"required,min=1,max=100"`
	Payload     json.RawMessage `json:"payload" db:"payload"`
	Status      JobStatus       `json:"status" db:"status"`
	Priority    JobPriority     `json:"priority" db:"priority"`
	MaxRetries  int             `json:"max_retries" db:"max_retries"`
	RetryCount  int             `json:"retry_count" db:"retry_count"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	ScheduledAt *time.Time      `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty" db:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty" db:"completed_at"`
	Error       *string         `json:"error,omitempty" db:"error"`
	Result      json.RawMessage `json:"result,omitempty" db:"result"`
	WorkerID    *string         `json:"worker_id,omitempty" db:"worker_id"`
	Metadata    map[string]any  `json:"metadata,omitempty" db:"metadata"`
}

// NewJob creates a new job with default values
func NewJob(jobType string, payload json.RawMessage, priority JobPriority) *Job {
	now := time.Now().UTC()
	return &Job{
		ID:         uuid.New(),
		Type:       jobType,
		Payload:    payload,
		Status:     JobStatusPending,
		Priority:   priority,
		MaxRetries: 3,
		RetryCount: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
		Metadata:   make(map[string]any),
	}
}

// JobRequest represents a request to create a new job
type JobRequest struct {
	Type        string          `json:"type" validate:"required,min=1,max=100"`
	Payload     json.RawMessage `json:"payload" validate:"required"`
	Priority    JobPriority     `json:"priority,omitempty"`
	MaxRetries  *int            `json:"max_retries,omitempty" validate:"omitempty,min=0,max=10"`
	ScheduledAt *time.Time      `json:"scheduled_at,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
}

// JobResult represents the result of a completed job
type JobResult struct {
	JobID       uuid.UUID       `json:"job_id"`
	Status      JobStatus       `json:"status"`
	Result      json.RawMessage `json:"result,omitempty"`
	Error       *string         `json:"error,omitempty"`
	CompletedAt time.Time       `json:"completed_at"`
	Duration    time.Duration   `json:"duration"`
}
