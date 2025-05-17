package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// initLogger initializes the Zap logger with structured logging and UTC-3 timezone.
func initLogger(serviceName, version, environment string) {
	config := zap.NewProductionEncoderConfig()
	// Configure timestamp format with UTC-3 offset
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05-03:00") // RFC3339 with UTC-3 offset
	config.EncodeLevel = zapcore.CapitalLevelEncoder                             // Remove color encoding from level

	encoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), zapcore.InfoLevel)

	logger = zap.New(core, zap.AddCaller()).With(
		zap.String("service", serviceName),
		zap.String("version", version),
		zap.String("environment", environment),
	)

	// Replace the global logger for standard library compatibility
	zap.ReplaceGlobals(logger)
}

func main() {
	// Get configuration from environment variables
	port := getEnv("PORT", "8081")
	environment := getEnv("ENVIRONMENT", "development")
	version := getEnv("VERSION", "unknown")
	schedulerIntervalStr := getEnv("SCHEDULER_INTERVAL", "1")

	// Initialize the logger with service context
	initLogger("worker", version, environment)

	logger.Info("service_started",
		zap.String("port", port),
		zap.String("scheduler_interval", schedulerIntervalStr),
	)

	// Parse scheduler interval
	schedulerInterval, err := strconv.Atoi(schedulerIntervalStr)
	if err != nil || schedulerInterval <= 0 {
		logger.Error("invalid_scheduler_interval",
			zap.Error(err),
			zap.String("interval_string", schedulerIntervalStr),
			zap.String("message", "using default interval of 1 minute"),
		)
		schedulerInterval = 1 // Default to 1 minute if invalid
	}

	// --- Simple Scheduler using Ticker ---
	ticker := time.NewTicker(time.Duration(schedulerInterval) * time.Minute)
	done := make(chan bool)

	go func() {
		logger.Info("scheduler_started", zap.Int("interval_minutes", schedulerInterval))
		// Execute the task immediately on startup
		executeCleanupJob()
		for {
			select {
			case <-ticker.C:
				executeCleanupJob()
			case <-done:
				logger.Info("scheduler_stopped")
				return
			}
		}
	}()
	// --- End Simple Scheduler ---

	// --- Health Check Server for Probes ---
	// Create a new HTTP server for health checks
	healthRouter := http.NewServeMux()

	// Health check endpoints following Kubernetes best practices
	healthRouter.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		// Liveness check should be fast and not check external dependencies
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	healthRouter.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Readiness check can verify external dependencies
		// For now, we'll just return OK, but you could add checks for:
		// - Database connection
		// - Cache connection
		// - External service dependencies
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	healthRouter.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Startup check is similar to liveness but with higher threshold
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	healthServerAddr := fmt.Sprintf(":%s", port)
	logger.Info("starting_health_server", zap.String("addr", healthServerAddr))

	// Start the health server in a goroutine
	go func() {
		if err := http.ListenAndServe(healthServerAddr, healthRouter); err != nil && err != http.ErrServerClosed {
			logger.Fatal("health_server_failed", zap.Error(err))
		}
	}()
	// --- End Health Check Server ---

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Stop the ticker and signal done to the scheduler goroutine
	ticker.Stop()
	done <- true

	logger.Info("service_shutting_down",
		zap.String("service", "worker"),
	)

	// TODO: Add logic here to gracefully stop any long-running worker tasks if necessary

	logger.Info("service_stopped",
		zap.String("service", "worker"),
	)
}

// executeCleanupJob simulates the cron task.
func executeCleanupJob() {
	logger.Info("cleanup_job_started", zap.String("job", "cleanup"))
	start := time.Now()
	// Simulate a cleanup task
	time.Sleep(5 * time.Second)
	duration := time.Since(start).Seconds()
	logger.Info("cleanup_job_completed",
		zap.String("job", "cleanup"),
		zap.Float64("duration", duration),
	)
}
