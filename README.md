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

### 2. Start Backend
```bash
# Basic start
go run main.go

# With custom Redis and port
go run main.go -redis localhost:6379 -port 8080
```

### 3. Start Frontend
```bash
cd .\frontend\

npm run dev
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
|   |   └──server.go
│   ├── config/          # Configuration management
|   |   └──config.go
│   ├── coordinator/     # Coordinator implementation
|   |   └──coordinator.go
│   ├── task/           # Task definitions and scheduling
│   |   ├── scheduler.go
│   |   └── task.go
│   └── worker/         # Worker implementation
│       ├── autoscaler.go
│       ├── metrics.go
│       ├── stealing.go
│       └── worker.go
├── main.go
└── README.md
```
## Dashboard

![image](https://github.com/user-attachments/assets/8f4fcea1-16d9-4e7b-9e34-56261d27ca8f)


