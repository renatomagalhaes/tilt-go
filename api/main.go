package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
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
	// Liveness check should be fast and not check external dependencies
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// readyzHandler handles readiness probe
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	// Readiness check can verify external dependencies
	// For now, we'll just return OK, but you could add checks for:
	// - Database connection
	// - Cache connection
	// - External service dependencies
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// healthzHandler handles startup probe
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Startup check is similar to liveness but with higher threshold
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Simple HTTP server that responds with "Hello, World!"
func main() {
	port := getEnv("PORT", "8080")
	environment := getEnv("ENVIRONMENT", "development")
	version := getEnv("VERSION", "unknown")

	// Initialize the logger with service context
	initLogger("api", version, environment)

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
