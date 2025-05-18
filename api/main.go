package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/renatomagalhaes/tilt-go/internal/database"
)

var (
	logger          *zap.Logger
	db              *database.DB
	memcachedClient *memcache.Client
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

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// livezHandler handles liveness probe
func livezHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyzHandler handles readiness probe
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := db.CheckConnection(); err != nil {
		logger.Error("database_connection_failed",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database connection failed"))
		return
	}
	// Check Memcached connection
	if memcachedClient != nil {
		if err := memcachedClient.Ping(); err != nil {
			logger.Error("memcached_connection_failed",
				zap.Error(err),
			)
			// Note: We don't necessarily fail readiness if cache is down, depending on policy.
			// For now, we'll allow it to be ready as DB is primary source.
			// w.WriteHeader(http.StatusServiceUnavailable)
			// w.Write([]byte("Memcached connection failed"))
			// return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// healthzHandler handles startup probe
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := db.CheckConnection(); err != nil {
		logger.Error("database_connection_failed",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database connection failed"))
		return
	}
	// Check Memcached connection
	if memcachedClient != nil {
		if err := memcachedClient.Ping(); err != nil {
			logger.Error("memcached_connection_failed",
				zap.Error(err),
			)
			// Note: We don't necessarily fail healthz if cache is down, depending on policy.
			// For now, we'll allow it to be healthy as DB is primary source.
			// w.WriteHeader(http.StatusServiceUnavailable)
			// w.Write([]byte("Memcached connection failed"))
			// return
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// quoteHandler handles the random quote endpoint
// This handler implements a cache-aside pattern:
// 1. Try to get data from cache first
// 2. If cache miss or error, get from database
// 3. If database fetch succeeds, update cache for future requests
func quoteHandler(w http.ResponseWriter, r *http.Request) {
	// Start timing the entire request for performance monitoring
	startRequest := time.Now()
	logger.Info("request_received",
		zap.String("endpoint", "/quotes/random"),
		zap.String("method", r.Method),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Cache key is constant as we're caching a batch of quotes
	const cacheKey = "quote_batch"
	var quotes []database.Quote

	// Step 1: Try to get quotes from cache
	// This is the first attempt to get data, as cache is faster than database
	startCache := time.Now()
	logger.Debug("attempting_cache_read",
		zap.String("cache_key", cacheKey),
	)
	item, err := memcachedClient.Get(cacheKey)
	cacheDuration := time.Since(startCache)

	if err == nil && item != nil {
		// Cache hit: We found data in the cache
		logger.Info("cache_hit",
			zap.String("endpoint", "/quotes/random"),
			zap.Duration("cache_read_duration", cacheDuration),
		)
		// Unmarshal the cached JSON data into our quotes slice
		logger.Debug("unmarshaling_cached_quotes",
			zap.Int("cached_data_size", len(item.Value)),
		)
		err = json.Unmarshal(item.Value, &quotes)
		if err != nil {
			// If we can't parse the cached data, fall back to database
			logger.Error("failed_to_unmarshal_cached_quotes",
				zap.Error(err),
			)
			goto fetchFromDB
		}

		// We have valid quotes from cache, select one randomly
		if len(quotes) > 0 {
			// Use current time as seed for better randomness
			rand.Seed(time.Now().UnixNano())
			selectedIndex := rand.Intn(len(quotes))
			quote := quotes[selectedIndex]
			logger.Debug("selected_random_quote_from_cache",
				zap.Int("total_quotes", len(quotes)),
				zap.Int("selected_index", selectedIndex),
			)
			totalDuration := time.Since(startRequest)
			logger.Info("quote_retrieved_successfully",
				zap.String("endpoint", "/quotes/random"),
				zap.String("method", r.Method),
				zap.Int("quote_id", quote.ID),
				zap.String("source", "cache"),
				zap.Int("status_code", http.StatusOK),
				zap.Duration("cache_read_duration", cacheDuration),
				zap.Duration("total_request_duration", totalDuration),
			)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(quote)
			return
		} else {
			// Cache hit but no quotes found (empty batch)
			logger.Warn("cached_quote_batch_empty",
				zap.String("endpoint", "/quotes/random"),
			)
			goto fetchFromDB
		}
	} else {
		// Cache miss: No data in cache or error occurred
		logger.Info("cache_miss",
			zap.String("endpoint", "/quotes/random"),
			zap.Error(err), // err will indicate the reason (e.g., ErrCacheMiss)
			zap.Duration("cache_read_duration", cacheDuration),
		)
	}

fetchFromDB:
	// Step 2: Fallback to database when cache fails
	logger.Info("fetching_quotes_from_db",
		zap.String("endpoint", "/quotes/random"),
	)
	startDB := time.Now()

	// Get batch size from environment or use default
	// This controls how many quotes we fetch at once
	quoteBatchSize := getEnvInt("QUOTE_BATCH_SIZE", 500)
	logger.Debug("fetching_quotes_from_db",
		zap.Int("batch_size", quoteBatchSize),
	)
	dbQuotes, err := db.GetRandomQuotes(quoteBatchSize)
	dbDuration := time.Since(startDB)

	if err != nil {
		// Database error - return 500
		logger.Error("failed_to_get_quote_from_db",
			zap.Error(err),
			zap.String("endpoint", "/quotes/random"),
			zap.String("method", r.Method),
			zap.Duration("database_call_duration", dbDuration),
			zap.Duration("total_request_duration", time.Since(startRequest)),
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get quote"})
		return
	}

	// If DB returned quotes, pick one randomly
	var quote database.Quote
	if len(dbQuotes) > 0 {
		// Select a random quote from the database results
		rand.Seed(time.Now().UnixNano())
		selectedIndex := rand.Intn(len(dbQuotes))
		quote = dbQuotes[selectedIndex]
		logger.Debug("selected_random_quote_from_db",
			zap.Int("total_quotes", len(dbQuotes)),
			zap.Int("selected_index", selectedIndex),
		)
		totalDuration := time.Since(startRequest)

		logger.Info("quote_retrieved_successfully",
			zap.String("endpoint", "/quotes/random"),
			zap.String("method", r.Method),
			zap.Int("quote_id", quote.ID),
			zap.String("source", "database"),
			zap.Int("status_code", http.StatusOK),
			zap.Duration("database_call_duration", dbDuration),
			zap.Duration("total_request_duration", totalDuration),
		)

		// Step 3: Update cache with new quotes for future requests
		// This is done in a goroutine to not block the response
		go func() {
			logger.Debug("preparing_to_cache_quotes",
				zap.Int("quotes_to_cache", len(dbQuotes)),
			)
			// Convert quotes to JSON for caching
			quotedata, marshalErr := json.Marshal(dbQuotes)
			if marshalErr != nil {
				logger.Error("failed_to_marshal_quotes_for_cache_on_miss",
					zap.Error(marshalErr),
				)
				return
			}
			// Get cache TTL from environment or use default (5 minutes)
			quoteCacheTTL := getEnvInt("QUOTE_CACHE_TTL_SECONDS", 300)
			item := &memcache.Item{
				Key:        cacheKey,
				Value:      quotedata,
				Expiration: int32(quoteCacheTTL),
			}
			logger.Debug("setting_cache_item",
				zap.Int("data_size", len(quotedata)),
				zap.Int32("ttl_seconds", item.Expiration),
			)
			if setErr := memcachedClient.Set(item); setErr != nil {
				logger.Error("failed_to_set_quote_cache_on_miss",
					zap.Error(setErr),
				)
			} else {
				logger.Info("cache_updated_with_new_quotes",
					zap.Int("batch_size", len(dbQuotes)),
					zap.Int("cache_ttl_seconds", quoteCacheTTL),
				)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quote)
		return
	}

	// No quotes found in database - return 404
	totalDuration := time.Since(startRequest)
	logger.Error("no_quotes_found",
		zap.String("endpoint", "/quotes/random"),
		zap.String("method", r.Method),
		zap.Duration("total_request_duration", totalDuration),
	)
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "No quotes found"})
}

// Simple HTTP server that responds with "Hello, World!"
func main() {
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	version := getEnv("VERSION", "unknown")

	// Initialize the logger with service context
	initLogger("api", version, environment)

	// Initialize Memcached client
	memcachedHost := getEnv("MEMCACHED_HOST", "memcached")
	memcachedPort := getEnv("MEMCACHED_PORT", "11211")
	memcachedClient = memcache.New(fmt.Sprintf("%s:%s", memcachedHost, memcachedPort))
	logger.Info("memcached_client_initialized", zap.String("memcached_addr", fmt.Sprintf("%s:%s", memcachedHost, memcachedPort)))

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

	logger.Info("service_started",
		zap.String("port", port),
	)

	rand.Seed(time.Now().UnixNano())

	router := chi.NewRouter()

	// Health check endpoints following Kubernetes best practices
	router.Get("/livez", livezHandler)     // Liveness probe
	router.Get("/readyz", readyzHandler)   // Readiness probe
	router.Get("/healthz", healthzHandler) // Startup probe

	// Define handler for root endpoint
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request_received",
			zap.String("service", "api"),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
		)
		fmt.Fprintf(w, "Hello, World Tilt!")
	})

	// Add quote endpoint
	router.Get("/quotes/random", quoteHandler)

	// Start server in a goroutine
	go func() {
		logger.Info("server_starting",
			zap.String("service", "api"),
			zap.String("port", port),
		)
		serverAddr := fmt.Sprintf(":%s", port)
		if err := http.ListenAndServe(serverAddr, router); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server_failed",
				zap.String("service", "api"),
				zap.Error(err),
			)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("service_shutting_down",
		zap.String("service", "api"),
	)

	logger.Info("service_stopped",
		zap.String("service", "api"))
}
