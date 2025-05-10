package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
	exporter            *prometheus.Exporter
)

// Função que simula carga de CPU e memória
func simulateLoad() float64 {
	// Aloca um slice menor para simular uso de memória
	data := make([]float64, 100000)

	// Preenche o slice com números aleatórios
	for i := range data {
		data[i] = rand.Float64()
	}

	// Realiza cálculos intensivos
	var result float64
	for i := 0; i < 100; i++ {
		for _, v := range data {
			result += math.Sqrt(v) * math.Sin(v) * math.Cos(v)
		}
	}

	// Força GC para simular pressão na memória
	data = nil
	runtime.GC()

	return result
}

// Função que simula erros aleatórios
func simulateError() (int, string) {
	errors := map[int]string{
		500: "Internal Server Error",
		502: "Bad Gateway",
		503: "Service Unavailable",
		504: "Gateway Timeout",
	}

	// Escolhe um erro aleatório
	codes := make([]int, 0, len(errors))
	for code := range errors {
		codes = append(codes, code)
	}

	statusCode := codes[rand.Intn(len(codes))]
	return statusCode, errors[statusCode]
}

func initMetrics() error {
	var err error
	exporter, err = prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(meterProvider)

	meter := otel.GetMeterProvider().Meter("api-service")

	httpRequestsTotal, err = meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_requests_total counter: %w", err)
	}

	httpRequestDuration, err = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http_request_duration_seconds histogram: %w", err)
	}

	return nil
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture the status code
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()

		// Add status code to the metrics
		httpRequestsTotal.Add(context.Background(), 1, metric.WithAttributes(
			attribute.String("status", fmt.Sprintf("%d", rw.statusCode)),
		))
		httpRequestDuration.Record(context.Background(), duration)
	})
}

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Simple HTTP server that responds with "Hello, World!"
func main() {
	if err := initMetrics(); err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	r := mux.NewRouter()
	r.Use(metricsMiddleware)

	// Define handler for root endpoint
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World Tilt@!")
	})

	// Define handler for load test endpoint
	r.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		result := simulateLoad()
		fmt.Fprintf(w, "Load test completed. Result: %f", result)
	})

	// Define handler for error simulation endpoint
	r.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		statusCode, message := simulateError()
		http.Error(w, message, statusCode)
	})

	// Expose metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Printf("Starting API server on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
