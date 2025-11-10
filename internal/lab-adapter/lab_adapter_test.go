package labadapter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
)

// Mock Repository for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateObservation(ctx context.Context, observation *fhir.Observation) (*fhir.Observation, error) {
	args := m.Called(ctx, observation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fhir.Observation), args.Error(1)
}

func (m *MockRepository) GetObservation(ctx context.Context, id string) (*fhir.Observation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fhir.Observation), args.Error(1)
}

func (m *MockRepository) ListObservationsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Observation, string, error) {
	args := m.Called(ctx, patientID, limit, offset)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).([]*fhir.Observation), args.String(1), args.Error(2)
}

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

func TestService_ProcessHL7Message_Valid(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||12345^^^MRN||Doe^John||19800101|M
OBR|1|ORDER001|RESULT001|CBC^Complete Blood Count|||20240101120000|
OBX|1|NM|WBC^White Blood Cell Count||7.5|10^3/uL|4.0-11.0|N|||F`

	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	mockRepo.On("CreateObservation", mock.Anything, mock.Anything).Return(
		&fhir.Observation{
			ID:            "obs-123",
			Status:        "final",
			Subject:       fhir.Reference{Reference: "Patient/12345"},
			Code:          fhir.CodeableConcept{Text: "WBC^White Blood Cell Count"},
			ValueQuantity: &fhir.Quantity{Value: 7.5, Unit: "10^3/uL"},
		},
		nil,
	)
	mockPub.On("Publish", mock.Anything).Return(nil)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	obs, err := svc.ProcessHL7Message(context.Background(), hl7Message)
	require.NoError(t, err)
	assert.NotNil(t, obs)
	assert.Equal(t, "obs-123", obs.ID)
	assert.Equal(t, "Patient/12345", obs.Subject.Reference)
}

func TestService_ProcessHL7Message_InvalidMessageType(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ADT^A01|MSG001|P|2.5
PID|1||12345^^^MRN||Doe^John||19800101|M`

	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	_, err := svc.ProcessHL7Message(context.Background(), hl7Message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported message type")
}

func TestService_ProcessHL7Message_EmptyMessage(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	_, err := svc.ProcessHL7Message(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestService_ProcessHL7Message_RepositoryError(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||12345^^^MRN||Doe^John||19800101|M
OBR|1|ORDER001|RESULT001|CBC||||20240101120000|
OBX|1|NM|WBC||7.5|10^3/uL|4.0-11.0||||F`

	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	mockRepo.On("CreateObservation", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	_, err := svc.ProcessHL7Message(context.Background(), hl7Message)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "persist")
}

func TestService_GetObservation(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	obs := &fhir.Observation{
		ID:      "obs-456",
		Status:  "final",
		Subject: fhir.Reference{Reference: "Patient/54321"},
	}

	mockRepo.On("GetObservation", mock.Anything, "obs-456").Return(obs, nil)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	result, err := svc.GetObservation(context.Background(), "obs-456")
	require.NoError(t, err)
	assert.Equal(t, "obs-456", result.ID)
}

func TestService_GetObservation_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	mockRepo.On("GetObservation", mock.Anything, "obs-999").Return(nil, errors.New("not found"))

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	_, err := svc.GetObservation(context.Background(), "obs-999")
	assert.Error(t, err)
}

func TestService_ListObservationsByPatient(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	obs := []*fhir.Observation{
		{ID: "obs-1", Status: "final", Subject: fhir.Reference{Reference: "Patient/pat-123"}},
		{ID: "obs-2", Status: "final", Subject: fhir.Reference{Reference: "Patient/pat-123"}},
	}

	mockRepo.On("ListObservationsByPatient", mock.Anything, "pat-123", int32(10), int32(0)).Return(obs, "", nil)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	result, token, err := svc.ListObservationsByPatient(context.Background(), "pat-123", 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "", token)
}

func TestService_ListObservationsByPatient_MissingID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	_, _, err := svc.ListObservationsByPatient(context.Background(), "", 10, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "patient ID")
}

// Integration-style test with complete HL7 message flow
func TestService_ProcessHL7Message_IntegrationStyle(t *testing.T) {
	// Complete HL7 message with all fields
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240115140302||ORU^R01|MSG12345|P|2.5
PID|1||99999^^^MRN||Smith^Robert||19700623|M
OBR|1|LAB00123|LAB00123|PANEL^Comprehensive Panel|||||20240115140302
OBX|1|NM|CHOL^Cholesterol||180|mg/dL|<200|N|||F
OBX|2|NM|HDL^HDL Cholesterol||45|mg/dL|>40|N|||F
OBX|3|NM|LDL^LDL Cholesterol||110|mg/dL|<100|H|||F`

	mockRepo := new(MockRepository)
	mockPub := new(MockPublisher)

	// Only the first OBX will be processed
	mockRepo.On("CreateObservation", mock.Anything, mock.MatchedBy(func(obs *fhir.Observation) bool {
		return obs.Code.Text == "CHOL^Cholesterol"
	})).Return(
		&fhir.Observation{
			ID:            "obs-final",
			Status:        "final",
			Subject:       fhir.Reference{Reference: "Patient/99999"},
			Code:          fhir.CodeableConcept{Text: "CHOL^Cholesterol"},
			ValueQuantity: &fhir.Quantity{Value: 180, Unit: "mg/dL"},
		},
		nil,
	)
	mockPub.On("Publish", mock.Anything).Return(nil)

	tracer := otel.Tracer("test")
	svc := NewService(mockRepo, mockPub, tracer)

	obs, err := svc.ProcessHL7Message(context.Background(), hl7Message)
	require.NoError(t, err)
	assert.Equal(t, "Patient/99999", obs.Subject.Reference)
	assert.Equal(t, float64(180), obs.ValueQuantity.Value)
	assert.Equal(t, "mg/dL", obs.ValueQuantity.Unit)
}
