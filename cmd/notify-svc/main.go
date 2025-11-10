package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// TODO: Initialize OpenTelemetry
	// TODO: Connect to NATS
	// TODO: Subscribe to ObservationCreated events
	// TODO: Process notifications (email, SMS, etc.)

	log.Println("Notify Service starting...")

	// Setup gRPC health check server
	grpcPort := "50054"

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("Notify Service health check starting on port %s...", grpcPort)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Event listener placeholder
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		// Sample event data for testing
		sampleEvent := map[string]interface{}{
			"id":        "event-123",
			"type":      events.ObservationCreated,
			"timestamp": time.Now(),
			"source":    "lab-adapter",
			"data": map[string]interface{}{
				"observation": map[string]interface{}{
					"patient_id": "12345",
					"test_type":  "CBC",
					"status":     "final",
				},
			},
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Process sample observation event
				log.Println("Notify Service: Received ObservationCreated event...")
				handleObservationCreated(sampleEvent)
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Notify Service...")
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	grpcServer.GracefulStop()
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Notify Service exited")
}

func handleObservationCreated(event map[string]interface{}) {
	// Parse event data
	eventType, ok := event["type"].(events.EventType)
	if !ok {
		log.Printf("Invalid event type")
		return
	}

	log.Printf("Processing event type: %s", eventType)

	// Extract observation data
	data, ok := event["data"].(map[string]interface{})
	if !ok {
		log.Printf("Invalid event data")
		return
	}

	observation, ok := data["observation"].(map[string]interface{})
	if !ok {
		log.Printf("No observation found in event data")
		return
	}

	// Extract patient information
	patientID, _ := observation["patient_id"].(string)
	testType, _ := observation["test_type"].(string)
	status, _ := observation["status"].(string)

	// TODO: Send actual notifications (email, SMS, push notification)
	log.Printf("Notification: New %s test result for patient %s (Status: %s)", testType, patientID, status)
	log.Printf("Would send notification to patient %s about their %s results", patientID, testType)
}
