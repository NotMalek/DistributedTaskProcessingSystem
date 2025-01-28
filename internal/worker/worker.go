package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/task"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type WorkerMetrics struct {
	TasksProcessed uint64
	QueueLength    int64
	ActiveWorkers  int32
	IdleWorkers    int32
	CPUUsage       float64
	MemoryUsage    uint64
}

type Worker struct {
	id       string
	logger   *log.Logger
	redis    *redis.Client
	poolSize int
	tasks    chan *task.Task
	results  chan *task.Result
	metrics  *WorkerMetrics
	wg       sync.WaitGroup
	shutdown chan struct{}
}

type Option func(*Worker)

func WithLogger(logger *log.Logger) Option {
	return func(w *Worker) {
		w.logger = logger
	}
}

func WithRedis(url string) Option {
	return func(w *Worker) {
		w.redis = redis.NewClient(&redis.Options{
			Addr: url,
		})
	}
}

func WithPoolSize(size int) Option {
	return func(w *Worker) {
		w.poolSize = size
	}
}

func NewWorker(opts ...Option) *Worker {
	w := &Worker{
		id:       uuid.New().String(),
		poolSize: 1,
		tasks:    make(chan *task.Task, 1000),
		results:  make(chan *task.Result, 1000),
		metrics: &WorkerMetrics{
			IdleWorkers: 1,
		},
		shutdown: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(w)
	}

	if w.logger == nil {
		w.logger = log.New(os.Stdout, fmt.Sprintf("[Worker %s] ", w.id), log.LstdFlags)
	}

	return w
}

func (w *Worker) Start(ctx context.Context) error {
	w.logger.Printf("Starting worker with pool size %d", w.poolSize)

	err := w.register(ctx)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	atomic.StoreInt32(&w.metrics.IdleWorkers, int32(w.poolSize))
	atomic.StoreInt32(&w.metrics.ActiveWorkers, int32(w.poolSize))

	for i := 0; i < w.poolSize; i++ {
		w.wg.Add(1)
		go w.processTask(ctx)
	}

	go w.sendHeartbeat(ctx)
	go w.checkForWork(ctx)
	go w.submitResults(ctx)

	<-ctx.Done()
	w.logger.Printf("Context cancelled, initiating shutdown")
	close(w.shutdown)
	w.wg.Wait()
	return ctx.Err()
}

func (w *Worker) register(ctx context.Context) error {
	pipe := w.redis.Pipeline()

	// Register worker
	pipe.HSet(ctx, "workers", w.id, time.Now().Unix())

	// Clean up any previous state
	pipe.Del(ctx, fmt.Sprintf("worker:%s:tasks", w.id))
	pipe.Del(ctx, fmt.Sprintf("worker:%s:results", w.id))
	pipe.Del(ctx, fmt.Sprintf("worker:%s:processing", w.id))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	w.logger.Printf("Worker registered successfully")
	return nil
}

func (w *Worker) sendHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case <-ticker.C:
			err := w.redis.HSet(ctx, "workers", w.id, time.Now().Unix()).Err()
			if err != nil {
				w.logger.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}

func (w *Worker) checkForWork(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case <-ticker.C:
			tasks, err := w.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:tasks", w.id)).Result()
			if err != nil {
				w.logger.Printf("Failed to fetch tasks: %v", err)
				continue
			}

			if len(tasks) > 0 {
				w.logger.Printf("Found %d tasks to process", len(tasks))
			}

			atomic.StoreInt64(&w.metrics.QueueLength, int64(len(tasks)))

			for taskID, taskStr := range tasks {
				var t task.Task
				if err := json.Unmarshal([]byte(taskStr), &t); err != nil {
					w.logger.Printf("Failed to unmarshal task %s: %v", taskID, err)
					// Move to failed tasks
					w.redis.HSet(ctx, "failed_tasks", taskID, taskStr)
					w.redis.HDel(ctx, fmt.Sprintf("worker:%s:tasks", w.id), taskID)
					continue
				}

				// Try to send task for processing
				select {
				case w.tasks <- &t:
					w.logger.Printf("Task %s queued for processing", t.ID)
					w.redis.HDel(ctx, fmt.Sprintf("worker:%s:tasks", w.id), taskID)
				case <-time.After(100 * time.Millisecond):
					w.logger.Printf("Failed to queue task %s - processing channel full", t.ID)
				}
			}
		}
	}
}

func (w *Worker) processTask(ctx context.Context) {
	defer w.wg.Done()
	defer atomic.AddInt32(&w.metrics.ActiveWorkers, -1)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case t := <-w.tasks:
			if t == nil {
				continue
			}

			atomic.AddInt32(&w.metrics.IdleWorkers, -1)
			w.logger.Printf("Processing task %s", t.ID)

			result := &task.Result{
				TaskID:    t.ID,
				StartTime: time.Now(),
				WorkerID:  w.id,
				Status:    task.StatusProcessing,
			}

			// Mark task as processing
			t.Status = task.StatusProcessing
			taskBytes, _ := json.Marshal(t)
			w.redis.HSet(ctx, fmt.Sprintf("worker:%s:processing", w.id), t.ID, taskBytes)

			// Simulate work
			time.Sleep(time.Duration(t.ComplexityScore) * time.Second)

			result.EndTime = time.Now()
			result.Status = task.StatusCompleted

			atomic.AddUint64(&w.metrics.TasksProcessed, 1)
			atomic.AddInt32(&w.metrics.IdleWorkers, 1)

			// Remove from processing set
			w.redis.HDel(ctx, fmt.Sprintf("worker:%s:processing", w.id), t.ID)

			// Queue the result
			select {
			case w.results <- result:
				w.logger.Printf("Task %s completed and result queued", t.ID)
			case <-time.After(100 * time.Millisecond):
				w.logger.Printf("Failed to queue result for task %s", t.ID)
			}
		}
	}
}

func (w *Worker) submitResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case result := <-w.results:
			if result == nil {
				continue
			}

			w.logger.Printf("Submitting result for task %s", result.TaskID)
			resultBytes, err := json.Marshal(result)
			if err != nil {
				w.logger.Printf("Failed to marshal result for task %s: %v", result.TaskID, err)
				continue
			}

			err = w.redis.HSet(ctx,
				fmt.Sprintf("worker:%s:results", w.id),
				result.TaskID,
				resultBytes,
			).Err()

			if err != nil {
				w.logger.Printf("Failed to store result for task %s: %v", result.TaskID, err)
				continue
			}

			w.logger.Printf("Successfully submitted result for task %s", result.TaskID)
		}
	}
}

func (w *Worker) GetMetrics() *WorkerMetrics {
	return w.metrics
}
