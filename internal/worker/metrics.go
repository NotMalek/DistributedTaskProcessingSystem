package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
)

type MetricsCollector struct {
	workerID string
	redis    *redis.Client
	metrics  *WorkerMetrics
}

type MetricsSnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	QueueLength    int64     `json:"queue_length"`
	TasksProcessed uint64    `json:"tasks_processed"`
	ActiveWorkers  int32     `json:"active_workers"`
	IdleWorkers    int32     `json:"idle_workers"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    uint64    `json:"memory_usage"`
}

func NewMetricsCollector(workerID string, redis *redis.Client, metrics *WorkerMetrics) *MetricsCollector {
	return &MetricsCollector{
		workerID: workerID,
		redis:    redis,
		metrics:  metrics,
	}
}

func (mc *MetricsCollector) Start(ctx context.Context) {
	go mc.collectAndPublish(ctx)
}

func (mc *MetricsCollector) collectAndPublish(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := mc.collectMetrics()
			mc.publishMetrics(ctx, snapshot)
		}
	}
}

func (mc *MetricsCollector) collectMetrics() *MetricsSnapshot {
	return &MetricsSnapshot{
		Timestamp:      time.Now(),
		QueueLength:    atomic.LoadInt64(&mc.metrics.QueueLength),
		TasksProcessed: atomic.LoadUint64(&mc.metrics.TasksProcessed),
		ActiveWorkers:  atomic.LoadInt32(&mc.metrics.ActiveWorkers),
		IdleWorkers:    atomic.LoadInt32(&mc.metrics.IdleWorkers),
		CPUUsage:       mc.collectCPUUsage(),
		MemoryUsage:    mc.collectMemoryUsage(),
	}
}

func (mc *MetricsCollector) publishMetrics(ctx context.Context, snapshot *MetricsSnapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return
	}

	// Publish current metrics
	mc.redis.HSet(ctx, "worker:metrics", mc.workerID, data)

	// Store historical metrics (last 24 hours)
	key := fmt.Sprintf("worker:%s:metrics:history", mc.workerID)
	mc.redis.ZAdd(ctx, key, &redis.Z{
		Score:  float64(snapshot.Timestamp.Unix()),
		Member: data,
	})

	// Cleanup old metrics
	mc.redis.ZRemRangeByScore(ctx, key,
		"0",
		fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()),
	)
}

func (mc *MetricsCollector) collectCPUUsage() float64 {
	// This is a placeholder - in a real implementation,
	// you would use something like github.com/shirou/gopsutil
	// to get actual CPU usage
	return 0.0
}

func (mc *MetricsCollector) collectMemoryUsage() uint64 {
	// This is a placeholder - in a real implementation,
	// you would use something like github.com/shirou/gopsutil
	// to get actual memory usage
	return 0
}
