package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	TasksProcessed uint64    `json:"tasksProcessed"`
	ActiveTasks    int       `json:"activeTasks"`
	Status         string    `json:"status"`
}

func NewServer(redis *redis.Client) *Server {
	return &Server{
		redis:  redis,
		logger: log.New(log.Writer(), "[API Server] ", log.LstdFlags),
	}
}

func (s *Server) handleDebug(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	debug := make(map[string]interface{})

	// Get all queue contents
	for priority := 1; priority <= 10; priority++ {
		queueKey := fmt.Sprintf("tasks:priority:%d", priority)
		tasks, err := s.redis.ZRange(ctx, queueKey, 0, -1).Result()
		if err == nil {
			debug[fmt.Sprintf("queue_%d", priority)] = tasks
		}
	}

	// Get all worker states
	workers, _ := s.redis.HGetAll(ctx, "workers").Result()
	workerStates := make(map[string]interface{})

	for workerID := range workers {
		state := make(map[string]interface{})

		// Get assigned tasks
		tasks, _ := s.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:tasks", workerID)).Result()
		state["assigned_tasks"] = tasks

		// Get processing tasks
		processing, _ := s.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:processing", workerID)).Result()
		state["processing_tasks"] = processing

		// Get completed tasks
		completed, _ := s.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:results", workerID)).Result()
		state["completed_tasks"] = completed

		workerStates[workerID] = state
	}
	debug["workers"] = workerStates

	// Get all results
	results, _ := s.redis.HGetAll(ctx, "results").Result()
	debug["results"] = results

	// Get all failed tasks
	failed, _ := s.redis.HGetAll(ctx, "failed_tasks").Result()
	debug["failed_tasks"] = failed

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debug)
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/api/metrics", corsMiddleware(s.handleMetrics))
	mux.Handle("/api/workers", corsMiddleware(s.handleWorkers))
	mux.Handle("/api/debug", corsMiddleware(s.handleDebug)) // Add debug endpoint

	go s.collectMetrics()

	s.logger.Printf("API server starting on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func corsMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
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

	json.NewEncoder(w).Encode(metrics)
}

func (s *Server) handleWorkers(w http.ResponseWriter, r *http.Request) {
	metrics, ok := s.metrics.Load("current")
	if !ok {
		http.Error(w, "No metrics available", http.StatusNotFound)
		return
	}

	if sysMetrics, ok := metrics.(*SystemMetrics); ok {
		json.NewEncoder(w).Encode(sysMetrics.WorkerMetrics)
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

		// Collect total tasks in priority queues
		total := int64(0)
		for priority := 1; priority <= 10; priority++ {
			queueKey := fmt.Sprintf("tasks:priority:%d", priority)
			length, err := s.redis.ZCard(context.Background(), queueKey).Result()
			if err == nil {
				metrics.QueueLengths[priority] = length
				total += length
			}
		}
		metrics.TotalTasks = total

		// Collect all results (completed tasks)
		processed, _ := s.redis.HLen(context.Background(), "results").Result()
		metrics.ProcessedTasks = int64(processed)

		// Collect failed tasks
		failed, _ := s.redis.HLen(context.Background(), "failed_tasks").Result()
		metrics.FailedTasks = int64(failed)

		// Collect active workers and their tasks
		workers, _ := s.redis.HGetAll(context.Background(), "workers").Result()
		metrics.ActiveWorkers = len(workers)

		for workerID, lastSeenStr := range workers {
			lastSeen, _ := strconv.ParseInt(lastSeenStr, 10, 64)

			// Get tasks currently assigned to this worker
			assignedTasks, _ := s.redis.HGetAll(context.Background(),
				fmt.Sprintf("worker:%s:tasks", workerID)).Result()

			// Get tasks being processed by this worker
			processingTasks, _ := s.redis.HGetAll(context.Background(),
				fmt.Sprintf("worker:%s:processing", workerID)).Result()

			// Get completed tasks by this worker
			completedTasks, _ := s.redis.HGetAll(context.Background(),
				fmt.Sprintf("worker:%s:results", workerID)).Result()

			workerInfo := WorkerInfo{
				ID:             workerID,
				LastSeen:       time.Unix(lastSeen, 0),
				TasksProcessed: uint64(len(completedTasks)),
				ActiveTasks:    len(assignedTasks) + len(processingTasks),
				Status:         "active",
			}

			// Check if worker is actually active (last seen within 30 seconds)
			if time.Since(workerInfo.LastSeen) > 30*time.Second {
				workerInfo.Status = "inactive"
			}

			metrics.WorkerMetrics[workerID] = workerInfo
		}

		s.metrics.Store("current", metrics)

		// Log current state
		s.logger.Printf("Current State - Active Workers: %d, Total Tasks: %d, Processed: %d, Failed: %d",
			metrics.ActiveWorkers,
			metrics.TotalTasks,
			metrics.ProcessedTasks,
			metrics.FailedTasks)

		// Log queue lengths
		for priority, length := range metrics.QueueLengths {
			if length > 0 {
				s.logger.Printf("Priority %d queue length: %d", priority, length)
			}
		}
	}
}
