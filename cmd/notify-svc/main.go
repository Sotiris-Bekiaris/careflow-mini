package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// TODO: Initialize OpenTelemetry
	// TODO: Connect to NATS
	// TODO: Subscribe to ObservationCreated events
	// TODO: Process notifications (email, SMS, etc.)

	log.Println("Notify Service starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Event listener placeholder
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// TODO: Listen for events from NATS
				time.Sleep(5 * time.Second)
				log.Println("Notify Service: Waiting for events...")
			}
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Notify Service...")
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Notify Service exited")
}

func handleObservationCreated(event map[string]interface{}) {
	// TODO: Parse event
	// TODO: Send notification
	log.Printf("Handling ObservationCreated event: %v", event)
}
