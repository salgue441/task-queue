package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"task-queue/internal/models"
	"task-queue/pkg/errors"
	"task-queue/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisQueue implements Queue interface using Redis
type RedisQueue struct {
	client    *redis.Client
	config    Config
	logger    logger.Logger
	keyPrefix string
}

// NewRedisQueue creates a new Redis-based queue
func NewRedisQueue(client *redis.Client, config Config, log logger.Logger) (*RedisQueue, error) {
	if client == nil {
		return nil, errors.New("redis client is required").
			WithCode(errors.CodeConfiguration)
	}

	if config.Name == "" {
		config.Name = "default"
	}

	return &RedisQueue{
		client:    client,
		config:    config,
		logger:    log.Named("redis-queue"),
		keyPrefix: fmt.Sprintf("queue:%s", config.Name),
	}, nil
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *models.Job) error {
	if job == nil {
		return errors.New("job is nil").WithCode(errors.CodeValidation)
	}

	data, err := json.Marshal(job)
	if err != nil {
		return errors.Wrap(err, "failed to marshal job").
			WithCode(errors.CodeSerialization)
	}

	queueKey := q.getQueueKey(job.Priority)
	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		score := float64(job.ScheduledAt.Unix())
		err = q.client.ZAdd(ctx, q.getDelayedKey(), redis.Z{
			Score:  score,
			Member: data,
		}).Err()
	} else {
		err = q.client.RPush(ctx, queueKey, data).Err()
	}

	if err != nil {
		return errors.Wrap(err, "failed to enqueue job").
			WithCode(errors.CodeInternal)
	}

	q.updateEnqueueStats(ctx)
	q.logger.Debug("job enqueued",
		"job_id", job.ID,
		"type", job.Type,
		"priority", job.Priority,
	)

	return nil
}

// EnqueueBatch adds multiple jobs to the queue
func (q *RedisQueue) EnqueueBatch(ctx context.Context, jobs []*models.Job) error {
	if len(jobs) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()

	for _, job := range jobs {
		data, err := json.Marshal(job)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal job %s", job.ID).
				WithCode(errors.CodeSerialization)
		}

		queueKey := q.getQueueKey(job.Priority)

		if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
			score := float64(job.ScheduledAt.Unix())
			pipe.ZAdd(ctx, q.getDelayedKey(), redis.Z{
				Score:  score,
				Member: data,
			})
		} else {
			pipe.RPush(ctx, queueKey, data)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to enqueue batch").WithCode(errors.CodeInternal)
	}

	q.updateEnqueueStats(ctx)
	q.logger.Debug("batch enqueued", "count", len(jobs))

	return nil
}

// Dequeue retrieves the next job from the queue
func (q *RedisQueue) Dequeue(ctx context.Context) (*models.Job, error) {
	if err := q.processScheduledJobs(ctx); err != nil {
		q.logger.Warn("failed to process scheduled jobs", "error", err)
	}

	priorities := []models.JobPriority{
		models.JobPriorityCritical,
		models.JobPriorityHigh,
		models.JobPriorityNormal,
		models.JobPriorityLow,
	}

	for _, priority := range priorities {
		queueKey := q.getQueueKey(priority)
		processingKey := q.getProcessingKey()
		result, err := q.client.BLMove(ctx,
			queueKey,
			processingKey,
			"LEFT",
			"RIGHT",
			100*time.Millisecond,
		).Result()

		if err == redis.Nil {
			continue
		}

		if err != nil {
			return nil, errors.Wrap(err, "failed to dequeue job").
				WithCode(errors.CodeInternal)
		}

		var job models.Job
		if err := json.Unmarshal([]byte(result), &job); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal job").
				WithCode(errors.CodeSerialization)
		}

		if err := q.setVisibilityTimeout(ctx, &job); err != nil {
			q.logger.Warn("failed to set visibility timeout",
				"job_id", job.ID,
				"error", err,
			)
		}

		q.updateDequeueStats(ctx)
		q.logger.Debug("job dequeued",
			"job_id", job.ID,
			"type", job.Type,
		)

		return &job, nil
	}

	return nil, nil
}

// DequeueBatch retrieves multiple jobs from the queue
func (q *RedisQueue) DequeueBatch(ctx context.Context, limit int) ([]*models.Job, error) {
	if limit <= 0 {
		limit = q.config.BatchSize
	}

	jobs := make([]*models.Job, 0, limit)
	for i := 0; i < limit; i++ {
		job, err := q.Dequeue(ctx)
		if err != nil {
			return jobs, err
		}

		if job == nil {
			break
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Ack acknowledges successful job processing
func (q *RedisQueue) Ack(ctx context.Context, jobID uuid.UUID) error {
	processingKey := q.getProcessingKey()
	jobs, err := q.client.LRange(ctx, processingKey, 0, -1).Result()
	if err != nil {
		return errors.Wrap(err, "failed to get processing jobs").WithCode(errors.CodeInternal)
	}

	for _, jobData := range jobs {
		var job models.Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		if job.ID == jobID {
			count, err := q.client.LRem(ctx, processingKey, 1, jobData).Result()
			if err != nil {
				return errors.Wrap(err, "failed to remove job from processing").
					WithCode(errors.CodeInternal)
			}

			if count > 0 {
				q.clearVisibilityTimeout(ctx, jobID)
				q.logger.Debug("job acknowledged", "job_id", jobID)

				return nil
			}
		}
	}

	return errors.Newf("job %s not found in processing queue", jobID).
		WithCode(errors.CodeNotFound)
}

// Nack returns a job to the queue for reprocessing
func (q *RedisQueue) Nack(ctx context.Context, jobID uuid.UUID, reason string) error {
	processingKey := q.getProcessingKey()
	jobs, err := q.client.LRange(ctx, processingKey, 0, -1).Result()
	if err != nil {
		return errors.Wrap(err, "failed to get processing jobs").
			WithCode(errors.CodeInternal)
	}

	for _, jobData := range jobs {
		var job models.Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		if job.ID == jobID {
			job.RetryCount++
			job.Error = &reason
			job.UpdatedAt = time.Now()

			if job.RetryCount >= job.MaxRetries {
				if err := q.moveToDeadLetter(ctx, &job); err != nil {
					return err
				}
			} else {
				backoffDuration := time.Duration(job.RetryCount) * time.Minute
				job.ScheduledAt = ptr(time.Now().Add(backoffDuration))

				if err := q.Enqueue(ctx, &job); err != nil {
					return err
				}
			}

			_, err := q.client.LRem(ctx, processingKey, 1, jobData).Result()
			if err != nil {
				return errors.Wrap(err, "failed to remove job from processing").
					WithCode(errors.CodeInternal)
			}

			q.logger.Debug("job nacked",
				"job_id", jobID,
				"retry_count", job.RetryCount,
				"reason", reason,
			)

			return nil
		}
	}

	return errors.Newf("job %s not found in processing queue", jobID).
		WithCode(errors.CodeNotFound)
}

// Delete removes a job from the queue
func (q *RedisQueue) Delete(ctx context.Context, jobID uuid.UUID) error {
	keys := []string{
		q.getQueueKey(models.JobPriorityLow),
		q.getQueueKey(models.JobPriorityNormal),
		q.getQueueKey(models.JobPriorityHigh),
		q.getQueueKey(models.JobPriorityCritical),
		q.getProcessingKey(),
		q.getDelayedKey(),
		q.getDeadLetterKey(),
	}

	for _, key := range keys {
		if key != q.getDelayedKey() {
			jobs, err := q.client.LRange(ctx, key, 0, -1).Result()
			if err != nil {
				continue
			}

			for _, jobData := range jobs {
				var job models.Job
				if err := json.Unmarshal([]byte(jobData), &job); err != nil {
					continue
				}

				if job.ID == jobID {
					_, err := q.client.LRem(ctx, key, 1, jobData).Result()
					if err == nil {
						q.logger.Debug("job deleted", "job_id", jobID)
						return nil
					}
				}
			}
		} else {
			jobs, err := q.client.ZRange(ctx, key, 0, -1).Result()
			if err != nil {
				continue
			}

			for _, jobData := range jobs {
				var job models.Job
				if err := json.Unmarshal([]byte(jobData), &job); err != nil {
					continue
				}

				if job.ID == jobID {
					_, err := q.client.ZRem(ctx, key, jobData).Result()

					if err == nil {
						q.logger.Debug("job deleted", "job_id", jobID)
						return nil
					}
				}
			}
		}
	}

	return errors.Newf("job %s not found", jobID).WithCode(errors.CodeNotFound)
}

// Extend extends the visibility timeout for a job
func (q *RedisQueue) Extend(ctx context.Context, jobID uuid.UUID,
	duration time.Duration) error {
	key := q.getVisibilityKey(jobID)
	return q.client.Expire(ctx, key, duration).Err()
}

// Size returns the number of jobs in the queue
func (q *RedisQueue) Size(ctx context.Context) (int64, error) {
	var total int64
	priorities := []models.JobPriority{
		models.JobPriorityLow,
		models.JobPriorityNormal,
		models.JobPriorityHigh,
		models.JobPriorityCritical,
	}

	for _, priority := range priorities {
		count, err := q.client.LLen(ctx, q.getQueueKey(priority)).Result()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get queue size").
				WithCode(errors.CodeInternal)
		}

		total += count
	}

	// Add delayed jobs
	delayedCount, err := q.client.ZCard(ctx, q.getDelayedKey()).Result()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get delayed queue size").
			WithCode(errors.CodeInternal)
	}

	total += delayedCount

	return total, nil
}

// Clear removes all jobs from the queue
func (q *RedisQueue) Clear(ctx context.Context) error {
	keys := []string{
		q.getQueueKey(models.JobPriorityLow),
		q.getQueueKey(models.JobPriorityNormal),
		q.getQueueKey(models.JobPriorityHigh),
		q.getQueueKey(models.JobPriorityCritical),
		q.getProcessingKey(),
		q.getDelayedKey(),
	}

	for _, key := range keys {
		if err := q.client.Del(ctx, key).Err(); err != nil {
			return errors.Wrapf(err, "failed to delete key %s", key).WithCode(errors.CodeInternal)
		}
	}

	q.logger.Info("queue cleared")
	return nil
}

// Close closes the queue connection
func (q *RedisQueue) Close() error {
	// Redis client is managed externally
	return nil
}

// Stats returns queue statistics
func (q *RedisQueue) Stats(ctx context.Context) (*QueueStats, error) {
	stats := &QueueStats{
		Name: q.config.Name,
	}

	size, err := q.Size(ctx)
	if err != nil {
		return nil, err
	}

	stats.Size = size
	processingCount, err := q.client.LLen(ctx, q.getProcessingKey()).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get processing count").
			WithCode(errors.CodeInternal)
	}

	stats.Processing = processingCount
	delayedCount, err := q.client.ZCard(ctx, q.getDelayedKey()).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get delayed count").
			WithCode(errors.CodeInternal)
	}

	stats.Delayed = delayedCount
	deadLetterCount, err := q.client.LLen(ctx, q.getDeadLetterKey()).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dead letter count").
			WithCode(errors.CodeInternal)
	}

	stats.DeadLetter = deadLetterCount
	statsKey := fmt.Sprintf("%s:stats", q.keyPrefix)
	statsData, err := q.client.HGetAll(ctx, statsKey).Result()
	if err == nil && len(statsData) > 0 {
		if enqueueRate, ok := statsData["enqueue_rate"]; ok {
			fmt.Sscanf(enqueueRate, "%f", &stats.EnqueueRate)
		}

		if dequeueRate, ok := statsData["dequeue_rate"]; ok {
			fmt.Sscanf(dequeueRate, "%f", &stats.DequeueRate)
		}
	}

	return stats, nil
}

// Helper methods

func (q *RedisQueue) getQueueKey(priority models.JobPriority) string {
	return fmt.Sprintf("%s:%s", q.keyPrefix, GetQueueName(priority))
}

func (q *RedisQueue) getProcessingKey() string {
	return fmt.Sprintf("%s:processing", q.keyPrefix)
}

func (q *RedisQueue) getDelayedKey() string {
	return fmt.Sprintf("%s:delayed", q.keyPrefix)
}

func (q *RedisQueue) getDeadLetterKey() string {
	return fmt.Sprintf("%s:dead_letter", q.keyPrefix)
}

func (q *RedisQueue) getVisibilityKey(jobID uuid.UUID) string {
	return fmt.Sprintf("%s:visibility:%s", q.keyPrefix, jobID)
}

func (q *RedisQueue) processScheduledJobs(ctx context.Context) error {
	now := time.Now().Unix()
	delayedKey := q.getDelayedKey()
	jobs, err := q.client.ZRangeByScore(ctx, delayedKey, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now),
	}).Result()

	if err != nil || len(jobs) == 0 {
		return err
	}

	pipe := q.client.Pipeline()
	for _, jobData := range jobs {
		var job models.Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		queueKey := q.getQueueKey(job.Priority)
		pipe.RPush(ctx, queueKey, jobData)
		pipe.ZRem(ctx, delayedKey, jobData)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) setVisibilityTimeout(ctx context.Context, job *models.Job) error {
	key := q.getVisibilityKey(job.ID)
	return q.client.Set(ctx, key, "1", q.config.VisibilityTimeout).Err()
}

func (q *RedisQueue) clearVisibilityTimeout(ctx context.Context,
	jobID uuid.UUID) {
	key := q.getVisibilityKey(jobID)
	q.client.Del(ctx, key)
}

func (q *RedisQueue) moveToDeadLetter(ctx context.Context, job *models.Job) error {
	job.Status = models.JobStatusDead
	job.UpdatedAt = time.Now()

	data, err := json.Marshal(job)
	if err != nil {
		return errors.Wrap(err, "failed to marshal job").WithCode(errors.CodeSerialization)
	}

	return q.client.RPush(ctx, q.getDeadLetterKey(), data).Err()
}

func (q *RedisQueue) updateEnqueueStats(ctx context.Context) {
	statsKey := fmt.Sprintf("%s:stats", q.keyPrefix)

	q.client.HIncrBy(ctx, statsKey, "total_enqueued", 1)
	q.client.HSet(ctx, statsKey, "last_enqueue_time", time.Now().Unix())
}

func (q *RedisQueue) updateDequeueStats(ctx context.Context) {
	statsKey := fmt.Sprintf("%s:stats", q.keyPrefix)

	q.client.HIncrBy(ctx, statsKey, "total_dequeued", 1)
	q.client.HSet(ctx, statsKey, "last_dequeue_time", time.Now().Unix())
}

// Helper function to create pointer
func ptr[T any](v T) *T {
	return &v
}
