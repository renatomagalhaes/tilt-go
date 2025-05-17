package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
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

func main() {
	if err := initLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get configuration from environment variables
	port := getEnv("PORT", "8081")
	environment := getEnv("ENVIRONMENT", "development")
	version := getEnv("VERSION", "1.0.0")
	schedulerInterval, _ := strconv.Atoi(getEnv("SCHEDULER_INTERVAL", "1"))

	logger.Info("service_started",
		zap.String("service", "worker"),
		zap.String("version", version),
		zap.String("environment", environment),
		zap.String("port", port),
		zap.Int("scheduler_interval", schedulerInterval),
	)

	// Cria o scheduler
	loc := time.FixedZone("UTC-3", -3*60*60)
	scheduler := gocron.NewScheduler(loc)

	// Agenda a tarefa de limpeza
	_, err := scheduler.Every(schedulerInterval).Minutes().Do(func() {
		logger.Info("cleanup_job_started",
			zap.String("service", "worker"),
			zap.String("job", "cleanup"),
			zap.String("scheduled_time", time.Now().UTC().Format(time.RFC3339)),
		)

		start := time.Now()
		if err := cleanupOldData(); err != nil {
			logger.Error("cleanup_job_failed",
				zap.String("service", "worker"),
				zap.String("job", "cleanup"),
				zap.Error(err),
				zap.Duration("duration", time.Since(start)),
			)
			return
		}

		logger.Info("cleanup_job_completed",
			zap.String("service", "worker"),
			zap.String("job", "cleanup"),
			zap.Duration("duration", time.Since(start)),
		)
	})

	if err != nil {
		logger.Fatal("scheduler_initialization_failed",
			zap.String("service", "worker"),
			zap.Error(err),
		)
	}

	// Inicia o scheduler
	scheduler.StartAsync()

	logger.Info("scheduler_started",
		zap.String("service", "worker"),
		zap.Int("jobs_count", len(scheduler.Jobs())),
	)

	// Inicie o servidor HTTP para o health check
	http.HandleFunc("/health", healthCheckHandler)
	go http.ListenAndServe(":"+port, nil)

	// Mantém o processo rodando
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("service_shutting_down",
		zap.String("service", "worker"),
	)

	// Para o scheduler
	scheduler.Stop()

	logger.Info("service_stopped",
		zap.String("service", "worker"),
	)
}

func cleanupOldData() error {
	// Simula uma tarefa de limpeza
	time.Sleep(5 * time.Second)
	return nil
}
