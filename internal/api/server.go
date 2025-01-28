package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/task"
	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/worker"
	"github.com/go-redis/redis/v8"
)

type Server struct {
	redis   *redis.Client
	metrics sync.Map
	workers sync.Map // Track active worker instances
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

// Request structures
type StartWorkerRequest struct {
	PoolSize    int  `json:"poolSize"`
	EnableSteal bool `json:"enableSteal"`
	MinWorkers  int  `json:"minWorkers"`
	MaxWorkers  int  `json:"maxWorkers"`
}

type SubmitTaskRequest struct {
	Priority int    `json:"priority"`
	Deadline string `json:"deadline,omitempty"`
	Retries  int    `json:"retries"`
	TaskType string `json:"taskType"`
	Payload  string `json:"payload"`
}

func NewServer(redis *redis.Client) *Server {
	return &Server{
		redis:  redis,
		logger: log.New(os.Stdout, "[API Server] ", log.LstdFlags),
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// System endpoints
	mux.Handle("/api/metrics", corsMiddleware(s.handleMetrics))
	mux.Handle("/api/debug", corsMiddleware(s.handleDebug))
	mux.Handle("/api/system/reset", corsMiddleware(s.handleReset))

	// Worker endpoints
	mux.Handle("/api/workers", corsMiddleware(s.handleWorkers))
	mux.Handle("/api/workers/start", corsMiddleware(s.handleStartWorker))
	mux.Handle("/api/workers/stop", corsMiddleware(s.handleStopWorker))

	// Task endpoints
	mux.Handle("/api/tasks/submit", corsMiddleware(s.handleSubmitTask))
	mux.Handle("/api/tasks/status", corsMiddleware(s.handleTaskStatus))

	go s.collectMetrics()

	s.logger.Printf("API server starting on %s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func corsMiddleware(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	})
}

func (s *Server) handleStartWorker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Create and start new worker
	newWorker := worker.NewWorker(
		worker.WithLogger(log.New(os.Stdout, "[Worker] ", log.LstdFlags)),
		worker.WithRedis(s.redis.Options().Addr),
		worker.WithPoolSize(req.PoolSize),
	)

	go func() {
		if err := newWorker.Start(context.Background()); err != nil {
			s.logger.Printf("Worker failed: %v", err)
		}
	}()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "Worker started",
		"config": req,
	})
}

func (s *Server) handleStopWorker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workerID := r.URL.Query().Get("id")
	if workerID == "" {
		http.Error(w, "Worker ID required", http.StatusBadRequest)
		return
	}

	// Remove worker from Redis
	s.redis.HDel(context.Background(), "workers", workerID)

	// Clean up worker data
	s.redis.Del(context.Background(),
		fmt.Sprintf("worker:%s:tasks", workerID),
		fmt.Sprintf("worker:%s:results", workerID),
		fmt.Sprintf("worker:%s:processing", workerID),
	)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Worker stopped",
		"id":     workerID,
	})
}

func (s *Server) handleSubmitTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Create new task
	newTask := task.NewTask(req.TaskType, []byte(req.Payload))
	newTask.Priority = req.Priority
	newTask.MaxRetries = req.Retries

	if req.Deadline != "" {
		deadline, err := time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			http.Error(w, "Invalid deadline format", http.StatusBadRequest)
			return
		}
		newTask.Deadline = &deadline
	}

	// Queue the task
	taskBytes, _ := json.Marshal(newTask)
	queueKey := fmt.Sprintf("tasks:priority:%d", newTask.Priority)
	err := s.redis.ZAdd(context.Background(), queueKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: taskBytes,
	}).Err()

	if err != nil {
		http.Error(w, "Failed to queue task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"taskId": newTask.ID,
		"status": "queued",
	})
}

func (s *Server) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Check results
	result, err := s.redis.HGet(context.Background(), "results", taskID).Result()
	if err == nil {
		var taskResult task.Result
		json.Unmarshal([]byte(result), &taskResult)
		json.NewEncoder(w).Encode(taskResult)
		return
	}

	// Check failed tasks
	failed, err := s.redis.HGet(context.Background(), "failed_tasks", taskID).Result()
	if err == nil {
		var taskResult task.Result
		json.Unmarshal([]byte(failed), &taskResult)
		json.NewEncoder(w).Encode(taskResult)
		return
	}

	http.Error(w, "Task not found", http.StatusNotFound)
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	pipe := s.redis.Pipeline()

	// Clear all task queues
	for priority := 1; priority <= 10; priority++ {
		pipe.Del(ctx, fmt.Sprintf("tasks:priority:%d", priority))
	}

	// Clear worker data
	workers, _ := s.redis.HGetAll(ctx, "workers").Result()
	for workerID := range workers {
		pipe.Del(ctx, fmt.Sprintf("worker:%s:tasks", workerID))
		pipe.Del(ctx, fmt.Sprintf("worker:%s:results", workerID))
		pipe.Del(ctx, fmt.Sprintf("worker:%s:processing", workerID))
	}

	// Clear global keys
	pipe.Del(ctx, "workers")
	pipe.Del(ctx, "results")
	pipe.Del(ctx, "failed_tasks")

	_, err := pipe.Exec(ctx)
	if err != nil {
		http.Error(w, "Failed to reset system", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "System reset successful",
	})
}

// Your existing handlers and metrics collection
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

// Your existing metrics collection method
func (s *Server) collectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := &SystemMetrics{
			QueueLengths:  make(map[int]int64),
			WorkerMetrics: make(map[string]WorkerInfo),
		}

		// Collection logic from your existing code
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

		processed, _ := s.redis.HLen(context.Background(), "results").Result()
		metrics.ProcessedTasks = int64(processed)

		failed, _ := s.redis.HLen(context.Background(), "failed_tasks").Result()
		metrics.FailedTasks = int64(failed)

		workers, _ := s.redis.HGetAll(context.Background(), "workers").Result()
		metrics.ActiveWorkers = len(workers)

		for workerID, lastSeenStr := range workers {
			lastSeen, _ := strconv.ParseInt(lastSeenStr, 10, 64)
			assignedTasks, _ := s.redis.HGetAll(context.Background(),
				fmt.Sprintf("worker:%s:tasks", workerID)).Result()
			processingTasks, _ := s.redis.HGetAll(context.Background(),
				fmt.Sprintf("worker:%s:processing", workerID)).Result()
			completedTasks, _ := s.redis.HGetAll(context.Background(),
				fmt.Sprintf("worker:%s:results", workerID)).Result()

			workerInfo := WorkerInfo{
				ID:             workerID,
				LastSeen:       time.Unix(lastSeen, 0),
				TasksProcessed: uint64(len(completedTasks)),
				ActiveTasks:    len(assignedTasks) + len(processingTasks),
				Status:         "active",
			}

			if time.Since(workerInfo.LastSeen) > 30*time.Second {
				workerInfo.Status = "inactive"
			}

			metrics.WorkerMetrics[workerID] = workerInfo
		}

		s.metrics.Store("current", metrics)

		// Logging current state
		s.logger.Printf("Current State - Active Workers: %d, Total Tasks: %d, Processed: %d, Failed: %d",
			metrics.ActiveWorkers,
			metrics.TotalTasks,
			metrics.ProcessedTasks,
			metrics.FailedTasks)

		for priority, length := range metrics.QueueLengths {
			if length > 0 {
				s.logger.Printf("Priority %d queue length: %d", priority, length)
			}
		}
	}
}

func (s *Server) handleDebug(w http.ResponseWriter, r *http.Request) {
	// Your existing debug handler code
	ctx := context.Background()
	debug := make(map[string]interface{})

	for priority := 1; priority <= 10; priority++ {
		queueKey := fmt.Sprintf("tasks:priority:%d", priority)
		tasks, err := s.redis.ZRange(ctx, queueKey, 0, -1).Result()
		if err == nil {
			debug[fmt.Sprintf("queue_%d", priority)] = tasks
		}
	}

	workers, _ := s.redis.HGetAll(ctx, "workers").Result()
	workerStates := make(map[string]interface{})

	for workerID := range workers {
		state := make(map[string]interface{})
		tasks, _ := s.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:tasks", workerID)).Result()
		state["assigned_tasks"] = tasks
		processing, _ := s.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:processing", workerID)).Result()
		state["processing_tasks"] = processing
		completed, _ := s.redis.HGetAll(ctx, fmt.Sprintf("worker:%s:results", workerID)).Result()
		state["completed_tasks"] = completed
		workerStates[workerID] = state
	}
	debug["workers"] = workerStates

	results, _ := s.redis.HGetAll(ctx, "results").Result()
	debug["results"] = results

	failed, _ := s.redis.HGetAll(ctx, "failed_tasks").Result()
	debug["failed_tasks"] = failed

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(debug)
}
