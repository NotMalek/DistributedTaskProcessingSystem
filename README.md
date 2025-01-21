# Distributed Task Processing System

A scalable distributed task processing system built in Go, featuring a coordinator-worker architecture with Redis-based task distribution.

## Features

- **Distributed Architecture**: Coordinator-worker pattern for scalable task processing
- **Redis Backend**: Reliable task queue and result storage using Redis
- **Graceful Shutdown**: Clean shutdown with proper cleanup
- **Configurable Workers**: Adjustable worker pool size for performance tuning
- **Health Monitoring**: Worker heartbeat monitoring and automatic task reassignment

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

## Usage

The system supports three main commands:

### 1. Start Redis (Required First)
```bash
redis-server
```

### 2. Run Coordinator
```bash
go run main.go -command run -role coordinator -redis localhost:6379
```

### 3. Run Worker
```bash
go run main.go -command run -role worker -redis localhost:6379 -workers 5
```

### 4. Submit Tasks
```bash
# Submit a single test task
go run main.go -command submit -redis localhost:6379

# Submit and monitor tasks
go run main.go -command submit -redis localhost:6379 -monitor
```

## Configuration

Command-line flags:
- `-command`: Required. One of "run" or "submit"
- `-role`: Required for "run" command. Either "coordinator" or "worker"
- `-redis`: Redis connection URL (default: "localhost:6379")
- `-workers`: Number of worker goroutines (default: 5)
- `-monitor`: For submit command, monitors task progress (default: false)

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
│   ├── coordinator/      # Coordinator implementation
│   ├── worker/          # Worker implementation
│   └── task/            # Task definitions
├── main.go              # Application entry point
├── go.mod
└── README.md
```

## Contributing

Feel free to submit issues, forks, and pull requests.