# Distributed Task Processing System

A sophisticated distributed task processing system built in Go, featuring auto-scaling workers, work stealing, priority-based scheduling, and comprehensive monitoring capabilities.

## Features

- **Advanced Worker Management**
    - Auto-scaling worker pools
    - Work stealing between workers
    - Health monitoring and heartbeats
    - Graceful shutdown handling

- **Intelligent Task Processing**
    - Priority-based task scheduling
    - Task dependencies support
    - Retry mechanism with backoff
    - Deadlines and timeouts

- **Observability**
    - OpenTelemetry integration
    - Real-time metrics collection
    - Web-based monitoring dashboard
    - Distributed tracing

- **Robust Architecture**
    - Redis-backed task distribution
    - Configurable via YAML and environment
    - Clean separation of concerns
    - Production-ready error handling

## Architecture

```
┌─────────────┐     ┌─────────────┐
│  Coordinator│◄────┤Task Producer│
└─────┬───────┘     └─────────────┘
      │
   Redis Queue
      │
      ▼
┌─────────────┐     ┌─────────────┐
│   Worker 1  │     │   Worker 2  │
└─────────────┘     └─────────────┘
```

## Prerequisites

- Go 1.23 or later
- Redis server

## Installation

1. Clone the repository:
```bash
git clone https://github.com/NotMalek/DistributedTaskProcessingSystem.git
cd DistributedTaskProcessingSystem
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the project:
```bash
go build
```

## Running the System

The system supports three main commands:

### 1. Start Redis
```bash
redis-server
```

### 2. Start Coordinator
```bash
# Basic start
go run main.go -command run -role coordinator -redis localhost:6379

# With monitoring dashboard
go run main.go -command run -role coordinator -redis localhost:6379 -monitor
```

### 3. Start Workers
```bash
# Start a worker with default settings
go run main.go -command run -role worker -redis localhost:6379

# Start a worker with custom pool size
go run main.go -command run -role worker -redis localhost:6379 -workers 5

# Start a worker with work stealing enabled
go run main.go -command run -role worker -redis localhost:6379 -steal
```

### 4. Submit Tasks
```bash
# Submit a simple task
go run main.go -command submit -redis localhost:6379

# Submit with monitoring
go run main.go -command submit -redis localhost:6379 -monitor

# Submit a high-priority task
go run main.go -command submit -redis localhost:6379 -priority 10

# Submit task with deadline
go run main.go -command submit -redis localhost:6379 -deadline "2024-01-22T15:04:05Z"
```

## Configuration

### Command Line Flags

### Global Flags:

- **`-command`**: Required. `"run"` or `"submit"`
- **`-redis`**: Redis URL (default: `localhost:6379`)
- **`-monitor`**: Enable monitoring (default: `false`)

### Run Command Flags:

- **`-role`**: Required. `"coordinator"` or `"worker"`
- **`-workers`**: Worker pool size (default: `5`)
- **`-steal`**: Enable work stealing (default: `false`)
- **`-min`**: Minimum workers for auto-scaling (default: `1`)
- **`-max`**: Maximum workers for auto-scaling (default: `10`)

### Submit Command Flags:

- **`-priority`**: Task priority (1-10, default: `1`)
- **`-deadline`**: Task deadline (RFC3339 format)
- **`-retries`**: Maximum retry attempts (default: `3`)

## System Components

### Coordinator
- Manages task distribution
- Monitors worker health
- Handles task reassignment for failed workers
- Collects and aggregates results

### Worker
- Processes assigned tasks
- Sends heartbeat signals
- Reports task completion status
- Manages local worker pool

### Task Flow
1. Tasks are submitted via command line
2. Coordinator queues tasks in Redis
3. Workers poll for available tasks
4. Workers process tasks and submit results
5. Coordinator collects and stores results

## Project Structure

```
.
├── internal/
|   ├── api/          # Configuration management
|       └──server.go
│   ├── config/          # Configuration management
|       └──config.go
│   ├── coordinator/     # Coordinator implementation
|       └──coordinator.go
│   ├── task/           # Task definitions and scheduling
│       ├── scheduler.go
│       └── task.go
│   ├── telemetry/      # Tracing and metrics
│       ├── middleware.go
│       └── setup.go
│   └── worker/         # Worker implementation
│       ├── autoscaler.go
│       ├── metrics.go
│       ├── stealing.go
│       └── worker.go
├── main.go
└── README.md
```
## Dashboard

![image](https://github.com/user-attachments/assets/0c9fd0f2-e135-4338-b111-84ee4db23d0b)

