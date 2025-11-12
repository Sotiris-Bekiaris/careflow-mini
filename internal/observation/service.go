package observation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	observationv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/observation/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements observation business logic
type Service struct {
	repo      Repository
	publisher events.Publisher
	tracer    trace.Tracer
}

// NewService creates a new observation service
func NewService(repo Repository, publisher events.Publisher, tracer trace.Tracer) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		tracer:    tracer,
	}
}

// CreateObservation creates a new observation with validation
func (s *Service) CreateObservation(ctx context.Context, req *observationv1.CreateObservationRequest) (*observationv1.CreateObservationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "observation.CreateObservation")
	defer span.End()

	if req.Observation == nil {
		return nil, errors.New("observation is required")
	}

	// Validate required fields
	if req.Observation.PatientId == "" {
		return nil, errors.New("patient_id is required")
	}
	if req.Observation.Code == "" {
		return nil, errors.New("code is required")
	}

	// Convert proto to FHIR
	fhirObservation := protoToFHIR(req.Observation)
	if fhirObservation.Status == "" {
		fhirObservation.Status = "final" // Default status
	}

	// Create in database
	createdObservation, err := s.repo.Create(ctx, fhirObservation)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create observation: %w", err)
	}

	span.SetAttributes(
		attribute.String("observation.id", createdObservation.ID),
		attribute.String("observation.patient_id", req.Observation.PatientId),
	)

	// Publish event (non-blocking)
	go func() {
		event := events.Event{
			ID:        createdObservation.ID,
			Type:      events.ObservationCreated,
			Timestamp: time.Now().UTC(),
			Source:    "observation-svc",
			Data: map[string]interface{}{
				"id":         createdObservation.ID,
				"patient_id": req.Observation.PatientId,
				"code":       req.Observation.Code,
				"status":     createdObservation.Status,
			},
		}
		if err := s.publisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish observation.created event: %v\n", err)
		}
	}()

	// Convert back to proto for response
	protoObservation := fhirToProto(createdObservation)

	return &observationv1.CreateObservationResponse{
		Observation: protoObservation,
	}, nil
}

// GetObservation retrieves an observation by ID
func (s *Service) GetObservation(ctx context.Context, req *observationv1.GetObservationRequest) (*observationv1.GetObservationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "observation.GetObservation")
	defer span.End()

	if req.Id == "" {
		return nil, errors.New("observation ID is required")
	}

	span.SetAttributes(attribute.String("observation.id", req.Id))

	observation, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	protoObservation := fhirToProto(observation)

	return &observationv1.GetObservationResponse{
		Observation: protoObservation,
	}, nil
}

// ListObservations lists observations for a patient
func (s *Service) ListObservations(ctx context.Context, req *observationv1.ListObservationsRequest) (*observationv1.ListObservationsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "observation.ListObservations")
	defer span.End()

	if req.PatientId == "" {
		return nil, errors.New("patient_id is required")
	}

	span.SetAttributes(attribute.String("patient.id", req.PatientId))

	// Parse pagination
	var offset int32
	if req.PageToken != "" {
		// Simple implementation: page token is "offset:123"
		_, err := fmt.Sscanf(req.PageToken, "offset:%d", &offset)
		if err != nil {
			return nil, errors.New("invalid page token format")
		}
	}

	observations, nextToken, err := s.repo.ListByPatient(ctx, req.PatientId, req.PageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list observations: %w", err)
	}

	// Convert to proto
	protoObservations := make([]*observationv1.Observation, len(observations))
	for i, o := range observations {
		protoObservations[i] = fhirToProto(o)
	}

	return &observationv1.ListObservationsResponse{
		Observations: protoObservations,
		NextPageToken: nextToken,
	}, nil
}

// UpdateObservationStatus updates the status of an observation
func (s *Service) UpdateObservationStatus(ctx context.Context, req *observationv1.UpdateObservationStatusRequest) (*observationv1.UpdateObservationStatusResponse, error) {
	ctx, span := s.tracer.Start(ctx, "observation.UpdateObservationStatus")
	defer span.End()

	if req.Id == "" {
		return nil, errors.New("observation ID is required")
	}
	if req.Status == "" {
		return nil, errors.New("status is required")
	}

	span.SetAttributes(
		attribute.String("observation.id", req.Id),
		attribute.String("observation.status", req.Status),
	)

	// Get current observation
	observation, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get observation: %w", err)
	}

	// Update status
	observation.Status = req.Status

	updatedObservation, err := s.repo.Update(ctx, observation)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update observation: %w", err)
	}

	// Publish event (non-blocking)
	go func() {
		event := events.Event{
			ID:        updatedObservation.ID,
			Type:      events.ObservationCreated,
			Timestamp: time.Now().UTC(),
			Source:    "observation-svc",
			Data: map[string]interface{}{
				"id":     updatedObservation.ID,
				"status": req.Status,
			},
		}
		if err := s.publisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish observation status update event: %v\n", err)
		}
	}()

	protoObservation := fhirToProto(updatedObservation)

	return &observationv1.UpdateObservationStatusResponse{
		Observation: protoObservation,
	}, nil
}

// GenerateLabObservations generates a set of realistic lab observations for a patient
func (s *Service) GenerateLabObservations(ctx context.Context, patientID string) ([]*fhir.Observation, error) {
	ctx, span := s.tracer.Start(ctx, "observation.GenerateLabObservations")
	defer span.End()

	if patientID == "" {
		return nil, errors.New("patient_id is required")
	}

	span.SetAttributes(
		attribute.String("patient.id", patientID),
		attribute.Int("observation.count", len(labTestTemplates)),
	)

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Generate timestamps for observations (8 observations over 30 days)
	timestamps := generateObservationDateTimes(len(labTestTemplates))

	createdObservations := make([]*fhir.Observation, 0, len(labTestTemplates))

	// Generate one observation for each template
	for i, template := range labTestTemplates {
		value := template.generateRandomValue()
		status := template.generateStatus()

		observation := &fhir.Observation{
			Status: status,
			Code: fhir.CodeableConcept{
				Coding: []fhir.Coding{
					{
						System:  template.System,
						Code:    template.LoincCode,
						Display: template.Display,
					},
				},
				Text: template.Display,
			},
			Subject: fhir.Reference{
				Reference: "Patient/" + patientID,
			},
			EffectiveDateTime: timestamps[i],
			ValueQuantity: &fhir.Quantity{
				Value:  value,
				Unit:   template.Unit,
				System: template.UnitSystem,
				Code:   template.UnitCode,
			},
			ReferenceRange: []fhir.ReferenceRange{
				{
					Low: &fhir.Quantity{
						Value: template.MinValue,
						Unit:  template.Unit,
					},
					High: &fhir.Quantity{
						Value: template.MaxValue,
						Unit:  template.Unit,
					},
				},
			},
		}

		// Create in database
		created, err := s.repo.Create(ctx, observation)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to create observation for %s: %w", template.Display, err)
		}

		createdObservations = append(createdObservations, created)

		// Publish event (non-blocking)
		go func(obs *fhir.Observation) {
			event := events.Event{
				ID:        obs.ID,
				Type:      events.ObservationCreated,
				Timestamp: time.Now().UTC(),
				Source:    "observation-svc",
				Data: map[string]interface{}{
					"id":         obs.ID,
					"patient_id": patientID,
					"code":       template.LoincCode,
					"status":     obs.Status,
					"generated":  true,
				},
			}
			if err := s.publisher.Publish(event); err != nil {
				fmt.Printf("Failed to publish observation.created event: %v\n", err)
			}
		}(created)
	}

	span.SetAttributes(attribute.Int("observations.created", len(createdObservations)))

	return createdObservations, nil
}

// Conversion functions

// protoToFHIR converts protobuf Observation to FHIR Observation
func protoToFHIR(p *observationv1.Observation) *fhir.Observation {
	if p == nil {
		return nil
	}

	var valueQuantity *fhir.Quantity
	if p.ValueQuantityValue != "" {
		value, _ := strconv.ParseFloat(p.ValueQuantityValue, 64)
		valueQuantity = &fhir.Quantity{
			Value: value,
			Unit:  p.ValueQuantityUnit,
		}
	}

	referenceRange := []fhir.ReferenceRange{}
	if p.ReferenceRangeLow != "" || p.ReferenceRangeHigh != "" {
		var low, high *fhir.Quantity
		if p.ReferenceRangeLow != "" {
			lowVal, _ := strconv.ParseFloat(p.ReferenceRangeLow, 64)
			low = &fhir.Quantity{Value: lowVal}
		}
		if p.ReferenceRangeHigh != "" {
			highVal, _ := strconv.ParseFloat(p.ReferenceRangeHigh, 64)
			high = &fhir.Quantity{Value: highVal}
		}
		referenceRange = append(referenceRange, fhir.ReferenceRange{
			Low:  low,
			High: high,
		})
	}

	var effectiveDateTime time.Time
	if p.EffectiveDatetime != nil {
		effectiveDateTime = p.EffectiveDatetime.AsTime()
	}

	var issued time.Time
	if p.Issued != nil {
		issued = p.Issued.AsTime()
	}

	category := []fhir.CodeableConcept{}
	for _, c := range p.Category {
		category = append(category, fhir.CodeableConcept{
			Text: c,
		})
	}

	return &fhir.Observation{
		ID:                p.Id,
		Status:            p.Status,
		Category:          category,
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{
				{
					System: p.CodeSystem,
					Code:   p.Code,
				},
			},
			Text: p.Code,
		},
		Subject: fhir.Reference{
			Reference: "Patient/" + p.PatientId,
		},
		EffectiveDateTime: effectiveDateTime,
		Issued:            issued,
		ValueQuantity:     valueQuantity,
		ValueString:       p.ValueString,
		ReferenceRange:    referenceRange,
		Meta: fhir.Meta{
			LastUpdated: time.Now().UTC(),
		},
	}
}

// fhirToProto converts FHIR Observation to protobuf Observation
func fhirToProto(f *fhir.Observation) *observationv1.Observation {
	if f == nil {
		return nil
	}

	patientID := ""
	if f.Subject.Reference != "" && len(f.Subject.Reference) > 8 {
		patientID = f.Subject.Reference[8:] // Remove "Patient/" prefix
	}

	valueQuantityValue := ""
	valueQuantityUnit := ""
	if f.ValueQuantity != nil {
		valueQuantityValue = fmt.Sprintf("%v", f.ValueQuantity.Value)
		valueQuantityUnit = f.ValueQuantity.Unit
	}

	referenceLow := ""
	referenceHigh := ""
	if len(f.ReferenceRange) > 0 {
		if f.ReferenceRange[0].Low != nil {
			referenceLow = fmt.Sprintf("%v", f.ReferenceRange[0].Low.Value)
		}
		if f.ReferenceRange[0].High != nil {
			referenceHigh = fmt.Sprintf("%v", f.ReferenceRange[0].High.Value)
		}
	}

	code := ""
	codeSystem := ""
	if len(f.Code.Coding) > 0 {
		code = f.Code.Coding[0].Code
		codeSystem = f.Code.Coding[0].System
	} else {
		code = f.Code.Text
	}

	categories := make([]string, len(f.Category))
	for i, c := range f.Category {
		categories[i] = c.Text
	}

	return &observationv1.Observation{
		Id:                   f.ID,
		Status:               f.Status,
		PatientId:            patientID,
		EffectiveDatetime:    timestamppb.New(f.EffectiveDateTime),
		Issued:               timestamppb.New(f.Issued),
		Code:                 code,
		CodeSystem:           codeSystem,
		ValueQuantityValue:   valueQuantityValue,
		ValueQuantityUnit:    valueQuantityUnit,
		ValueString:          f.ValueString,
		Category:             categories,
		ReferenceRangeLow:    referenceLow,
		ReferenceRangeHigh:   referenceHigh,
		CreatedAt:            timestamppb.New(f.Meta.LastUpdated),
		UpdatedAt:            timestamppb.New(f.Meta.LastUpdated),
	}
}
