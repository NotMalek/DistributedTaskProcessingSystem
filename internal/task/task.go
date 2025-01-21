package task

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a task
type Status string

const (
	StatusPending    Status = "pending"
	StatusAssigned   Status = "assigned"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Task represents a unit of work to be processed
type Task struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Payload         []byte    `json:"payload"`
	Status          Status    `json:"status"`
	ComplexityScore int       `json:"complexity_score"`
	Priority        int       `json:"priority"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewTask creates a new task with default values
func NewTask(taskType string, payload []byte) *Task {
	now := time.Now()
	return &Task{
		ID:              uuid.New().String(),
		Type:            taskType,
		Payload:         payload,
		Status:          StatusPending,
		ComplexityScore: 1,
		Priority:        1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Result represents the outcome of processing a task
type Result struct {
	TaskID    string    `json:"task_id"`
	Status    Status    `json:"status"`
	Output    []byte    `json:"output,omitempty"`
	Error     string    `json:"error,omitempty"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
