package appointment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	appointmentv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/appointment/v1"
	"go.opentelemetry.io/otel"
)

// Mock Publisher for testing
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(event events.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateAppointment(ctx context.Context, appointment *fhir.Appointment) (*fhir.Appointment, error) {
	args := m.Called(ctx, appointment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fhir.Appointment), args.Error(1)
}

func (m *MockRepository) GetAppointment(ctx context.Context, id string) (*fhir.Appointment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fhir.Appointment), args.Error(1)
}

func (m *MockRepository) UpdateAppointment(ctx context.Context, appointment *fhir.Appointment) (*fhir.Appointment, error) {
	args := m.Called(ctx, appointment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fhir.Appointment), args.Error(1)
}

func (m *MockRepository) ListAppointmentsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Appointment, string, error) {
	args := m.Called(ctx, patientID, limit, offset)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).([]*fhir.Appointment), args.String(1), args.Error(2)
}

func (m *MockRepository) DeleteAppointment(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Test helper function to create test appointment
func createTestAppointment(id, patientID string, startOffset time.Duration) *fhir.Appointment {
	now := time.Now().UTC()
	return &fhir.Appointment{
		ID:     id,
		Status: "booked",
		Participant: []fhir.Participant{
			{
				Actor: fhir.Reference{
					Reference: "Patient/" + patientID,
				},
				Required: "required",
				Status:   "accepted",
			},
		},
		Start:       now.Add(startOffset),
		End:         now.Add(startOffset + time.Hour),
		Description: "Regular checkup",
		Meta: fhir.Meta{
			LastUpdated: now,
		},
	}
}

// ============================================================
// SERVICE LAYER TESTS
// ============================================================

func TestService_CreateAppointment(t *testing.T) {
	tests := []struct {
		name    string
		req     *appointmentv1.CreateAppointmentRequest
		mockFn  func(*MockRepository, *MockPublisher)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful create",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:          "apt-123",
					PatientId:   "pat-123",
					Status:      "booked",
					Start:       timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
					End:         timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
					Description: "Checkup",
					ServiceType: "general",
				},
			},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				mr.On("CreateAppointment", mock.Anything, mock.Anything).Return(
					createTestAppointment("apt-123", "pat-123", 24*time.Hour), nil,
				)
				mp.On("Publish", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "missing appointment",
			req:     &appointmentv1.CreateAppointmentRequest{},
			wantErr: true,
			errMsg:  "appointment is required",
		},
		{
			name: "missing patient ID",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:    "apt-123",
					Start: timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
					End:   timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
				},
			},
			wantErr: true,
			errMsg:  "patient_id is required",
		},
		{
			name: "missing start time",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:        "apt-123",
					PatientId: "pat-123",
					End:       timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
				},
			},
			wantErr: true,
			errMsg:  "start time is required",
		},
		{
			name: "missing end time",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:        "apt-123",
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
				},
			},
			wantErr: true,
			errMsg:  "end time is required",
		},
		{
			name: "start time after end time",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:        "apt-123",
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
					End:       timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
				},
			},
			wantErr: true,
			errMsg:  "start time must be before end time",
		},
		{
			name: "appointment in the past",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:        "apt-123",
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(-1 * time.Hour)),
					End:       timestamppb.New(time.Now().UTC()),
				},
			},
			wantErr: true,
			errMsg:  "cannot be in the past",
		},
		{
			name: "database error",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:        "apt-123",
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
					End:       timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
				},
			},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				mr.On("CreateAppointment", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "failed to create appointment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockPub := new(MockPublisher)

			if tt.mockFn != nil {
				tt.mockFn(mockRepo, mockPub)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, mockPub, tracer)

			resp, err := svc.CreateAppointment(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "apt-123", resp.Appointment.Id)
			}
		})
	}
}

func TestService_GetAppointment(t *testing.T) {
	tests := []struct {
		name    string
		req     *appointmentv1.GetAppointmentRequest
		mockFn  func(*MockRepository)
		wantErr bool
	}{
		{
			name: "successful get",
			req:  &appointmentv1.GetAppointmentRequest{Id: "apt-123"},
			mockFn: func(mr *MockRepository) {
				mr.On("GetAppointment", mock.Anything, "apt-123").Return(
					createTestAppointment("apt-123", "pat-123", 24*time.Hour), nil,
				)
			},
			wantErr: false,
		},
		{
			name:    "missing ID",
			req:     &appointmentv1.GetAppointmentRequest{},
			wantErr: true,
		},
		{
			name: "not found",
			req:  &appointmentv1.GetAppointmentRequest{Id: "apt-999"},
			mockFn: func(mr *MockRepository) {
				mr.On("GetAppointment", mock.Anything, "apt-999").Return(
					nil, errors.New("appointment not found"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, new(MockPublisher), tracer)

			resp, err := svc.GetAppointment(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestService_ListAppointments(t *testing.T) {
	tests := []struct {
		name    string
		req     *appointmentv1.ListAppointmentsRequest
		mockFn  func(*MockRepository)
		wantErr bool
	}{
		{
			name: "successful list",
			req: &appointmentv1.ListAppointmentsRequest{
				PatientId: "pat-123",
				PageSize:  10,
			},
			mockFn: func(mr *MockRepository) {
				mr.On("ListAppointmentsByPatient", mock.Anything, "pat-123", int32(10), int32(0)).Return(
					[]*fhir.Appointment{
						createTestAppointment("apt-1", "pat-123", 24*time.Hour),
						createTestAppointment("apt-2", "pat-123", 48*time.Hour),
					},
					"offset:10",
					nil,
				)
			},
			wantErr: false,
		},
		{
			name:    "missing patient ID",
			req:     &appointmentv1.ListAppointmentsRequest{},
			wantErr: true,
		},
		{
			name: "database error",
			req: &appointmentv1.ListAppointmentsRequest{
				PatientId: "pat-123",
			},
			mockFn: func(mr *MockRepository) {
				mr.On("ListAppointmentsByPatient", mock.Anything, "pat-123", mock.Anything, mock.Anything).Return(
					nil, "", errors.New("db error"),
				)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, new(MockPublisher), tracer)

			resp, err := svc.ListAppointments(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestService_CancelAppointment(t *testing.T) {
	tests := []struct {
		name    string
		req     *appointmentv1.CancelAppointmentRequest
		mockFn  func(*MockRepository, *MockPublisher)
		wantErr bool
		errMsg  string
	}{
		{
			name: "successful cancel",
			req:  &appointmentv1.CancelAppointmentRequest{Id: "apt-123", Reason: "Patient request"},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				apt := createTestAppointment("apt-123", "pat-123", 24*time.Hour)
				mr.On("GetAppointment", mock.Anything, "apt-123").Return(apt, nil)

				cancelledApt := apt
				cancelledApt.Status = "cancelled"
				mr.On("UpdateAppointment", mock.Anything, mock.Anything).Return(cancelledApt, nil)

				mp.On("Publish", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:    "missing ID",
			req:     &appointmentv1.CancelAppointmentRequest{},
			wantErr: true,
			errMsg:  "appointment ID is required",
		},
		{
			name: "appointment not found",
			req:  &appointmentv1.CancelAppointmentRequest{Id: "apt-999"},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				mr.On("GetAppointment", mock.Anything, "apt-999").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "appointment already started",
			req:  &appointmentv1.CancelAppointmentRequest{Id: "apt-123"},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				pastApt := createTestAppointment("apt-123", "pat-123", -1*time.Hour)
				mr.On("GetAppointment", mock.Anything, "apt-123").Return(pastApt, nil)
			},
			wantErr: true,
			errMsg:  "cannot cancel an appointment that has already started",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockPub := new(MockPublisher)

			if tt.mockFn != nil {
				tt.mockFn(mockRepo, mockPub)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, mockPub, tracer)

			resp, err := svc.CancelAppointment(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "cancelled", resp.Appointment.Status)
			}
		})
	}
}

// ============================================================
// HANDLER LAYER TESTS
// ============================================================

func TestHandler_CreateAppointment(t *testing.T) {
	tests := []struct {
		name     string
		req      *appointmentv1.CreateAppointmentRequest
		mockFn   func(*MockRepository, *MockPublisher)
		wantErr  bool
		wantCode codes.Code
		wantMsg  string
	}{
		{
			name: "successful create",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					Id:        "apt-123",
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
					End:       timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
				},
			},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				mr.On("CreateAppointment", mock.Anything, mock.Anything).Return(
					createTestAppointment("apt-123", "pat-123", 24*time.Hour), nil,
				)
				mp.On("Publish", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "validation error",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
					End:       timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
				},
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "database error",
			req: &appointmentv1.CreateAppointmentRequest{
				Appointment: &appointmentv1.Appointment{
					PatientId: "pat-123",
					Start:     timestamppb.New(time.Now().UTC().Add(24 * time.Hour)),
					End:       timestamppb.New(time.Now().UTC().Add(25 * time.Hour)),
				},
			},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				mr.On("CreateAppointment", mock.Anything, mock.Anything).Return(
					nil, errors.New("db error"),
				)
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockPub := new(MockPublisher)

			if tt.mockFn != nil {
				tt.mockFn(mockRepo, mockPub)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, mockPub, tracer)
			handler := NewHandler(svc)

			resp, err := handler.CreateAppointment(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok, "error should be a gRPC status error")
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHandler_GetAppointment(t *testing.T) {
	tests := []struct {
		name     string
		req      *appointmentv1.GetAppointmentRequest
		mockFn   func(*MockRepository)
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "successful get",
			req:  &appointmentv1.GetAppointmentRequest{Id: "apt-123"},
			mockFn: func(mr *MockRepository) {
				mr.On("GetAppointment", mock.Anything, "apt-123").Return(
					createTestAppointment("apt-123", "pat-123", 24*time.Hour), nil,
				)
			},
			wantErr: false,
		},
		{
			name: "not found",
			req:  &appointmentv1.GetAppointmentRequest{Id: "apt-999"},
			mockFn: func(mr *MockRepository) {
				mr.On("GetAppointment", mock.Anything, "apt-999").Return(
					nil, errors.New("not found"),
				)
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, new(MockPublisher), tracer)
			handler := NewHandler(svc)

			resp, err := handler.GetAppointment(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHandler_ListAppointments(t *testing.T) {
	tests := []struct {
		name     string
		req      *appointmentv1.ListAppointmentsRequest
		mockFn   func(*MockRepository)
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "successful list",
			req: &appointmentv1.ListAppointmentsRequest{
				PatientId: "pat-123",
				PageSize:  10,
			},
			mockFn: func(mr *MockRepository) {
				mr.On("ListAppointmentsByPatient", mock.Anything, "pat-123", int32(10), int32(0)).Return(
					[]*fhir.Appointment{
						createTestAppointment("apt-1", "pat-123", 24*time.Hour),
					},
					"offset:10",
					nil,
				)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, new(MockPublisher), tracer)
			handler := NewHandler(svc)

			resp, err := handler.ListAppointments(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHandler_CancelAppointment(t *testing.T) {
	tests := []struct {
		name     string
		req      *appointmentv1.CancelAppointmentRequest
		mockFn   func(*MockRepository, *MockPublisher)
		wantErr  bool
		wantCode codes.Code
	}{
		{
			name: "successful cancel",
			req:  &appointmentv1.CancelAppointmentRequest{Id: "apt-123"},
			mockFn: func(mr *MockRepository, mp *MockPublisher) {
				apt := createTestAppointment("apt-123", "pat-123", 24*time.Hour)
				mr.On("GetAppointment", mock.Anything, "apt-123").Return(apt, nil)
				cancelledApt := apt
				cancelledApt.Status = "cancelled"
				mr.On("UpdateAppointment", mock.Anything, mock.Anything).Return(cancelledApt, nil)
				mp.On("Publish", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockPub := new(MockPublisher)

			if tt.mockFn != nil {
				tt.mockFn(mockRepo, mockPub)
			}

			tracer := otel.Tracer("test")
			svc := NewService(mockRepo, mockPub, tracer)
			handler := NewHandler(svc)

			resp, err := handler.CancelAppointment(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// ============================================================
// CONVERSION FUNCTION TESTS
// ============================================================

func TestConversions(t *testing.T) {
	t.Run("protoToFHIR", func(t *testing.T) {
		startTime := time.Now().UTC().Add(24 * time.Hour)
		endTime := startTime.Add(time.Hour)

		protoApt := &appointmentv1.Appointment{
			Id:             "apt-123",
			Status:         "booked",
			PatientId:      "pat-123",
			PractitionerId: "prac-456",
			Start:          timestamppb.New(startTime),
			End:            timestamppb.New(endTime),
			ServiceType:    "checkup",
			Description:    "Annual checkup",
		}

		fhirApt := protoToFHIR(protoApt)

		assert.Equal(t, "apt-123", fhirApt.ID)
		assert.Equal(t, "booked", fhirApt.Status)
		assert.Equal(t, startTime.Unix(), fhirApt.Start.Unix())
		assert.Equal(t, endTime.Unix(), fhirApt.End.Unix())
		assert.Equal(t, "Annual checkup", fhirApt.Description)
	})

	t.Run("fhirToProto", func(t *testing.T) {
		startTime := time.Now().UTC().Add(24 * time.Hour)
		fhirApt := &fhir.Appointment{
			ID:     "apt-123",
			Status: "booked",
			Participant: []fhir.Participant{
				{
					Actor:  fhir.Reference{Reference: "Patient/pat-123"},
					Status: "accepted",
				},
				{
					Actor:  fhir.Reference{Reference: "Practitioner/prac-456"},
					Status: "accepted",
				},
			},
			Start: startTime,
			End:   startTime.Add(time.Hour),
			ServiceType: []fhir.CodeableConcept{
				{Text: "checkup"},
			},
			Description: "Annual checkup",
		}

		protoApt := fhirToProto(fhirApt)

		assert.Equal(t, "apt-123", protoApt.Id)
		assert.Equal(t, "pat-123", protoApt.PatientId)
		assert.Equal(t, "prac-456", protoApt.PractitionerId)
		assert.Equal(t, "checkup", protoApt.ServiceType)
	})
}
