package worker

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

type WorkStealer struct {
	workerID string
	redis    *redis.Client
	metrics  *WorkerMetrics
}

func NewWorkStealer(workerID string, redis *redis.Client, metrics *WorkerMetrics) *WorkStealer {
	return &WorkStealer{
		workerID: workerID,
		redis:    redis,
		metrics:  metrics,
	}
}

func (ws *WorkStealer) Start(ctx context.Context) {
	go ws.monitorAndSteal(ctx)
}

func (ws *WorkStealer) monitorAndSteal(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if atomic.LoadInt32(&ws.metrics.IdleWorkers) > 0 {
				ws.attemptSteal(ctx)
			}
		}
	}
}

func (ws *WorkStealer) attemptSteal(ctx context.Context) error {
	// Get all workers
	workers, err := ws.redis.HGetAll(ctx, "workers").Result()
	if err != nil {
		return err
	}

	// Find busy workers
	for workerID, _ := range workers {
		if workerID == ws.workerID {
			continue
		}

		// Check their queue length
		queueKey := fmt.Sprintf("worker:%s:tasks", workerID)
		queueLen, err := ws.redis.HLen(ctx, queueKey).Result()
		if err != nil {
			continue
		}

		// If they have more than 2 tasks, try to steal
		if queueLen > 2 {
			ws.stealTasks(ctx, workerID, queueKey)
		}
	}

	return nil
}

func (ws *WorkStealer) stealTasks(ctx context.Context, targetWorker, queueKey string) {
	// Get tasks from target worker
	tasks, err := ws.redis.HGetAll(ctx, queueKey).Result()
	if err != nil {
		return
	}

	// Try to steal half of their tasks
	stealCount := len(tasks) / 2
	stolen := 0

	for taskID, taskData := range tasks {
		if stolen >= stealCount {
			break
		}

		// Try to move task to our queue
		err = ws.redis.HSetNX(ctx,
			fmt.Sprintf("worker:%s:tasks", ws.workerID),
			taskID,
			taskData,
		).Err()

		if err == nil {
			// If successful, remove from original worker
			ws.redis.HDel(ctx, queueKey, taskID)
			stolen++
		}
	}
}
