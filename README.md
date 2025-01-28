# Distributed Task Processing System

A sophisticated distributed task processing system built in Go, featuring auto-scaling workers, work stealing, priority-based scheduling, and comprehensive monitoring capabilities.

## Features
- **Advanced Worker Management**
    - Auto-scaling worker pools
    - Work stealing between workers
    - Health monitoring and heartbeats
    - Graceful shutdown handling
- **Intelligent Task Processing**
    - Priority-based task scheduling (1-10)
    - Task deadlines and timeouts
    - Configurable retry mechanism
    - Real-time task status updates
- **Modern Dashboard**
    - Real-time metrics visualization
    - Dark mode interface
    - Interactive worker management
    - Task submission interface
    - Priority queue visualization
- **Robust Architecture**
    - Redis-backed task distribution
    - React-based frontend
    - RESTful API integration
    - Production-ready error handling

## Architecture

```
┌─────────────┐     ┌─────────────┐
│   Next.js   │◄────┤    API      │
│  Frontend   │     │   Server    │
└─────────────┘     └──────┬──────┘
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
- Node.js 18 or later
- npm/yarn

## Installation

1. Clone the repository:
```bash
git clone https://github.com/NotMalek/DistributedTaskProcessingSystem.git
cd DistributedTaskProcessingSystem
```

2. Install backend dependencies:
```bash
go mod tidy
```

3. Install frontend dependencies:
```bash
cd frontend
npm install
# or
yarn install
```

## Running the System

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
cd frontend
npm run dev
# or
yarn dev
```

The dashboard will be available at `http://localhost:3000`

## API Endpoints

### Worker Management
```bash
# Start a new worker
POST /api/workers/start
{
    "poolSize": 5,
    "enableSteal": true,
    "minWorkers": 1,
    "maxWorkers": 10
}

# Stop a worker
POST /api/workers/stop?id={workerId}

# Get worker list and status
GET /api/workers
```

### Task Management
```bash
# Submit a new task
POST /api/tasks/submit
{
    "priority": 5,
    "deadline": "2024-01-30T15:04:05Z",
    "retries": 3,
    "taskType": "test",
    "payload": "task data here"
}

# Get task status
GET /api/tasks/status?id={taskId}
```

### System Management
```bash
# Get system metrics
GET /api/metrics

# Get detailed debug information
GET /api/debug

# Reset the entire system
POST /api/system/reset
```

## Dashboard Features

### Real-time Monitoring
- Active worker count
- Total tasks in system
- Processed tasks count
- Failed tasks count
- Priority queue lengths visualization

### Worker Management
- Add/remove workers
- Configure worker pool size
- Enable/disable task stealing
- Set min/max worker limits
- Monitor worker status and health

### Task Management
- Submit new tasks
- Set task priorities
- Configure deadlines
- Specify retry attempts
- Monitor task status

### System Control
- Reset system state
- Real-time metrics updates
- Error tracking and reporting

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


