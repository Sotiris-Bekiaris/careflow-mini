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
	// TODO: Setup HL7 message listener/processor
	// TODO: Implement HL7 ORU^R01 -> FHIR Observation mapping

	log.Println("Lab Adapter starting...")

	// Worker loop placeholder
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// TODO: Poll for HL7 messages or listen to queue
				log.Println("Lab Adapter: Waiting for HL7 messages...")
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
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Lab Adapter exited")
}

func processHL7Message(hl7Msg string) error {
	// TODO: Parse HL7 ORU^R01
	// TODO: Map to FHIR Observation
	// TODO: Publish ObservationCreated event to NATS
	log.Printf("Processing HL7 message: %s", hl7Msg)
	return nil
}
