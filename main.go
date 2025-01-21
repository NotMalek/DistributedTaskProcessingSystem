package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/coordinator"
	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/task"
	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/worker"
	"github.com/go-redis/redis/v8"
)

type Config struct {
	Command     string
	Role        string
	RedisURL    string
	WorkerCount int
	Monitor     bool
}

func parseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.Command, "command", "", "Command to execute (run/submit)")
	flag.StringVar(&cfg.Role, "role", "", "Service role (coordinator/worker)")
	flag.StringVar(&cfg.RedisURL, "redis", "localhost:6379", "Redis connection URL")
	flag.IntVar(&cfg.WorkerCount, "workers", 5, "Number of worker goroutines")
	flag.BoolVar(&cfg.Monitor, "monitor", false, "Monitor task progress after submission")
	flag.Parse()

	if cfg.Command == "" {
		log.Fatal("Command must be specified (run/submit)")
	}
	if cfg.Command == "run" && cfg.Role == "" {
		log.Fatal("Role must be specified for run command")
	}
	return cfg
}

func main() {
	cfg := parseFlags()
	switch cfg.Command {
	case "run":
		runService(cfg)
	case "submit":
		submitAndMonitor(cfg)
	default:
		log.Fatalf("Unknown command: %s", cfg.Command)
	}
}

func runService(cfg *Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	switch cfg.Role {
	case "coordinator":
		runCoordinator(ctx, cfg, logger)
	case "worker":
		runWorker(ctx, cfg, logger)
	}
}

func runCoordinator(ctx context.Context, cfg *Config, logger *log.Logger) {
	coord := coordinator.New(
		coordinator.WithLogger(logger),
		coordinator.WithRedis(cfg.RedisURL),
	)

	if err := coord.Start(ctx); err != nil {
		logger.Fatalf("Coordinator failed: %v", err)
	}
}

func runWorker(ctx context.Context, cfg *Config, logger *log.Logger) {
	w := worker.New(
		worker.WithLogger(logger),
		worker.WithRedis(cfg.RedisURL),
		worker.WithPoolSize(cfg.WorkerCount),
	)

	if err := w.Start(ctx); err != nil {
		logger.Fatalf("Worker failed: %v", err)
	}
}

func submitAndMonitor(cfg *Config) {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	defer rdb.Close()

	newTask := task.NewTask("test", []byte("hello world"))
	newTask.ComplexityScore = 2

	taskBytes, err := json.Marshal(newTask)
	if err != nil {
		log.Fatalf("Failed to marshal task: %v", err)
	}

	err = rdb.RPush(context.Background(), "tasks", taskBytes).Err()
	if err != nil {
		log.Fatalf("Failed to submit task: %v", err)
	}

	fmt.Printf("Successfully submitted task: %s\n", newTask.ID)

	if !cfg.Monitor {
		return
	}

	fmt.Println("\nMonitoring task progress...")
	for i := 0; i < 30; i++ {
		workers, _ := rdb.HGetAll(context.Background(), "workers").Result()
		queueLen, _ := rdb.LLen(context.Background(), "tasks").Result()
		results, _ := rdb.HGetAll(context.Background(), "results").Result()

		fmt.Printf("\nActive workers: %d\n", len(workers))
		fmt.Printf("Tasks in queue: %d\n", queueLen)
		fmt.Printf("Completed tasks: %d\n", len(results))

		for taskID, resultStr := range results {
			var result task.Result
			if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
				continue
			}
			fmt.Printf("Task %s completed in %.2f seconds\n",
				taskID,
				result.EndTime.Sub(result.StartTime).Seconds())
		}

		time.Sleep(1 * time.Second)
	}
}
