package appointment

import (
	"context"
	"errors"

	appointmentv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/appointment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the AppointmentService gRPC interface
type Handler struct {
	service *Service
	appointmentv1.UnimplementedAppointmentServiceServer
}

// NewHandler creates a new appointment handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateAppointment implements AppointmentService.CreateAppointment
func (h *Handler) CreateAppointment(ctx context.Context, req *appointmentv1.CreateAppointmentRequest) (*appointmentv1.CreateAppointmentResponse, error) {
	resp, err := h.service.CreateAppointment(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// GetAppointment implements AppointmentService.GetAppointment
func (h *Handler) GetAppointment(ctx context.Context, req *appointmentv1.GetAppointmentRequest) (*appointmentv1.GetAppointmentResponse, error) {
	resp, err := h.service.GetAppointment(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// ListAppointments implements AppointmentService.ListAppointments
func (h *Handler) ListAppointments(ctx context.Context, req *appointmentv1.ListAppointmentsRequest) (*appointmentv1.ListAppointmentsResponse, error) {
	resp, err := h.service.ListAppointments(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// UpdateAppointment implements AppointmentService.UpdateAppointment
func (h *Handler) UpdateAppointment(ctx context.Context, req *appointmentv1.UpdateAppointmentRequest) (*appointmentv1.UpdateAppointmentResponse, error) {
	resp, err := h.service.UpdateAppointment(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
}

// CancelAppointment implements AppointmentService.CancelAppointment
func (h *Handler) CancelAppointment(ctx context.Context, req *appointmentv1.CancelAppointmentRequest) (*appointmentv1.CancelAppointmentResponse, error) {
	resp, err := h.service.CancelAppointment(ctx, req)
	if err != nil {
		return nil, mapError(err)
	}
	return resp, nil
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
	if errors.Is(err, ErrAppointmentNotFound) || contains(errMsg, "not found") {
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
	ErrAppointmentNotFound = errors.New("appointment not found")
)
