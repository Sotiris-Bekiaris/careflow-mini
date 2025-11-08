package patient

import (
	"context"
	"fmt"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"go.opentelemetry.io/otel/trace"
)

// Service defines the business logic interface for patient operations.
type Service interface {
	CreatePatient(ctx context.Context, req *patientv1.CreatePatientRequest) (*patientv1.CreatePatientResponse, error)
	GetPatient(ctx context.Context, req *patientv1.GetPatientRequest) (*patientv1.GetPatientResponse, error)
	UpdatePatient(ctx context.Context, req *patientv1.UpdatePatientRequest) (*patientv1.UpdatePatientResponse, error)
	ListPatients(ctx context.Context, req *patientv1.ListPatientsRequest) (*patientv1.ListPatientsResponse, error)
}

// service implements Service interface.
type service struct {
	repo      Repository
	publisher events.Publisher
	tracer    trace.Tracer
}

// NewService creates a new patient service.
func NewService(repo Repository, publisher events.Publisher, tracer trace.Tracer) Service {
	return &service{
		repo:      repo,
		publisher: publisher,
		tracer:    tracer,
	}
}

// CreatePatient creates a new patient record.
func (s *service) CreatePatient(ctx context.Context, req *patientv1.CreatePatientRequest) (*patientv1.CreatePatientResponse, error) {
	ctx, span := s.tracer.Start(ctx, "patient.CreatePatient")
	defer span.End()

	// Validate request
	if req.Patient == nil {
		return nil, fmt.Errorf("patient is required")
	}
	if err := validatePatient(req.Patient); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert proto to FHIR
	fhirPatient := protoToFHIR(req.Patient)
	fhirPatient.Active = true // New patients are active by default

	// Create in repository
	if err := s.repo.Create(ctx, fhirPatient); err != nil {
		return nil, fmt.Errorf("failed to create patient: %w", err)
	}

	// Publish patient.created event
	event := events.Event{
		Type:      events.PatientCreated,
		Timestamp: time.Now(),
		Source:    "patient-svc",
		Data: map[string]interface{}{
			"patient_id": fhirPatient.ID,
			"name":       fhirPatient.Name.Family,
		},
	}
	if err := s.publisher.Publish(event); err != nil {
		// Log error but don't fail the request
		// In production, consider using a dead-letter queue
		span.RecordError(err)
	}

	// Convert back to proto and return
	return &patientv1.CreatePatientResponse{
		Patient: fhirToProto(fhirPatient),
	}, nil
}

// GetPatient retrieves a patient by ID.
func (s *service) GetPatient(ctx context.Context, req *patientv1.GetPatientRequest) (*patientv1.GetPatientResponse, error) {
	ctx, span := s.tracer.Start(ctx, "patient.GetPatient")
	defer span.End()

	// Validate request
	if req.Id == "" {
		return nil, fmt.Errorf("patient ID is required")
	}

	// Get from repository
	fhirPatient, err := s.repo.GetByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}

	// Convert to proto and return
	return &patientv1.GetPatientResponse{
		Patient: fhirToProto(fhirPatient),
	}, nil
}

// UpdatePatient updates an existing patient record.
func (s *service) UpdatePatient(ctx context.Context, req *patientv1.UpdatePatientRequest) (*patientv1.UpdatePatientResponse, error) {
	ctx, span := s.tracer.Start(ctx, "patient.UpdatePatient")
	defer span.End()

	// Validate request
	if req.Patient == nil {
		return nil, fmt.Errorf("patient is required")
	}
	if req.Patient.Id == "" {
		return nil, fmt.Errorf("patient ID is required")
	}
	if err := validatePatient(req.Patient); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert proto to FHIR
	fhirPatient := protoToFHIR(req.Patient)

	// Update in repository
	if err := s.repo.Update(ctx, fhirPatient); err != nil {
		return nil, fmt.Errorf("failed to update patient: %w", err)
	}

	// Publish patient.updated event
	event := events.Event{
		Type:      events.PatientUpdated,
		Timestamp: time.Now(),
		Source:    "patient-svc",
		Data: map[string]interface{}{
			"patient_id": fhirPatient.ID,
			"name":       fhirPatient.Name.Family,
		},
	}
	if err := s.publisher.Publish(event); err != nil {
		// Log error but don't fail the request
		span.RecordError(err)
	}

	// Convert back to proto and return
	return &patientv1.UpdatePatientResponse{
		Patient: fhirToProto(fhirPatient),
	}, nil
}

// ListPatients lists patients with pagination.
func (s *service) ListPatients(ctx context.Context, req *patientv1.ListPatientsRequest) (*patientv1.ListPatientsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "patient.ListPatients")
	defer span.End()

	// Parse pagination parameters
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10 // Default page size
	}

	// For simplicity, we'll use offset-based pagination
	// In production, consider cursor-based pagination
	offset := 0
	if req.PageToken != "" {
		// Parse page token as offset
		_, _ = fmt.Sscanf(req.PageToken, "%d", &offset)
	}

	// Get from repository
	fhirPatients, err := s.repo.List(ctx, pageSize+1, offset) // Get one extra to check for next page
	if err != nil {
		return nil, fmt.Errorf("failed to list patients: %w", err)
	}

	// Check if there's a next page
	var nextPageToken string
	if len(fhirPatients) > pageSize {
		fhirPatients = fhirPatients[:pageSize]
		nextPageToken = fmt.Sprintf("%d", offset+pageSize)
	}

	// Convert to proto
	protoPatients := make([]*patientv1.Patient, len(fhirPatients))
	for i, fhirPatient := range fhirPatients {
		protoPatients[i] = fhirToProto(fhirPatient)
	}

	return &patientv1.ListPatientsResponse{
		Patients:      protoPatients,
		NextPageToken: nextPageToken,
	}, nil
}

// validatePatient validates a patient proto message.
func validatePatient(p *patientv1.Patient) error {
	if p.FamilyName == "" {
		return fmt.Errorf("family name is required")
	}
	if len(p.GivenNames) == 0 {
		return fmt.Errorf("at least one given name is required")
	}
	if p.BirthDate != "" {
		// Validate date format (YYYY-MM-DD)
		_, err := time.Parse("2006-01-02", p.BirthDate)
		if err != nil {
			return fmt.Errorf("invalid birth date format (expected YYYY-MM-DD): %w", err)
		}
	}
	return nil
}

// protoToFHIR converts a proto Patient to a FHIR Patient.
func protoToFHIR(p *patientv1.Patient) *fhir.Patient {
	patient := &fhir.Patient{
		ID:     p.Id,
		Active: p.Active,
		Name: fhir.HumanName{
			Family: p.FamilyName,
			Given:  p.GivenNames,
		},
		Gender:    p.Gender,
		BirthDate: p.BirthDate,
	}

	// Convert telecom
	if len(p.Telecom) > 0 {
		patient.Telecom = make([]fhir.Contact, len(p.Telecom))
		for i, t := range p.Telecom {
			patient.Telecom[i] = fhir.Contact{
				System: t.System,
				Value:  t.Value,
				Use:    t.Use,
			}
		}
	}

	// Convert address
	if p.Address != nil {
		patient.Address = []fhir.Address{{
			Line:       p.Address.Line,
			City:       p.Address.City,
			State:      p.Address.State,
			PostalCode: p.Address.PostalCode,
			Country:    p.Address.Country,
		}}
	}

	return patient
}

// fhirToProto converts a FHIR Patient to a proto Patient.
func fhirToProto(p *fhir.Patient) *patientv1.Patient {
	patient := &patientv1.Patient{
		Id:         p.ID,
		FamilyName: p.Name.Family,
		GivenNames: p.Name.Given,
		Gender:     p.Gender,
		BirthDate:  p.BirthDate,
		Active:     p.Active,
	}

	// Convert telecom
	if len(p.Telecom) > 0 {
		patient.Telecom = make([]*patientv1.Contact, len(p.Telecom))
		for i, t := range p.Telecom {
			patient.Telecom[i] = &patientv1.Contact{
				System: t.System,
				Value:  t.Value,
				Use:    t.Use,
			}
		}
	}

	// Convert address (take first address if multiple)
	if len(p.Address) > 0 {
		patient.Address = &patientv1.Address{
			Line:       p.Address[0].Line,
			City:       p.Address[0].City,
			State:      p.Address[0].State,
			PostalCode: p.Address[0].PostalCode,
			Country:    p.Address[0].Country,
		}
	}

	return patient
}
