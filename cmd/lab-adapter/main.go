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

	labadapter "github.com/Sotiris-Bekiaris/careflow-mini/internal/lab-adapter"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/db"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/observability"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	ctx := context.Background()

	// Initialize OpenTelemetry
	tracerProvider, err := observability.InitTracer("lab-adapter", getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"))
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tracerProvider(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := observability.GetTracer("lab-adapter")

	// Initialize metrics
	_, err = observability.InitMetrics("lab-adapter")
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

	// Initialize lab adapter components
	repo := labadapter.NewPostgresRepository(dbPool)
	svc := labadapter.NewService(repo, publisher, tracer)

	// Setup gRPC health check server
	grpcPort := getEnv("GRPC_PORT", "50053")

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("Lab Adapter health check starting on port %s...", grpcPort)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start worker loop
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		// Sample HL7 ORU^R01 message for testing
		sampleHL7 := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||12345^^^MRN||Doe^John||19800101|M
OBR|1|ORDER001|RESULT001|CBC^Complete Blood Count|||20240101120000
OBX|1|NM|WBC^White Blood Cell Count|1|7.5|10^3/uL|4.0-11.0|N|||F`

		for {
			select {
			case <-ticker.C:
				// Process sample HL7 message
				log.Println("Lab Adapter: Processing HL7 message...")
				_, err := svc.ProcessHL7Message(ctx, sampleHL7)
				if err != nil {
					log.Printf("Error processing HL7 message: %v", err)
				} else {
					log.Println("HL7 message processed successfully")
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Lab Adapter...")
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	grpcServer.GracefulStop()
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Lab Adapter exited")
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
