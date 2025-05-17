package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/renatomagalhaes/tilt-go/api/internal/database"
)

var (
	logger *zap.Logger
	db     *database.DB
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func initLogger(serviceName, version, environment string) {
	config := zap.NewProductionEncoderConfig()
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// quoteHandler handles the random quote endpoint
func quoteHandler(w http.ResponseWriter, r *http.Request) {
	// Log request received
	logger.Info("request_received",
		zap.String("endpoint", "/quotes/random"),
		zap.String("method", r.Method),
		zap.String("remote_addr", r.RemoteAddr),
	)

	// Measure database call duration
	startDB := time.Now()
	quote, err := db.GetRandomQuote()
	dbDuration := time.Since(startDB)

	if err != nil {
		// Log the error with context
		logger.Error("failed_to_get_quote",
			zap.Error(err),
			zap.String("endpoint", "/quotes/random"),
			zap.String("method", r.Method),
			zap.Duration("database_call_duration", dbDuration), // Log DB duration even on error
		)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get quote"})
		return
	}

	// Log success
	logger.Info("quote_retrieved_successfully",
		zap.String("endpoint", "/quotes/random"),
		zap.String("method", r.Method),
		zap.Int("quote_id", quote.ID), // Log the ID
		zap.Duration("database_call_duration", dbDuration),
		zap.Int("status_code", http.StatusOK), // Log the status code
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quote)
}

// Simple HTTP server that responds with "Hello, World!"
func main() {
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	version := getEnv("VERSION", "unknown")

	// Initialize the logger with service context
	initLogger("api", version, environment)

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
		if err := http.ListenAndServe(serverAddr, router); err != nil {
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
		zap.String("service", "api"),
	)
}
