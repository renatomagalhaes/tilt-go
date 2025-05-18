package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/renatomagalhaes/tilt-go/internal/database"
)

var (
	logger          *zap.Logger
	memcachedClient *memcache.Client
	db              *database.DB
)

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an integer environment variable or returns a default value.
func getEnvInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
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

	// Determine log level based on environment variable
	logLevel := zapcore.InfoLevel
	if os.Getenv("DEBUG_LOGGING") == "true" {
		logLevel = zapcore.DebugLevel
	}

	core := zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), logLevel)

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
	memcachedHost := getEnv("MEMCACHED_HOST", "localhost")
	memcachedPort := getEnv("MEMCACHED_PORT", "11211")
	quoteBatchSize := getEnvInt("QUOTE_BATCH_SIZE", 500)       // Default to 500 quotes
	quoteCacheTTL := getEnvInt("QUOTE_CACHE_TTL_SECONDS", 300) // Default to 5 minutes (300 seconds)

	// Initialize the logger with service context
	initLogger("worker", version, environment)

	logger.Info("service_started",
		zap.String("port", port),
		zap.String("scheduler_interval", schedulerIntervalStr),
		zap.String("memcached_addr", fmt.Sprintf("%s:%s", memcachedHost, memcachedPort)),
		zap.Int("quote_batch_size", quoteBatchSize),
		zap.Int("quote_cache_ttl_seconds", quoteCacheTTL),
	)

	// Initialize Memcached client
	memcachedClient = memcache.New(fmt.Sprintf("%s:%s", memcachedHost, memcachedPort))
	logger.Info("memcached_client_initialized")

	// Initialize database connection
	var err error
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASS", "root")
	dbName := getEnv("DB_NAME", "app")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)
	db, err = database.NewDB(dsn)
	if err != nil {
		logger.Fatal("failed_to_connect_to_database",
			zap.Error(err),
		)
	}
	defer db.Close()

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
	// Use a shorter interval for cache refresh than the cleanup job
	cacheRefreshInterval := time.Duration(quoteCacheTTL) * time.Second
	cacheTicker := time.NewTicker(cacheRefreshInterval)

	ticker := time.NewTicker(time.Duration(schedulerInterval) * time.Minute)
	done := make(chan bool)

	go func() {
		logger.Info("scheduler_started", zap.Int("cleanup_interval_minutes", schedulerInterval), zap.Duration("cache_refresh_interval", cacheRefreshInterval))

		// Execute cache refresh immediately on startup
		refreshQuoteCache(quoteBatchSize, quoteCacheTTL)

		for {
			select {
			case <-ticker.C:
				executeCleanupJob()
			case <-cacheTicker.C:
				refreshQuoteCache(quoteBatchSize, quoteCacheTTL)
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

	// Stop the tickers and signal done to the scheduler goroutine
	ticker.Stop()
	cacheTicker.Stop()
	done <- true

	logger.Info("service_shutting_down",
		zap.String("service", "worker"),
	)

	// TODO: Add logic here to gracefully stop any long-running worker tasks if necessary

	logger.Info("service_stopped",
		zap.String("service", "worker"),
	)
}

// refreshQuoteCache fetches a batch of quotes from the database and caches them in Memcached.
func refreshQuoteCache(batchSize, ttlSeconds int) {
	logger.Info("refreshing_quote_cache", zap.Int("batch_size", batchSize), zap.Int("ttl_seconds", ttlSeconds))

	// Fetch quotes from database
	quotes, err := db.GetRandomQuotes(batchSize)
	if err != nil {
		logger.Error("failed_to_fetch_quotes_for_cache",
			zap.Error(err),
		)
		return
	}

	// Serialize quotes to JSON
	quotedata, err := json.Marshal(quotes)
	if err != nil {
		logger.Error("failed_to_marshal_quotes",
			zap.Error(err),
		)
		return
	}

	// Store in Memcached
	item := &memcache.Item{
		Key:        "quote_batch", // Consistent cache key
		Value:      quotedata,
		Expiration: int32(ttlSeconds),
	}

	if err := memcachedClient.Set(item); err != nil {
		logger.Error("failed_to_set_quote_cache",
			zap.Error(err),
		)
		return
	}

	logger.Info("quote_cache_refreshed", zap.Int("num_quotes", len(quotes)))
}

// executeCleanupJob simulates the cron task.
func executeCleanupJob() {
	logger.Debug("cleanup_job_started", zap.String("job", "cleanup"))
	start := time.Now()
	// Simulate a cleanup task
	time.Sleep(5 * time.Second)
	duration := time.Since(start).Seconds()
	logger.Debug("cleanup_job_completed",
		zap.String("job", "cleanup"),
		zap.Float64("duration", duration),
	)
}
