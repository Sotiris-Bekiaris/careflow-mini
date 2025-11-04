package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/hl7"
	"github.com/google/uuid"
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
				if err := processHL7Message(sampleHL7); err != nil {
					log.Printf("Error processing HL7 message: %v", err)
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
	cancel()
	time.Sleep(1 * time.Second)
	log.Println("Lab Adapter exited")
}

func processHL7Message(hl7Msg string) error {
	// Parse HL7 ORU^R01
	msg, err := hl7.Parse(hl7Msg)
	if err != nil {
		return fmt.Errorf("failed to parse HL7 message: %w", err)
	}

	log.Printf("Parsed HL7 message type: %s", msg.Type)

	// Map to FHIR Observation
	obs, err := hl7.MapToFHIRObservation(msg)
	if err != nil {
		return fmt.Errorf("failed to map HL7 to FHIR: %w", err)
	}

	log.Printf("Mapped to FHIR Observation for patient: %s", obs.Subject.Reference)

	// Create ObservationCreated event
	event := events.Event{
		ID:        uuid.New().String(),
		Type:      events.ObservationCreated,
		Timestamp: time.Now(),
		Source:    "lab-adapter",
		Data: map[string]interface{}{
			"observation": obs,
		},
	}

	// TODO: Publish to NATS when publisher is available
	log.Printf("Would publish ObservationCreated event: ID=%s, Patient=%s", event.ID, obs.Subject.Reference)

	return nil
}
