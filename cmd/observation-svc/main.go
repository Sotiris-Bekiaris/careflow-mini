package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/internal/observation"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/db"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/observability"
	observationv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/observation/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	tracerProvider, err := observability.InitTracer("observation-svc", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"))
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tracerProvider(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := observability.GetTracer("observation-svc")

	// Initialize metrics
	_, err = observability.InitMetrics("observation-svc")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	// Connect to PostgreSQL
	dbConfig := db.Config{
		Host:              getEnv("DB_HOST", "localhost"),
		Port:              getEnvInt("DB_PORT", 5432),
		User:              getEnv("DB_USER", "postgres"),
		Password:          getEnv("DB_PASSWORD", "postgres"),
		Database:          getEnv("DB_NAME", "careflow"),
		MaxConns:          int32(getEnvInt("DB_MAX_CONNS", 25)),
		MinConns:          int32(getEnvInt("DB_MIN_CONNS", 5)),
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: time.Minute,
	}

	dbPool, err := db.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	log.Println("Connected to PostgreSQL")

	// Initialize NATS publisher
	natsConfig := events.NATSConfig{
		URL:           getEnv("NATS_URL", "nats://localhost:4222"),
		StreamName:    "CAREFLOW_EVENTS",
		MaxReconnects: 10,
	}

	publisher, err := events.NewNATSPublisher(natsConfig)
	if err != nil {
		log.Fatalf("Failed to create NATS publisher: %v", err)
	}
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Error closing publisher: %v", err)
		}
	}()

	log.Println("Connected to NATS")

	// Initialize observation components
	repo := observation.NewPostgresRepository(dbPool)
	svc := observation.NewService(repo, publisher, tracer)
	handler := observation.NewHandler(svc)

	// Setup gRPC server
	port := getEnv("GRPC_PORT", "50055")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// Register services
	observationv1.RegisterObservationServiceServer(grpcServer, handler)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("Observation Service starting on port %s...", port)

	// Start server in goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gRPC server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop accepting new connections
	grpcServer.GracefulStop()

	// Mark health as not serving
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	log.Println("Server exited")

	// Wait for context to finish
	<-shutdownCtx.Done()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
