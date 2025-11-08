package patient

import (
	"context"
	"errors"
	"testing"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/events"
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, patient *fhir.Patient) error {
	args := m.Called(ctx, patient)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*fhir.Patient, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*fhir.Patient), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, patient *fhir.Patient) error {
	args := m.Called(ctx, patient)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*fhir.Patient, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*fhir.Patient), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPublisher is a mock implementation of events.Publisher
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

func TestServiceCreatePatient(t *testing.T) {
	tests := []struct {
		name        string
		req         *patientv1.CreatePatientRequest
		setupMock   func(*MockRepository, *MockPublisher)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful create",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
					Gender:     "male",
					BirthDate:  "1990-01-01",
					Active:     true,
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				repo.On("Create", mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						patient := args.Get(1).(*fhir.Patient)
						patient.ID = "generated-id"
					})
				pub.On("Publish", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil patient",
			req:  &patientv1.CreatePatientRequest{Patient: nil},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "patient is required",
		},
		{
			name: "missing family name",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					GivenNames: []string{"John"},
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "family name is required",
		},
		{
			name: "missing given name",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{},
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "at least one given name is required",
		},
		{
			name: "invalid birth date format",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
					BirthDate:  "invalid-date",
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "invalid birth date format",
		},
		{
			name: "repository error",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
					Active:     true,
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				repo.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("database error"))
			},
			wantErr:     true,
			errContains: "failed to create patient",
		},
		{
			name: "event publish failure (should not fail request)",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
					Active:     true,
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				repo.On("Create", mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						patient := args.Get(1).(*fhir.Patient)
						patient.ID = "generated-id"
					})
				pub.On("Publish", mock.Anything).
					Return(errors.New("nats error"))
			},
			wantErr: false, // Event publish errors should not fail the request
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			pub := new(MockPublisher)
			tt.setupMock(repo, pub)

			svc := NewService(repo, pub, noop.NewTracerProvider().Tracer("test"))
			ctx := context.Background()

			resp, err := svc.CreatePatient(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patient)
				assert.NotEmpty(t, resp.Patient.Id)
			}

			repo.AssertExpectations(t)
			pub.AssertExpectations(t)
		})
	}
}

func TestServiceGetPatient(t *testing.T) {
	tests := []struct {
		name        string
		req         *patientv1.GetPatientRequest
		setupMock   func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful get",
			req:  &patientv1.GetPatientRequest{Id: "patient-123"},
			setupMock: func(repo *MockRepository) {
				repo.On("GetByID", mock.Anything, "patient-123").
					Return(&fhir.Patient{
						ID:     "patient-123",
						Active: true,
						Name: fhir.HumanName{
							Family: "Doe",
							Given:  []string{"John"},
						},
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			req:  &patientv1.GetPatientRequest{Id: ""},
			setupMock: func(repo *MockRepository) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "patient ID is required",
		},
		{
			name: "patient not found",
			req:  &patientv1.GetPatientRequest{Id: "non-existent"},
			setupMock: func(repo *MockRepository) {
				repo.On("GetByID", mock.Anything, "non-existent").
					Return(nil, errors.New("patient not found"))
			},
			wantErr:     true,
			errContains: "failed to get patient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			pub := new(MockPublisher)
			tt.setupMock(repo)

			svc := NewService(repo, pub, noop.NewTracerProvider().Tracer("test"))
			ctx := context.Background()

			resp, err := svc.GetPatient(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patient)
				assert.Equal(t, tt.req.Id, resp.Patient.Id)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestServiceUpdatePatient(t *testing.T) {
	tests := []struct {
		name        string
		req         *patientv1.UpdatePatientRequest
		setupMock   func(*MockRepository, *MockPublisher)
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update",
			req: &patientv1.UpdatePatientRequest{
				Patient: &patientv1.Patient{
					Id:         "patient-123",
					FamilyName: "UpdatedName",
					GivenNames: []string{"John"},
					Active:     true,
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				repo.On("Update", mock.Anything, mock.Anything).Return(nil)
				pub.On("Publish", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "nil patient",
			req:  &patientv1.UpdatePatientRequest{Patient: nil},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "patient is required",
		},
		{
			name: "missing patient ID",
			req: &patientv1.UpdatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "patient ID is required",
		},
		{
			name: "validation error",
			req: &patientv1.UpdatePatientRequest{
				Patient: &patientv1.Patient{
					Id:         "patient-123",
					GivenNames: []string{"John"},
					// Missing family name
				},
			},
			setupMock: func(repo *MockRepository, pub *MockPublisher) {
				// No calls expected
			},
			wantErr:     true,
			errContains: "family name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			pub := new(MockPublisher)
			tt.setupMock(repo, pub)

			svc := NewService(repo, pub, noop.NewTracerProvider().Tracer("test"))
			ctx := context.Background()

			resp, err := svc.UpdatePatient(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patient)
			}

			repo.AssertExpectations(t)
			pub.AssertExpectations(t)
		})
	}
}

func TestServiceListPatients(t *testing.T) {
	tests := []struct {
		name      string
		req       *patientv1.ListPatientsRequest
		setupMock func(*MockRepository)
		wantCount int
		hasNext   bool
		wantErr   bool
	}{
		{
			name: "default page size",
			req:  &patientv1.ListPatientsRequest{PageSize: 0},
			setupMock: func(repo *MockRepository) {
				// Service requests limit+1 to check for next page
				repo.On("List", mock.Anything, 11, 0).
					Return([]*fhir.Patient{
						{ID: "1", Name: fhir.HumanName{Family: "A"}},
						{ID: "2", Name: fhir.HumanName{Family: "B"}},
					}, nil)
			},
			wantCount: 2,
			hasNext:   false,
			wantErr:   false,
		},
		{
			name: "with pagination",
			req:  &patientv1.ListPatientsRequest{PageSize: 2, PageToken: ""},
			setupMock: func(repo *MockRepository) {
				// Returns 3 results (page_size + 1)
				repo.On("List", mock.Anything, 3, 0).
					Return([]*fhir.Patient{
						{ID: "1", Name: fhir.HumanName{Family: "A"}},
						{ID: "2", Name: fhir.HumanName{Family: "B"}},
						{ID: "3", Name: fhir.HumanName{Family: "C"}},
					}, nil)
			},
			wantCount: 2,
			hasNext:   true,
			wantErr:   false,
		},
		{
			name: "with page token",
			req:  &patientv1.ListPatientsRequest{PageSize: 2, PageToken: "4"},
			setupMock: func(repo *MockRepository) {
				repo.On("List", mock.Anything, 3, 4).
					Return([]*fhir.Patient{
						{ID: "5", Name: fhir.HumanName{Family: "E"}},
					}, nil)
			},
			wantCount: 1,
			hasNext:   false,
			wantErr:   false,
		},
		{
			name: "repository error",
			req:  &patientv1.ListPatientsRequest{PageSize: 10},
			setupMock: func(repo *MockRepository) {
				repo.On("List", mock.Anything, 11, 0).
					Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockRepository)
			pub := new(MockPublisher)
			tt.setupMock(repo)

			svc := NewService(repo, pub, noop.NewTracerProvider().Tracer("test"))
			ctx := context.Background()

			resp, err := svc.ListPatients(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantCount, len(resp.Patients))
				if tt.hasNext {
					assert.NotEmpty(t, resp.NextPageToken)
				} else {
					assert.Empty(t, resp.NextPageToken)
				}
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestProtoToFHIRConversion(t *testing.T) {
	proto := &patientv1.Patient{
		Id:         "test-123",
		FamilyName: "Doe",
		GivenNames: []string{"John", "Middle"},
		Gender:     "male",
		BirthDate:  "1990-01-01",
		Active:     true,
		Telecom: []*patientv1.Contact{
			{System: "phone", Value: "+1234567890", Use: "mobile"},
		},
		Address: &patientv1.Address{
			Line:       []string{"123 Main St"},
			City:       "City",
			State:      "State",
			PostalCode: "12345",
			Country:    "Country",
		},
	}

	fhir := protoToFHIR(proto)

	assert.Equal(t, proto.Id, fhir.ID)
	assert.Equal(t, proto.FamilyName, fhir.Name.Family)
	assert.Equal(t, proto.GivenNames, fhir.Name.Given)
	assert.Equal(t, proto.Gender, fhir.Gender)
	assert.Equal(t, proto.BirthDate, fhir.BirthDate)
	assert.Equal(t, proto.Active, fhir.Active)
	assert.Len(t, fhir.Telecom, 1)
	assert.Equal(t, proto.Telecom[0].System, fhir.Telecom[0].System)
	assert.Len(t, fhir.Address, 1)
	assert.Equal(t, proto.Address.City, fhir.Address[0].City)
}

func TestFHIRToProtoConversion(t *testing.T) {
	fhir := &fhir.Patient{
		ID:     "test-123",
		Active: true,
		Name: fhir.HumanName{
			Family: "Doe",
			Given:  []string{"John", "Middle"},
		},
		Gender:    "male",
		BirthDate: "1990-01-01",
		Telecom: []fhir.Contact{
			{System: "email", Value: "test@example.com", Use: "work"},
		},
		Address: []fhir.Address{
			{
				Line:       []string{"456 Oak Ave"},
				City:       "Town",
				State:      "ST",
				PostalCode: "54321",
				Country:    "USA",
			},
		},
	}

	proto := fhirToProto(fhir)

	assert.Equal(t, fhir.ID, proto.Id)
	assert.Equal(t, fhir.Name.Family, proto.FamilyName)
	assert.Equal(t, fhir.Name.Given, proto.GivenNames)
	assert.Equal(t, fhir.Gender, proto.Gender)
	assert.Equal(t, fhir.BirthDate, proto.BirthDate)
	assert.Equal(t, fhir.Active, proto.Active)
	assert.Len(t, proto.Telecom, 1)
	assert.Equal(t, fhir.Telecom[0].System, proto.Telecom[0].System)
	assert.NotNil(t, proto.Address)
	assert.Equal(t, fhir.Address[0].City, proto.Address.City)
}
