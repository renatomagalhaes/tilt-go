package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

func initLogger() error {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Configure timezone for timestamps
	loc := time.FixedZone("UTC-3", -3*60*60)
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(loc).Format(time.RFC3339))
	}

	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.NameKey = "logger"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.FunctionKey = "function"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	var err error
	logger, err = config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	return nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Simple HTTP server that responds with "Hello, World!"
func main() {
	if err := initLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("service_started",
		zap.String("service", "api"),
		zap.String("version", "1.0.0"),
		zap.String("environment", "production"),
		zap.String("port", port),
	)

	// Define handler for root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request_received",
			zap.String("service", "api"),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
		)
		fmt.Fprintf(w, "Hello, World Tilt!")
	})

	// Add health check endpoint
	http.HandleFunc("/health", healthCheckHandler)

	// Start server in a goroutine
	go func() {
		logger.Info("server_starting",
			zap.String("service", "api"),
			zap.String("port", port),
		)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
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
