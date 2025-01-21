package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type Dashboard struct {
	redis     *redis.Client
	templates *template.Template
	metrics   sync.Map
}

type SystemMetrics struct {
	ActiveWorkers  int                      `json:"active_workers"`
	TotalTasks     int64                    `json:"total_tasks"`
	ProcessedTasks int64                    `json:"processed_tasks"`
	FailedTasks    int64                    `json:"failed_tasks"`
	QueueLengths   map[int]int64            `json:"queue_lengths"`
	WorkerMetrics  map[string]*WorkerStatus `json:"worker_metrics"`
}

type WorkerStatus struct {
	ID             string    `json:"id"`
	LastSeen       time.Time `json:"last_seen"`
	TasksProcessed uint64    `json:"tasks_processed"`
	ActiveTasks    int       `json:"active_tasks"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryUsage    uint64    `json:"memory_usage"`
	Status         string    `json:"status"`
}

func NewDashboard(redis *redis.Client) *Dashboard {
	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	return &Dashboard{
		redis:     redis,
		templates: tmpl,
	}
}

func (d *Dashboard) Start() {
	http.HandleFunc("/", d.handleDashboard)
	http.HandleFunc("/api/metrics", d.handleMetrics)
	http.HandleFunc("/api/workers", d.handleWorkers)

	go d.collectMetrics()

	http.ListenAndServe(":8080", nil)
}

func (d *Dashboard) collectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := &SystemMetrics{
			QueueLengths:  make(map[int]int64),
			WorkerMetrics: make(map[string]*WorkerStatus),
		}

		// Collect queue lengths per priority
		for priority := 1; priority <= 10; priority++ {
			queueKey := fmt.Sprintf("tasks:priority:%d", priority)
			length, err := d.redis.ZCard(context.Background(), queueKey).Result()
			if err == nil {
				metrics.QueueLengths[priority] = length
				metrics.TotalTasks += length
			}
		}

		// Collect worker metrics
		workers, _ := d.redis.HGetAll(context.Background(), "worker:metrics").Result()
		for workerID, data := range workers {
			var status WorkerStatus
			if err := json.Unmarshal([]byte(data), &status); err == nil {
				metrics.WorkerMetrics[workerID] = &status
				metrics.ActiveWorkers++
			}
		}

		// Store the latest metrics
		d.metrics.Store("current", metrics)
	}
}

func (d *Dashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	metrics, _ := d.metrics.Load("current")
	d.templates.ExecuteTemplate(w, "dashboard.html", metrics)
}

func (d *Dashboard) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, _ := d.metrics.Load("current")
	json.NewEncoder(w).Encode(metrics)
}

func (d *Dashboard) handleWorkers(w http.ResponseWriter, r *http.Request) {
	metrics, _ := d.metrics.Load("current")
	if sysMetrics, ok := metrics.(*SystemMetrics); ok {
		json.NewEncoder(w).Encode(sysMetrics.WorkerMetrics)
	}
}
