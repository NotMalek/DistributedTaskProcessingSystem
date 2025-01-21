# Distributed Task Processing System

A high-performance distributed task processing system written in Go, designed to handle large-scale workloads with fault tolerance and monitoring capabilities.

## Features

- **Distributed Architecture**: Coordinator-worker model for scalable task processing
- **Fault Tolerance**: Automatic task reassignment on worker failures
- **Real-time Monitoring**: Prometheus metrics for system observability
- **Redis Backend**: Efficient task queue and result storage
- **Concurrent Processing**: Multiple worker goroutines for optimal performance
- **Production-Ready**: Graceful shutdown, configuration, and logging

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

## Requirements

- Go 1.16+
- Redis server
- Prometheus (optional, for metrics)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/task-processor
cd task-processor
```

2. Install dependencies:
```bash
go mod download
```

## Usage

1. Start Redis server:
```bash
redis-server
```

2. Start the coordinator:
```bash
go run main.go -role coordinator -redis localhost:6379
```

3. Start one or more workers:
```bash
go run main.go -role worker -redis localhost:6379 -workers 5
```

## Configuration

The system can be configured using command-line flags:

- `-role`: Service role (coordinator/worker)
- `-redis`: Redis connection URL (default: localhost:6379)
- `-metrics-addr`: Metrics server address (default: :9090)
- `-workers`: Number of worker goroutines (default: 5)

## Monitoring

Access Prometheus metrics at:
```
http://localhost:9090/metrics
```

Key metrics include:
- Task submission/completion rates
- Processing times
- Queue lengths
- Worker pool status

## Project Structure

```
.
├── main.go                 # Application entry point
├── internal/
│   ├── coordinator/       # Coordinator service
│   ├── worker/           # Worker service
│   ├── task/            # Task definitions
│   └── metrics/         # Prometheus metrics
└── README.md
```

## System Flow

1. Tasks are submitted to the coordinator
2. Coordinator queues tasks in Redis
3. Workers poll for available tasks
4. Tasks are processed concurrently
5. Results are stored back in Redis
6. Coordinator monitors worker health and handles failures

## Contributing

Feel free to submit issues and enhancement requests!