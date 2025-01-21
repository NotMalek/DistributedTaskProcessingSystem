# Distributed Task Processing System

A scalable distributed task processing system built in Go, featuring a coordinator-worker architecture with Redis-based task distribution and Prometheus metrics monitoring.

## Features

- **Distributed Architecture**: Coordinator-worker pattern for scalable task processing
- **Redis Backend**: Reliable task queue and result storage using Redis
- **Metric Monitoring**: Prometheus metrics for system monitoring
- **Graceful Shutdown**: Clean shutdown with proper cleanup
- **Configurable Workers**: Adjustable worker pool size for performance tuning
- **Health Monitoring**: Worker heartbeat monitoring and automatic task reassignment

## Prerequisites

- Go 1.23 or later
- Redis server
- Prometheus (optional, for metrics)

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

## Configuration

The system can be configured using command-line flags:

- `-role`: Required. Either "coordinator" or "worker"
- `-redis`: Redis connection URL (default: "localhost:6379")
- `-metrics-addr`: Metrics server address (default: ":9090")
- `-workers`: Number of worker goroutines (default: 5)

## Running the System

1. Start a Redis server:
```bash
redis-server
```

2. Start the coordinator:
```bash
./DistributedTaskProcessingSystem -role coordinator
```

3. Start one or more workers:
```bash
./DistributedTaskProcessingSystem -role worker -workers 5
```

## Architecture

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
1. Tasks are submitted to the coordinator
2. Coordinator queues tasks in Redis
3. Workers poll for available tasks
4. Workers process tasks and submit results
5. Coordinator collects and stores results

## Metrics

The system exposes Prometheus metrics at `/metrics` including:
- Tasks submitted/assigned/completed
- Active worker count
- Processing time
- Queue time
- Worker pool size

## Project Structure

```
.
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── coordinator/      # Coordinator implementation
│   ├── worker/          # Worker implementation
│   ├── task/            # Task definitions
│   └── metrics/         # Metrics collection
├── go.mod
├── go.sum
└── README.md
```

## Contributing

1. Fork the repository
2. Create a new branch for your feature
3. Commit your changes
4. Push to your branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- The Go team for the excellent language and tools
- Redis for providing a robust message queue
- Prometheus team for the monitoring solution