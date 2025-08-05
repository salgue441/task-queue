package storage

import (
	"context"
	"database/sql"
	"task-queue/pkg/errors"
	"task-queue/pkg/logger"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// JobRepository handles job persistence
type JobRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

// NewJobRepository creates a new job repository
func NewJobRepository(db *sqlx.DB, log logger.Logger) *JobRepository {
	return &JobRepository{
		db:     db,
		logger: log.Named("job-repo"),
	}
}

// Create inserts a new job into the database
func (r *JobRepository) Create(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (
			id, type, payload, status, priority, max_retries,
			retry_count, created_at, updated_at, scheduled_at,
			metadata
		) VALUES (
			:id, :type, :payload, :status, :priority, :max_retries,
			:retry_count, :created_at, :updated_at, :scheduled_at,
			:metadata
		)`

	_, err := r.db.NamedExecContext(ctx, query, job)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return errors.Newf("job with ID %s already exists", job.ID).
					WithCode(errors.CodeAlreadyExists)
			}
		}

		return errors.Wrap(err, "failed to create job").
			WithCode(errors.CodeDatabase)
	}

	r.logger.Debug("job created", "job_id", job.ID, "type", job.Type)
	return nil
}

// Get retrieves a job by ID
func (r *JobRepository) Get(ctx context.Context, id uuid.UUID) (*models.Job, error) {
	var job models.Job
	query := `SELECT * FROM jobs WHERE id = $1`
	err := r.db.GetContext(ctx, &job, query, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.Newf("job %s not found", id).
				WithCode(errors.CodeNotFound)
		}

		return nil, errors.Wrap(err, "failed to get job").
			WithCode(errors.CodeDatabase)
	}

	return &job, nil
}

