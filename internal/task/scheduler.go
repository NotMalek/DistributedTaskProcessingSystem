package task

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Scheduler struct {
	redis *redis.Client
}

type ScheduleOptions struct {
	Priority     int        `json:"priority"`     // 1-10, higher is more important
	Deadline     *time.Time `json:"deadline"`     // Optional deadline
	MaxRetries   int        `json:"max_retries"`  // Maximum retry attempts
	Dependencies []string   `json:"dependencies"` // Task IDs that must complete first
}

func NewScheduler(redis *redis.Client) *Scheduler {
	return &Scheduler{
		redis: redis,
	}
}

func (s *Scheduler) ScheduleTask(ctx context.Context, task *Task, opts *ScheduleOptions) error {
	// Set task metadata
	task.Priority = opts.Priority
	if opts.Deadline != nil {
		task.Deadline = opts.Deadline
	}
	task.MaxRetries = opts.MaxRetries
	task.Dependencies = opts.Dependencies

	// Check if all dependencies are complete
	if len(task.Dependencies) > 0 {
		for _, depID := range task.Dependencies {
			exists, err := s.redis.HExists(ctx, "results", depID).Result()
			if err != nil {
				return fmt.Errorf("failed to check dependency %s: %w", depID, err)
			}
			if !exists {
				// Add to waiting list
				return s.scheduleDependentTask(ctx, task)
			}
		}
	}

	// Encode task
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Add to appropriate priority queue
	score := float64(time.Now().Unix())
	if task.Deadline != nil {
		// Adjust score based on deadline
		remaining := time.Until(*task.Deadline)
		if remaining < 0 {
			score -= 1000000 // Overdue tasks get highest priority
		} else {
			score -= float64(remaining.Seconds())
		}
	}

	queueKey := fmt.Sprintf("tasks:priority:%d", task.Priority)
	err = s.redis.ZAdd(ctx, queueKey, &redis.Z{
		Score:  score,
		Member: taskBytes,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to queue task: %w", err)
	}

	return nil
}

func (s *Scheduler) scheduleDependentTask(ctx context.Context, task *Task) error {
	taskBytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	// Store task in waiting list
	err = s.redis.HSet(ctx,
		fmt.Sprintf("tasks:waiting:%s", task.ID),
		"task",
		taskBytes,
	).Err()

	if err != nil {
		return fmt.Errorf("failed to schedule dependent task: %w", err)
	}

	// Add task ID to dependency tracking
	for _, depID := range task.Dependencies {
		err = s.redis.SAdd(ctx,
			fmt.Sprintf("tasks:dependencies:%s", depID),
			task.ID,
		).Err()
		if err != nil {
			return fmt.Errorf("failed to track dependency: %w", err)
		}
	}

	return nil
}

func (s *Scheduler) GetNextTask(ctx context.Context) (*Task, error) {
	// Try to get tasks from highest to lowest priority
	for priority := 10; priority > 0; priority-- {
		queueKey := fmt.Sprintf("tasks:priority:%d", priority)

		// Get oldest task in this priority queue
		result, err := s.redis.ZPopMin(ctx, queueKey).Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if len(result) > 0 {
			var task Task
			if err := json.Unmarshal([]byte(result[0].Member.(string)), &task); err != nil {
				return nil, err
			}
			return &task, nil
		}
	}

	return nil, redis.Nil
}

func (s *Scheduler) OnTaskComplete(ctx context.Context, taskID string) error {
	// Get dependent tasks
	dependentIDs, err := s.redis.SMembers(ctx,
		fmt.Sprintf("tasks:dependencies:%s", taskID),
	).Result()
	if err != nil {
		return fmt.Errorf("failed to get dependent tasks: %w", err)
	}

	// Check each dependent task
	for _, depTaskID := range dependentIDs {
		taskKey := fmt.Sprintf("tasks:waiting:%s", depTaskID)

		// Get task data
		taskBytes, err := s.redis.HGet(ctx, taskKey, "task").Result()
		if err != nil {
			continue
		}

		var task Task
		if err := json.Unmarshal([]byte(taskBytes), &task); err != nil {
			continue
		}

		// Check if all dependencies are now complete
		allComplete := true
		for _, depID := range task.Dependencies {
			exists, err := s.redis.HExists(ctx, "results", depID).Result()
			if err != nil || !exists {
				allComplete = false
				break
			}
		}

		if allComplete {
			// Remove from waiting list
			s.redis.Del(ctx, taskKey)

			// Schedule task
			opts := &ScheduleOptions{
				Priority:   task.Priority,
				Deadline:   task.Deadline,
				MaxRetries: task.MaxRetries,
			}
			s.ScheduleTask(ctx, &task, opts)
		}
	}

	// Cleanup dependency tracking
	s.redis.Del(ctx, fmt.Sprintf("tasks:dependencies:%s", taskID))
	return nil
}

func (s *Scheduler) RetryTask(ctx context.Context, task *Task) error {
	if task.RetryCount >= task.MaxRetries {
		return fmt.Errorf("max retries exceeded for task %s", task.ID)
	}

	task.RetryCount++
	task.Status = StatusPending
	task.UpdatedAt = time.Now()

	// Add exponential backoff delay
	backoff := time.Duration(1<<task.RetryCount) * time.Second
	task.NextRetryAt = time.Now().Add(backoff)

	return s.ScheduleTask(ctx, task, &ScheduleOptions{
		Priority:   task.Priority,
		Deadline:   task.Deadline,
		MaxRetries: task.MaxRetries,
	})
}
