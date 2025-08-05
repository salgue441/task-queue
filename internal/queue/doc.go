// Package queue provides the queue abstraction and implementations for the
// task queue system. It supports multiple backend implementations (Redis,
// RabbitMQ) with a unified interface for job enqueueing, dequeuing, and
// management.
//
// The package handles job priorities, delayed scheduling, visibility timeouts,
// and dead letter queue management. It ensures at-least-once delivery semantics
// with support for job acknowledgment and redelivery.
//
// Basic usage:
//
//	q, err := queue.NewRedisQueue(redisClient, queue.Config{
//	    Name: "default",
//	    MaxSize: 10000,
//	})
//
//	// Enqueue a job
//	err = q.Enqueue(ctx, job)
//
//	// Dequeue a job for processing
//	job, err := q.Dequeue(ctx)
//
//	// Acknowledge successful processing
//	err = q.Ack(ctx, job.ID)
package queue
