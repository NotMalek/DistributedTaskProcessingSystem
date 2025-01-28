package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/api"
	"github.com/NotMalek/DistributedTaskProcessingSystem/internal/coordinator"
	"github.com/go-redis/redis/v8"
)

type Config struct {
	RedisURL string
	APIPort  string
}

func main() {
	cfg := &Config{}
	flag.StringVar(&cfg.RedisURL, "redis", "localhost:6379", "Redis connection URL")
	flag.StringVar(&cfg.APIPort, "port", "8080", "API server port")
	flag.Parse()

	// Setup logger
	logger := log.New(os.Stdout, "[Server] ", log.LstdFlags)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create API server
	apiServer := api.NewServer(rdb)

	// Create coordinator
	coord := coordinator.New(
		coordinator.WithLogger(log.New(os.Stdout, "[Coordinator] ", log.LstdFlags)),
		coordinator.WithRedis(cfg.RedisURL),
	)

	// WaitGroup to manage components
	var wg sync.WaitGroup
	wg.Add(2) // API server and coordinator

	// Start API server
	go func() {
		defer wg.Done()
		if err := apiServer.Start(":" + cfg.APIPort); err != nil {
			logger.Printf("API server error: %v", err)
			cancel()
		}
	}()

	// Start coordinator
	go func() {
		defer wg.Done()
		if err := coord.Start(ctx); err != nil {
			logger.Printf("Coordinator error: %v", err)
			cancel()
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Println("Shutdown signal received, gracefully stopping...")
		cancel()
	}()

	// Wait for all components to shut down
	wg.Wait()
	logger.Println("Server stopped")
}
