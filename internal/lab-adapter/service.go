package labadapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/hl7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Service implements lab adapter business logic
type Service struct {
	repo      Repository
	publisher events.Publisher
	tracer    trace.Tracer
}

// NewService creates a new lab adapter service
func NewService(repo Repository, publisher events.Publisher, tracer trace.Tracer) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		tracer:    tracer,
	}
}

// ProcessHL7Message processes an HL7 ORU^R01 message and persists observations
func (s *Service) ProcessHL7Message(ctx context.Context, messageString string) (*fhir.Observation, error) {
	ctx, span := s.tracer.Start(ctx, "lab-adapter.ProcessHL7Message")
	defer span.End()

	if messageString == "" {
		return nil, errors.New("HL7 message is empty")
	}

	// Parse HL7 message
	msg, err := hl7.Parse(messageString)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse HL7 message: %w", err)
	}

	span.SetAttributes(attribute.String("hl7.message_type", msg.Type))

	// Validate message type
	if msg.Type != "ORU^R01" {
		return nil, fmt.Errorf("unsupported message type: %s (only ORU^R01 supported)", msg.Type)
	}

	// Map to FHIR Observation
	observation, err := hl7.MapToFHIRObservation(msg)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to map HL7 to FHIR: %w", err)
	}

	span.SetAttributes(
		attribute.String("observation.id", observation.ID),
		attribute.String("observation.code", observation.Code.Text),
	)

	// Persist to database
	createdObs, err := s.repo.CreateObservation(ctx, observation)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to persist observation: %w", err)
	}

	// Extract patient ID for event publishing
	patientID := ""
	if observation.Subject.Reference != "" && len(observation.Subject.Reference) > 8 {
		patientID = observation.Subject.Reference[8:]
	}

	// Publish event (non-blocking)
	go func() {
		event := events.Event{
			ID:        createdObs.ID,
			Type:      events.ObservationCreated,
			Timestamp: time.Now().UTC(),
			Source:    "lab-adapter",
			Data: map[string]interface{}{
				"id":               createdObs.ID,
				"patient_id":       patientID,
				"observation_code": createdObs.Code.Text,
				"value":            createdObs.ValueQuantity,
				"status":           createdObs.Status,
			},
		}
		if err := s.publisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish observation.created event: %v\n", err)
		}
	}()

	return createdObs, nil
}

// GetObservation retrieves an observation by ID
func (s *Service) GetObservation(ctx context.Context, id string) (*fhir.Observation, error) {
	ctx, span := s.tracer.Start(ctx, "lab-adapter.GetObservation")
	defer span.End()

	if id == "" {
		return nil, errors.New("observation ID is required")
	}

	span.SetAttributes(attribute.String("observation.id", id))

	observation, err := s.repo.GetObservation(ctx, id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	return observation, nil
}

// ListObservationsByPatient lists observations for a patient
func (s *Service) ListObservationsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Observation, string, error) {
	ctx, span := s.tracer.Start(ctx, "lab-adapter.ListObservationsByPatient")
	defer span.End()

	if patientID == "" {
		return nil, "", errors.New("patient ID is required")
	}

	span.SetAttributes(attribute.String("patient.id", patientID))

	observations, nextToken, err := s.repo.ListObservationsByPatient(ctx, patientID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, "", fmt.Errorf("failed to list observations: %w", err)
	}

	return observations, nextToken, nil
}
