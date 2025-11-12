package appointment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	appointmentv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/appointment/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements appointment business logic
type Service struct {
	repo      Repository
	publisher events.Publisher
	tracer    trace.Tracer
}

// NewService creates a new appointment service
func NewService(repo Repository, publisher events.Publisher, tracer trace.Tracer) *Service {
	return &Service{
		repo:      repo,
		publisher: publisher,
		tracer:    tracer,
	}
}

// CreateAppointment creates a new appointment with validation
func (s *Service) CreateAppointment(ctx context.Context, req *appointmentv1.CreateAppointmentRequest) (*appointmentv1.CreateAppointmentResponse, error) {
	ctx, span := s.tracer.Start(ctx, "appointment.CreateAppointment")
	defer span.End()

	if req.Appointment == nil {
		return nil, errors.New("appointment is required")
	}

	// Validate required fields
	if req.Appointment.PatientId == "" {
		return nil, errors.New("patient_id is required")
	}
	if req.Appointment.Start == nil {
		return nil, errors.New("start time is required")
	}
	if req.Appointment.End == nil {
		return nil, errors.New("end time is required")
	}

	// Validate start time is before end time
	startTime := req.Appointment.Start.AsTime()
	endTime := req.Appointment.End.AsTime()
	if !startTime.Before(endTime) {
		return nil, errors.New("start time must be before end time")
	}

	// Validate appointment is not in the past
	if startTime.Before(time.Now().UTC()) {
		return nil, errors.New("appointment start time cannot be in the past")
	}

	// Convert proto to FHIR
	fhirAppointment := protoToFHIR(req.Appointment)

	// Create in database
	createdAppointment, err := s.repo.CreateAppointment(ctx, fhirAppointment)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	span.SetAttributes(
		attribute.String("appointment.id", createdAppointment.ID),
		attribute.String("appointment.patient_id", req.Appointment.PatientId),
	)

	// Publish event (non-blocking)
	go func() {
		event := events.Event{
			ID:        createdAppointment.ID,
			Type:      events.AppointmentCreated,
			Timestamp: time.Now().UTC(),
			Source:    "appointment-svc",
			Data: map[string]interface{}{
				"id":              createdAppointment.ID,
				"patient_id":      req.Appointment.PatientId,
				"practitioner_id": req.Appointment.PractitionerId,
				"start":           createdAppointment.Start,
				"end":             createdAppointment.End,
			},
		}
		if err := s.publisher.Publish(event); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to publish appointment.created event: %v\n", err)
		}
	}()

	// Convert back to proto for response
	protoAppointment := fhirToProto(createdAppointment)

	return &appointmentv1.CreateAppointmentResponse{
		Appointment: protoAppointment,
	}, nil
}

// GetAppointment retrieves an appointment by ID
func (s *Service) GetAppointment(ctx context.Context, req *appointmentv1.GetAppointmentRequest) (*appointmentv1.GetAppointmentResponse, error) {
	ctx, span := s.tracer.Start(ctx, "appointment.GetAppointment")
	defer span.End()

	if req.Id == "" {
		return nil, errors.New("appointment ID is required")
	}

	span.SetAttributes(attribute.String("appointment.id", req.Id))

	appointment, err := s.repo.GetAppointment(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	protoAppointment := fhirToProto(appointment)

	return &appointmentv1.GetAppointmentResponse{
		Appointment: protoAppointment,
	}, nil
}

// ListAppointments lists appointments for a patient
func (s *Service) ListAppointments(ctx context.Context, req *appointmentv1.ListAppointmentsRequest) (*appointmentv1.ListAppointmentsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "appointment.ListAppointments")
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

	appointments, nextToken, err := s.repo.ListAppointmentsByPatient(ctx, req.PatientId, req.PageSize, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list appointments: %w", err)
	}

	// Convert to proto
	protoAppointments := make([]*appointmentv1.Appointment, len(appointments))
	for i, a := range appointments {
		protoAppointments[i] = fhirToProto(a)
	}

	return &appointmentv1.ListAppointmentsResponse{
		Appointments:  protoAppointments,
		NextPageToken: nextToken,
	}, nil
}

// UpdateAppointment updates an existing appointment
func (s *Service) UpdateAppointment(ctx context.Context, req *appointmentv1.UpdateAppointmentRequest) (*appointmentv1.UpdateAppointmentResponse, error) {
	ctx, span := s.tracer.Start(ctx, "appointment.UpdateAppointment")
	defer span.End()

	if req.Appointment == nil {
		return nil, errors.New("appointment is required")
	}

	if req.Appointment.Id == "" {
		return nil, errors.New("appointment ID is required")
	}

	// Validate required fields
	if req.Appointment.PatientId == "" {
		return nil, errors.New("patient_id is required")
	}
	if req.Appointment.Start == nil {
		return nil, errors.New("start time is required")
	}
	if req.Appointment.End == nil {
		return nil, errors.New("end time is required")
	}

	// Validate start time is before end time
	startTime := req.Appointment.Start.AsTime()
	endTime := req.Appointment.End.AsTime()
	if !startTime.Before(endTime) {
		return nil, errors.New("start time must be before end time")
	}

	span.SetAttributes(
		attribute.String("appointment.id", req.Appointment.Id),
		attribute.String("appointment.patient_id", req.Appointment.PatientId),
	)

	// Convert proto to FHIR
	fhirAppointment := protoToFHIR(req.Appointment)

	// Update in database
	updatedAppointment, err := s.repo.UpdateAppointment(ctx, fhirAppointment)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	// Publish event (non-blocking)
	go func() {
		event := events.Event{
			ID:        updatedAppointment.ID,
			Type:      events.AppointmentUpdated,
			Timestamp: time.Now().UTC(),
			Source:    "appointment-svc",
			Data: map[string]interface{}{
				"id":              updatedAppointment.ID,
				"patient_id":      req.Appointment.PatientId,
				"practitioner_id": req.Appointment.PractitionerId,
				"start":           updatedAppointment.Start,
				"end":             updatedAppointment.End,
				"status":          updatedAppointment.Status,
			},
		}
		if err := s.publisher.Publish(event); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to publish appointment.updated event: %v\n", err)
		}
	}()

	// Convert back to proto for response
	protoAppointment := fhirToProto(updatedAppointment)

	return &appointmentv1.UpdateAppointmentResponse{
		Appointment: protoAppointment,
	}, nil
}

// CancelAppointment cancels an appointment
func (s *Service) CancelAppointment(ctx context.Context, req *appointmentv1.CancelAppointmentRequest) (*appointmentv1.CancelAppointmentResponse, error) {
	ctx, span := s.tracer.Start(ctx, "appointment.CancelAppointment")
	defer span.End()

	if req.Id == "" {
		return nil, errors.New("appointment ID is required")
	}

	span.SetAttributes(attribute.String("appointment.id", req.Id))

	// Get current appointment
	appointment, err := s.repo.GetAppointment(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get appointment: %w", err)
	}

	// Only allow cancellation if appointment hasn't started
	if appointment.Start.Before(time.Now().UTC()) {
		return nil, errors.New("cannot cancel an appointment that has already started")
	}

	// Update status to cancelled
	appointment.Status = "cancelled"

	updatedAppointment, err := s.repo.UpdateAppointment(ctx, appointment)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to cancel appointment: %w", err)
	}

	// Publish event (non-blocking)
	go func() {
		event := events.Event{
			ID:        updatedAppointment.ID,
			Type:      events.AppointmentCancelled,
			Timestamp: time.Now().UTC(),
			Source:    "appointment-svc",
			Data: map[string]interface{}{
				"id":     updatedAppointment.ID,
				"reason": req.Reason,
			},
		}
		if err := s.publisher.Publish(event); err != nil {
			fmt.Printf("Failed to publish appointment.cancelled event: %v\n", err)
		}
	}()

	protoAppointment := fhirToProto(updatedAppointment)

	return &appointmentv1.CancelAppointmentResponse{
		Appointment: protoAppointment,
	}, nil
}

// Conversion functions

// protoToFHIR converts protobuf Appointment to FHIR Appointment
func protoToFHIR(p *appointmentv1.Appointment) *fhir.Appointment {
	if p == nil {
		return nil
	}

	startTime := p.Start.AsTime()
	endTime := p.End.AsTime()

	participants := []fhir.Participant{
		{
			Actor: fhir.Reference{
				Reference: "Patient/" + p.PatientId,
			},
			Required: "required",
			Status:   "accepted",
		},
	}

	if p.PractitionerId != "" {
		participants = append(participants, fhir.Participant{
			Actor: fhir.Reference{
				Reference: "Practitioner/" + p.PractitionerId,
			},
			Required: "required",
			Status:   "accepted",
		})
	}

	serviceType := []fhir.CodeableConcept{}
	if p.ServiceType != "" {
		serviceType = append(serviceType, fhir.CodeableConcept{
			Text: p.ServiceType,
		})
	}

	return &fhir.Appointment{
		ID:          p.Id,
		Status:      p.Status,
		Start:       startTime,
		End:         endTime,
		ServiceType: serviceType,
		Participant: participants,
		Description: p.Description,
		Meta: fhir.Meta{
			LastUpdated: time.Now().UTC(),
		},
	}
}

// fhirToProto converts FHIR Appointment to protobuf Appointment
func fhirToProto(f *fhir.Appointment) *appointmentv1.Appointment {
	if f == nil {
		return nil
	}

	var patientID, practitionerID string
	for _, p := range f.Participant {
		if len(p.Actor.Reference) > 8 && p.Actor.Reference[:8] == "Patient/" {
			patientID = p.Actor.Reference[8:]
		} else if len(p.Actor.Reference) > 13 && p.Actor.Reference[:13] == "Practitioner/" {
			practitionerID = p.Actor.Reference[13:]
		}
	}

	serviceType := ""
	if len(f.ServiceType) > 0 {
		serviceType = f.ServiceType[0].Text
	}

	return &appointmentv1.Appointment{
		Id:             f.ID,
		Status:         f.Status,
		PatientId:      patientID,
		PractitionerId: practitionerID,
		Start:          timestampProto(f.Start),
		End:            timestampProto(f.End),
		ServiceType:    serviceType,
		Description:    f.Description,
	}
}

// timestampProto converts time.Time to protobuf Timestamp
func timestampProto(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
