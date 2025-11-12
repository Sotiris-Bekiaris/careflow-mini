package observation

import (
	"context"
	"errors"

	observationv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/observation/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the ObservationService gRPC interface
type Handler struct {
	service *Service
	observationv1.UnimplementedObservationServiceServer
}

// NewHandler creates a new observation handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateObservation implements ObservationService.CreateObservation
func (h *Handler) CreateObservation(ctx context.Context, req *observationv1.CreateObservationRequest) (*observationv1.CreateObservationResponse, error) {
	resp, err := h.service.CreateObservation(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// GetObservation implements ObservationService.GetObservation
func (h *Handler) GetObservation(ctx context.Context, req *observationv1.GetObservationRequest) (*observationv1.GetObservationResponse, error) {
	resp, err := h.service.GetObservation(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// ListObservations implements ObservationService.ListObservations
func (h *Handler) ListObservations(ctx context.Context, req *observationv1.ListObservationsRequest) (*observationv1.ListObservationsResponse, error) {
	resp, err := h.service.ListObservations(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// UpdateObservationStatus implements ObservationService.UpdateObservationStatus
func (h *Handler) UpdateObservationStatus(ctx context.Context, req *observationv1.UpdateObservationStatusRequest) (*observationv1.UpdateObservationStatusResponse, error) {
	resp, err := h.service.UpdateObservationStatus(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// GenerateLabObservations implements ObservationService.GenerateLabObservations
func (h *Handler) GenerateLabObservations(ctx context.Context, req *observationv1.GenerateLabObservationsRequest) (*observationv1.GenerateLabObservationsResponse, error) {
	if req.PatientId == "" {
		return nil, status.Error(codes.InvalidArgument, "patient_id is required")
	}

	observations, err := h.service.GenerateLabObservations(ctx, req.PatientId)
	if err != nil {
		return nil, mapError(err)
	}

	// Convert FHIR observations to proto
	protoObservations := make([]*observationv1.Observation, len(observations))
	for i, obs := range observations {
		protoObservations[i] = fhirToProto(obs)
	}

	return &observationv1.GenerateLabObservationsResponse{
		Observations: protoObservations,
		Count:        int32(len(observations)),
	}, nil
}

// mapError maps domain errors to gRPC status codes
func mapError(err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for specific error conditions
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "context cancelled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "context deadline exceeded")
	}

	// Check for not found errors
	if errors.Is(err, ErrObservationNotFound) || contains(errMsg, "not found") {
		return status.Error(codes.NotFound, errMsg)
	}

	// Check for validation errors
	if contains(errMsg, "is required") || contains(errMsg, "invalid") || contains(errMsg, "must be") {
		return status.Error(codes.InvalidArgument, errMsg)
	}

	// Default to internal error
	return status.Error(codes.Internal, errMsg)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Custom errors
var (
	ErrObservationNotFound = errors.New("observation not found")
)
