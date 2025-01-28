package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Server struct {
	redis   *redis.Client
	metrics sync.Map
	logger  *log.Logger
}

type SystemMetrics struct {
	ActiveWorkers  int                   `json:"activeWorkers"`
	TotalTasks     int64                 `json:"totalTasks"`
	ProcessedTasks int64                 `json:"processedTasks"`
	FailedTasks    int64                 `json:"failedTasks"`
	QueueLengths   map[int]int64         `json:"queueLengths"`
	WorkerMetrics  map[string]WorkerInfo `json:"workerMetrics"`
}

type WorkerInfo struct {
	ID             string    `json:"id"`
	LastSeen       time.Time `json:"lastSeen"`
	TasksProcessed uint64    `json:"tasks_processed"`
	ActiveTasks    int       `json:"activeTasks"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    uint64    `json:"memory_usage"`
	Status         string    `json:"status"`
}

func NewServer(redis *redis.Client) *Server {
	return &Server{
		redis:  redis,
		logger: log.New(log.Writer(), "[API Server] ", log.LstdFlags),
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/api/metrics", corsMiddleware(s.handleMetrics))
	mux.Handle("/api/workers", corsMiddleware(s.handleWorkers))

	go s.collectMetrics()

	s.logger.Printf("API server starting on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func corsMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, ok := s.metrics.Load("current")
	if !ok {
		http.Error(w, "No metrics available", http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		s.logger.Printf("Error encoding metrics: %v\n", err)
		http.Error(w, "Error encoding metrics", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleWorkers(w http.ResponseWriter, r *http.Request) {
	metrics, ok := s.metrics.Load("current")
	if !ok {
		http.Error(w, "No metrics available", http.StatusNotFound)
		return
	}

	if sysMetrics, ok := metrics.(*SystemMetrics); ok {
		if err := json.NewEncoder(w).Encode(sysMetrics.WorkerMetrics); err != nil {
			s.logger.Printf("Error encoding worker metrics: %v\n", err)
			http.Error(w, "Error encoding worker metrics", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) collectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := &SystemMetrics{
			QueueLengths:  make(map[int]int64),
			WorkerMetrics: make(map[string]WorkerInfo),
		}

		// Collect queue lengths per priority
		for priority := 1; priority <= 10; priority++ {
			queueKey := fmt.Sprintf("tasks:priority:%d", priority)
			length, err := s.redis.ZCard(context.Background(), queueKey).Result()
			if err == nil {
				metrics.QueueLengths[priority] = length
				metrics.TotalTasks += length
			}
		}

		// Collect processed tasks
		processed, _ := s.redis.HLen(context.Background(), "results").Result()
		metrics.ProcessedTasks = int64(processed)

		// Collect failed tasks
		failed, _ := s.redis.HLen(context.Background(), "failed_tasks").Result()
		metrics.FailedTasks = int64(failed)

		// Collect worker metrics
		workers, _ := s.redis.HGetAll(context.Background(), "workers").Result()
		metrics.ActiveWorkers = len(workers)

		for workerID := range workers {
			tasks, _ := s.redis.HGetAll(context.Background(), fmt.Sprintf("worker:%s:tasks", workerID)).Result()

			workerInfo := WorkerInfo{
				ID:          workerID,
				LastSeen:    time.Now(),
				ActiveTasks: len(tasks),
				Status:      "active",
			}

			// Get task processing count
			processedCount, _ := s.redis.HLen(context.Background(), fmt.Sprintf("worker:%s:results", workerID)).Result()
			workerInfo.TasksProcessed = uint64(processedCount)

			metrics.WorkerMetrics[workerID] = workerInfo
		}

		s.metrics.Store("current", metrics)
		metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
		s.logger.Printf("Current metrics: %s\n", string(metricsJSON))
	}
}
