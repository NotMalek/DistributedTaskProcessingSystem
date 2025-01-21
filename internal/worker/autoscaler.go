package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type AutoScaler struct {
	minWorkers  int32
	maxWorkers  int32
	workerPool  chan struct{}
	metrics     *WorkerMetrics
	mu          sync.Mutex
	lastScaling time.Time
}

func NewAutoScaler(minWorkers, maxWorkers int32, metrics *WorkerMetrics) *AutoScaler {
	return &AutoScaler{
		minWorkers:  minWorkers,
		maxWorkers:  maxWorkers,
		workerPool:  make(chan struct{}, maxWorkers),
		metrics:     metrics,
		lastScaling: time.Now(),
	}
}

func (as *AutoScaler) Start(ctx context.Context, worker *Worker) {
	// Initialize with minimum workers
	for i := int32(0); i < as.minWorkers; i++ {
		as.workerPool <- struct{}{}
		atomic.AddInt32(&as.metrics.ActiveWorkers, 1)
	}

	// Start auto-scaling monitor
	go as.monitor(ctx, worker)
}

func (as *AutoScaler) monitor(ctx context.Context, worker *Worker) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			as.adjust(worker)
		}
	}
}

func (as *AutoScaler) adjust(worker *Worker) {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Don't scale more often than every 30 seconds
	if time.Since(as.lastScaling) < 30*time.Second {
		return
	}

	queueLength := atomic.LoadInt64(&as.metrics.QueueLength)
	activeWorkers := atomic.LoadInt32(&as.metrics.ActiveWorkers)
	idleWorkers := atomic.LoadInt32(&as.metrics.IdleWorkers)

	// Scale up if queue is growing and we have capacity
	if queueLength > int64(activeWorkers*2) && activeWorkers < as.maxWorkers {
		toAdd := minInt32(as.maxWorkers-activeWorkers, 2)
		for i := int32(0); i < toAdd; i++ {
			select {
			case as.workerPool <- struct{}{}:
				atomic.AddInt32(&as.metrics.ActiveWorkers, 1)
				go worker.processTask(context.Background())
			default:
				return
			}
		}
		as.lastScaling = time.Now()
		return
	}

	// Scale down if too many idle workers
	if idleWorkers > as.minWorkers/2 && activeWorkers > as.minWorkers {
		select {
		case <-as.workerPool:
			atomic.AddInt32(&as.metrics.ActiveWorkers, -1)
			as.lastScaling = time.Now()
		default:
		}
	}
}

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
