package task

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusAssigned   Status = "assigned"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusRetrying   Status = "retrying"
)

type Task struct {
	ID              string     `json:"id"`
	Type            string     `json:"type"`
	Payload         []byte     `json:"payload"`
	Status          Status     `json:"status"`
	Priority        int        `json:"priority"`
	ComplexityScore int        `json:"complexity_score"`
	Dependencies    []string   `json:"dependencies,omitempty"`
	RetryCount      int        `json:"retry_count"`
	MaxRetries      int        `json:"max_retries"`
	Deadline        *time.Time `json:"deadline,omitempty"`
	NextRetryAt     time.Time  `json:"next_retry_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	WorkerID        string     `json:"worker_id,omitempty"`
}

type Result struct {
	TaskID     string       `json:"task_id"`
	Status     Status       `json:"status"`
	Output     []byte       `json:"output,omitempty"`
	Error      string       `json:"error,omitempty"`
	StartTime  time.Time    `json:"start_time"`
	EndTime    time.Time    `json:"end_time"`
	RetryCount int          `json:"retry_count"`
	WorkerID   string       `json:"worker_id"`
	Metrics    *TaskMetrics `json:"metrics,omitempty"`
}

type TaskMetrics struct {
	ProcessingTime time.Duration `json:"processing_time"`
	QueueWaitTime  time.Duration `json:"queue_wait_time"`
	MemoryUsage    uint64        `json:"memory_usage"`
	CPUTime        float64       `json:"cpu_time"`
}

func NewTask(taskType string, payload []byte) *Task {
	now := time.Now()
	return &Task{
		ID:         uuid.New().String(),
		Type:       taskType,
		Payload:    payload,
		Status:     StatusPending,
		Priority:   1,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (t *Task) WithPriority(priority int) *Task {
	t.Priority = priority
	return t
}

func (t *Task) WithDeadline(deadline time.Time) *Task {
	t.Deadline = &deadline
	return t
}

func (t *Task) WithDependencies(deps ...string) *Task {
	t.Dependencies = deps
	return t
}

func (t *Task) WithMaxRetries(maxRetries int) *Task {
	t.MaxRetries = maxRetries
	return t
}

func (t *Task) IsOverdue() bool {
	if t.Deadline == nil {
		return false
	}
	return time.Now().After(*t.Deadline)
}

func (t *Task) CanRetry() bool {
	return t.RetryCount < t.MaxRetries
}

func (t *Task) ShouldProcess() bool {
	if t.NextRetryAt.IsZero() {
		return true
	}
	return time.Now().After(t.NextRetryAt)
}
