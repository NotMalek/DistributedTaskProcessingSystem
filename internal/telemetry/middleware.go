package telemetry

import (
	"context"
	"time"

	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/task"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type TaskProcessorFunc func(context.Context, *task.Task) (*task.Result, error)

// TracingMiddleware wraps a task processor with OpenTelemetry tracing
func TracingMiddleware(next TaskProcessorFunc) TaskProcessorFunc {
	tracer := otel.Tracer("task_processor")

	return func(ctx context.Context, t *task.Task) (*task.Result, error) {
		ctx, span := tracer.Start(ctx, "process_task",
			trace.WithAttributes(
				attribute.String("task.id", t.ID),
				attribute.String("task.type", t.Type),
				attribute.Int("task.priority", t.Priority),
			),
		)
		defer span.End()

		startTime := time.Now()

		// Process the task
		result, err := next(ctx, t)

		// Record the outcome
		if err != nil {
			span.RecordError(err)
		}

		span.SetAttributes(
			attribute.String("task.status", string(result.Status)),
			attribute.Float64("task.duration_ms", float64(time.Since(startTime).Milliseconds())),
		)

		return result, err
	}
}

// MetricsMiddleware wraps a task processor with metrics collection
func MetricsMiddleware(collector *Collector) func(TaskProcessorFunc) TaskProcessorFunc {
	return func(next TaskProcessorFunc) TaskProcessorFunc {
		return func(ctx context.Context, t *task.Task) (*task.Result, error) {
			startTime := time.Now()

			// Process the task
			result, err := next(ctx, t)

			// Record metrics
			duration := time.Since(startTime)
			collector.RecordTaskProcessed(ctx, duration, err == nil)

			if t.RetryCount > 0 {
				collector.RecordTaskRetry(ctx, t.ID)
			}

			return result, err
		}
	}
}

// ChainMiddleware combines multiple middleware functions
func ChainMiddleware(middlewares ...func(TaskProcessorFunc) TaskProcessorFunc) func(TaskProcessorFunc) TaskProcessorFunc {
	return func(next TaskProcessorFunc) TaskProcessorFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
