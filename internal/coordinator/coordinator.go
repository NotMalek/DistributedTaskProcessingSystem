package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/task"
	"github.com/go-redis/redis/v8"
)

type Coordinator struct {
	logger   *log.Logger
	redis    *redis.Client
	workers  sync.Map
	shutdown chan struct{}
}

type Option func(*Coordinator)

func WithLogger(logger *log.Logger) Option {
	return func(c *Coordinator) {
		c.logger = logger
	}
}

func WithRedis(url string) Option {
	return func(c *Coordinator) {
		c.redis = redis.NewClient(&redis.Options{
			Addr: url,
		})
	}
}

func New(opts ...Option) *Coordinator {
	c := &Coordinator{
		shutdown: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Coordinator) cleanup(ctx context.Context) error {
	pipe := c.redis.Pipeline()

	// Clear all priority queues
	for priority := 1; priority <= 10; priority++ {
		pipe.Del(ctx, fmt.Sprintf("tasks:priority:%d", priority))
	}

	// Get all workers to clean their data
	workers, err := c.redis.HGetAll(ctx, "workers").Result()
	if err != nil {
		return fmt.Errorf("failed to get workers: %w", err)
	}

	// Clean up worker data
	for workerID := range workers {
		pipe.Del(ctx, fmt.Sprintf("worker:%s:tasks", workerID))
		pipe.Del(ctx, fmt.Sprintf("worker:%s:results", workerID))
		pipe.Del(ctx, fmt.Sprintf("worker:%s:processing", workerID))
	}

	// Clean up global keys
	pipe.Del(ctx, "workers")
	pipe.Del(ctx, "results")
	pipe.Del(ctx, "failed_tasks")

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute cleanup: %w", err)
	}

	c.logger.Printf("System state cleaned up successfully")
	return nil
}

func (c *Coordinator) Start(ctx context.Context) error {
	// Clean up any existing state
	if err := c.cleanup(ctx); err != nil {
		c.logger.Printf("Warning: Failed to cleanup system state: %v", err)
	}

	go c.distributeWork(ctx)
	go c.collectResults(ctx)
	go c.monitorWorkers(ctx)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.shutdown:
		return nil
	}
}

func (c *Coordinator) distributeWork(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get active workers
			var availableWorkers []string
			c.workers.Range(func(key, value interface{}) bool {
				workerID := key.(string)
				availableWorkers = append(availableWorkers, workerID)
				return true
			})

			if len(availableWorkers) == 0 {
				continue
			}

			// Try getting tasks from highest to lowest priority
			for priority := 10; priority > 0; priority-- {
				queueKey := fmt.Sprintf("tasks:priority:%d", priority)

				// Try to get up to 5 tasks at once
				result, err := c.redis.ZRange(ctx, queueKey, 0, 4).Result()
				if err != nil || len(result) == 0 {
					continue
				}

				c.logger.Printf("Found %d tasks in priority %d queue", len(result), priority)

				// Process each task
				for _, taskStr := range result {
					var currentTask task.Task
					if err := json.Unmarshal([]byte(taskStr), &currentTask); err != nil {
						c.logger.Printf("Error unmarshaling task: %v", err)
						continue
					}

					// Pick a worker (round-robin)
					workerID := availableWorkers[0]
					availableWorkers = append(availableWorkers[1:], availableWorkers[0])

					c.logger.Printf("Assigning task %s to worker %s", currentTask.ID, workerID)

					// Assign task to worker
					err = c.redis.HSet(ctx,
						fmt.Sprintf("worker:%s:tasks", workerID),
						currentTask.ID,
						taskStr,
					).Err()

					if err != nil {
						c.logger.Printf("Failed to assign task to worker: %v", err)
						continue
					}

					// Remove task from priority queue
					c.redis.ZRem(ctx, queueKey, taskStr)
				}
			}
		}
	}
}

func (c *Coordinator) collectResults(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.workers.Range(func(key, value interface{}) bool {
				workerID := key.(string)
				results, err := c.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:results", workerID)).Result()
				if err != nil {
					return true
				}

				for taskID, resultStr := range results {
					c.redis.HSet(ctx, "results", taskID, resultStr)
					c.redis.HDel(ctx, fmt.Sprintf("worker:%s:results", workerID), taskID)
				}

				return true
			})
		}
	}
}

func (c *Coordinator) monitorWorkers(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			workers, err := c.redis.HGetAll(ctx, "workers").Result()
			if err != nil {
				continue
			}

			now := time.Now().Unix()
			for workerID, lastSeenStr := range workers {
				lastSeen, err := strconv.ParseInt(lastSeenStr, 10, 64)
				if err != nil {
					continue
				}

				if now-lastSeen <= 30 {
					c.workers.Store(workerID, time.Unix(lastSeen, 0))
				} else {
					c.workers.Delete(workerID)
					c.redis.HDel(ctx, "workers", workerID)

					tasks, _ := c.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:tasks", workerID)).Result()
					for _, taskStr := range tasks {
						c.redis.RPush(ctx, "tasks", taskStr)
					}

					c.redis.Del(ctx, fmt.Sprintf("worker:%s:tasks", workerID))
					c.redis.Del(ctx, fmt.Sprintf("worker:%s:results", workerID))
				}
			}
		}
	}
}

func (c *Coordinator) RegisterWorker(id string) {
	now := time.Now()
	c.redis.HSet(context.Background(), "workers", id, now.Unix())
	c.workers.Store(id, now)
}

func (c *Coordinator) UpdateWorkerHeartbeat(id string) {
	now := time.Now()
	c.redis.HSet(context.Background(), "workers", id, now.Unix())
	c.workers.Store(id, now)
}
