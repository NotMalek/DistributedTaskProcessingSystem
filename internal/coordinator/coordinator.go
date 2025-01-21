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
	logger *log.Logger
	redis  *redis.Client

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

func (c *Coordinator) Start(ctx context.Context) error {
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

func (c *Coordinator) SubmitTask(task *task.Task) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = c.redis.RPush(context.Background(), "tasks", taskBytes).Err()
	if err != nil {
		return fmt.Errorf("failed to queue task: %w", err)
	}

	return nil
}

func (c *Coordinator) distributeWork(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := c.redis.LPop(ctx, "tasks").Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				continue
			}

			var task task.Task
			if err := json.Unmarshal([]byte(result), &task); err != nil {
				continue
			}

			var workerID string
			c.workers.Range(func(key, value interface{}) bool {
				workerID = key.(string)
				return false
			})

			if workerID == "" {
				c.redis.RPush(ctx, "tasks", result)
				continue
			}

			err = c.redis.HSet(ctx, fmt.Sprintf("worker:%s:tasks", workerID), task.ID, result).Err()
			if err != nil {
				c.redis.RPush(ctx, "tasks", result)
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
