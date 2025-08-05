package queue

import (
	"context"
	"time"

	"task-queue/internal/models"

	"github.com/google/uuid"
)

// Queue defines the interface for job queue operations
type Queue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *models.Job) error

	// EnqueueBatch adds multiple jobs to the queue
	EnqueueBatch(ctx context.Context, jobs []*models.Job) error

	// Dequeue retrieves the next job from the queue
	Dequeue(ctx context.Context) (*models.Job, error)

	// DequeueBatch retrieves multiple jobs from the queue
	DequeueBatch(ctx context.Context, limit int) ([]*models.Job, error)

	// Ack acknowledges successful job processing
	Ack(ctx context.Context, jobID uuid.UUID) error

	// Nack returns a job to the queue for reprocessing
	Nack(ctx context.Context, jobID uuid.UUID, reason string) error

	// Delete removes a job from the queue
	Delete(ctx context.Context, jobID uuid.UUID) error

	// Extend extends the visibility timeout for a job
	Extend(ctx context.Context, jobID uuid.UUID, duration time.Duration) error

	// Size returns the number of jobs in the queue
	Size(ctx context.Context) (int64, error)

	// Clear removes all jobs from the queue
	Clear(ctx context.Context) error

	// Close closes the queue connection
	Close() error

	// Stats returns queue statistics
	Stats(ctx context.Context) (*QueueStats, error)
}

// QueueStats represents queue statistics
type QueueStats struct {
	Name            string        `json:"name"`
	Size            int64         `json:"size"`
	Processing      int64         `json:"processing"`
	Delayed         int64         `json:"delayed"`
	Failed          int64         `json:"failed"`
	DeadLetter      int64         `json:"dead_letter"`
	EnqueueRate     float64       `json:"enqueue_rate"`
	DequeueRate     float64       `json:"dequeue_rate"`
	ProcessingTime  time.Duration `json:"avg_processing_time"`
	LastEnqueueTime *time.Time    `json:"last_enqueue_time,omitempty"`
	LastDequeueTime *time.Time    `json:"last_dequeue_time,omitempty"`
}

// Config represents queue configuration
type Config struct {
	Name              string        `json:"name" yaml:"name"`
	MaxSize           int64         `json:"max_size" yaml:"max_size"`
	VisibilityTimeout time.Duration `json:"visibility_timeout" yaml:"visibility_timeout"`
	RetentionPeriod   time.Duration `json:"retention_period" yaml:"retention_period"`
	MaxRetries        int           `json:"max_retries" yaml:"max_retries"`
	DeadLetterQueue   string        `json:"dead_letter_queue" yaml:"dead_letter_queue"`
	PollInterval      time.Duration `json:"poll_interval" yaml:"poll_interval"`
	BatchSize         int           `json:"batch_size" yaml:"batch_size"`
}

// DefaultConfig returns default queue configuration
func DefaultConfig() Config {
	return Config{
		Name:              "default",
		MaxSize:           10000,
		VisibilityTimeout: 30 * time.Minute,
		RetentionPeriod:   7 * 24 * time.Hour,
		MaxRetries:        3,
		DeadLetterQueue:   "dead_letter",
		PollInterval:      1 * time.Second,
		BatchSize:         10,
	}
}

// Priority queue names based on job priority
const (
	QueuePriorityCritical = "critical"
	QueuePriorityHigh     = "high"
	QueuePriorityNormal   = "normal"
	QueuePriorityLow      = "low"
)

// GetQueueName returns the queue name based on priority
func GetQueueName(priority models.JobPriority) string {
	switch priority {
	case models.JobPriorityCritical:
		return QueuePriorityCritical

	case models.JobPriorityHigh:
		return QueuePriorityHigh

	case models.JobPriorityLow:
		return QueuePriorityLow

	default:
		return QueuePriorityNormal
	}
}
