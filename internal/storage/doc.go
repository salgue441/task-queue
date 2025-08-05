// Package storage provides the data persistence layer for the task queue
// system. It abstracts database operations and provides repositories for
// managing jobs, job events, and related entities.
//
// The package uses PostgreSQL as the primary datastore with support for
// transactions, connection pooling, and prepared statements for optimal
// performance.
//
// Repository pattern example:
//
//	repo := storage.NewJobRepository(db, logger)
//
//	// Create a job
//	err := repo.Create(ctx, job)
//
//	// Find jobs by status
//	jobs, err := repo.FindByStatus(ctx, models.JobStatusPending)
//
//	// Update job status
//	err := repo.UpdateStatus(ctx, jobID, models.JobStatusCompleted)
package storage
