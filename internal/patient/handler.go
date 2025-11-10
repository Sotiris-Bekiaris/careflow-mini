package patient

import (
	"context"

	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the gRPC PatientService server.
type Handler struct {
	patientv1.UnimplementedPatientServiceServer
	service Service
}

// NewHandler creates a new patient gRPC handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreatePatient implements the CreatePatient RPC method.
func (h *Handler) CreatePatient(ctx context.Context, req *patientv1.CreatePatientRequest) (*patientv1.CreatePatientResponse, error) {
	resp, err := h.service.CreatePatient(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// GetPatient implements the GetPatient RPC method.
func (h *Handler) GetPatient(ctx context.Context, req *patientv1.GetPatientRequest) (*patientv1.GetPatientResponse, error) {
	resp, err := h.service.GetPatient(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// UpdatePatient implements the UpdatePatient RPC method.
func (h *Handler) UpdatePatient(ctx context.Context, req *patientv1.UpdatePatientRequest) (*patientv1.UpdatePatientResponse, error) {
	resp, err := h.service.UpdatePatient(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// DeletePatient implements the DeletePatient RPC method.
func (h *Handler) DeletePatient(ctx context.Context, req *patientv1.DeletePatientRequest) (*patientv1.DeletePatientResponse, error) {
	resp, err := h.service.DeletePatient(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// ListPatients implements the ListPatients RPC method.
func (h *Handler) ListPatients(ctx context.Context, req *patientv1.ListPatientsRequest) (*patientv1.ListPatientsResponse, error) {
	resp, err := h.service.ListPatients(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// mapError maps internal errors to gRPC status codes.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	// Check for common error patterns
	errMsg := err.Error()

	// Not found errors
	if contains(errMsg, "not found") {
		return status.Error(codes.NotFound, errMsg)
	}

	// Validation errors
	if contains(errMsg, "validation failed") || contains(errMsg, "required") {
		return status.Error(codes.InvalidArgument, errMsg)
	}

	// Already exists errors
	if contains(errMsg, "already exists") || contains(errMsg, "duplicate") {
		return status.Error(codes.AlreadyExists, errMsg)
	}

	// Default to internal error
	return status.Error(codes.Internal, "internal server error")
}

// contains checks if a string contains a substring (case-insensitive helper).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring is a simple substring search helper.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
