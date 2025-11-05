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

	"github.com/Sotiris-Bekiaris/careflow-mini/internal/patient"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/db"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/observability"
	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	tracerProvider, err := observability.InitTracer("patient-svc", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"))
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tracerProvider(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := observability.GetTracer("patient-svc")

	// Initialize metrics
	metrics, err := observability.InitMetrics("patient-svc")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	_ = metrics // Will be used for recording metrics

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

	pool, err := db.NewPool(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

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

	// Initialize patient components
	repo := patient.NewRepository(pool)
	service := patient.NewService(repo, publisher, tracer)
	handler := patient.NewHandler(service)

	// Setup gRPC server
	port := getEnv("GRPC_PORT", "50051")

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register patient service
	patientv1.RegisterPatientServiceServer(grpcServer, handler)

	log.Printf("Patient Service starting on port %s...", port)

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
	grpcServer.GracefulStop()
	log.Println("Server exited")
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
