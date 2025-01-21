package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		tasks:    make(chan *task.Task, 100),
		results:  make(chan *task.Result, 100),
		metrics:  &WorkerMetrics{},
		shutdown: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *Worker) Start(ctx context.Context) error {
	err := w.register(ctx)
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	w.wg.Add(w.poolSize)
	for i := 0; i < w.poolSize; i++ {
		go w.processTask(ctx)
	}

	go w.sendHeartbeat(ctx)
	go w.checkForWork(ctx)
	go w.submitResults(ctx)

	select {
	case <-ctx.Done():
		close(w.shutdown)
		w.wg.Wait()
		return ctx.Err()
	case <-w.shutdown:
		w.wg.Wait()
		return nil
	}
}

func (w *Worker) register(ctx context.Context) error {
	err := w.redis.HSet(ctx, "workers", w.id, time.Now().Unix()).Err()
	if err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	err = w.redis.Del(ctx,
		fmt.Sprintf("worker:%s:tasks", w.id),
		fmt.Sprintf("worker:%s:results", w.id)).Err()
	if err != nil {
		return fmt.Errorf("failed to initialize worker keys: %w", err)
	}

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
			w.redis.HSet(ctx, "workers", w.id, time.Now().Unix())
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
				continue
			}

			atomic.StoreInt64(&w.metrics.QueueLength, int64(len(tasks)))

			for taskID, taskStr := range tasks {
				var t task.Task
				if err := json.Unmarshal([]byte(taskStr), &t); err != nil {
					continue
				}

				select {
				case w.tasks <- &t:
					w.redis.HDel(ctx, fmt.Sprintf("worker:%s:tasks", w.id), taskID)
				default:
				}
			}
		}
	}
}

func (w *Worker) processTask(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.shutdown:
			return
		case t := <-w.tasks:
			atomic.AddInt32(&w.metrics.IdleWorkers, -1)

			result := &task.Result{
				TaskID:    t.ID,
				StartTime: time.Now(),
				WorkerID:  w.id,
			}

			// Simulate work
			time.Sleep(time.Duration(t.ComplexityScore) * time.Second)

			result.EndTime = time.Now()
			result.Status = task.StatusCompleted

			atomic.AddUint64(&w.metrics.TasksProcessed, 1)
			atomic.AddInt32(&w.metrics.IdleWorkers, 1)

			select {
			case w.results <- result:
			default:
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
			resultBytes, err := json.Marshal(result)
			if err != nil {
				continue
			}

			err = w.redis.HSet(ctx,
				fmt.Sprintf("worker:%s:results", w.id),
				result.TaskID,
				resultBytes,
			).Err()

			if err != nil {
				continue
			}
		}
	}
}

func (w *Worker) GetMetrics() *WorkerMetrics {
	return w.metrics
}
