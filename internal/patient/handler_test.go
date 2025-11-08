package patient

import (
	"context"
	"errors"
	"testing"

	patientv1 "github.com/Sotiris-Bekiaris/careflow-mini/proto/patient/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockService is a mock implementation of Service
type MockService struct {
	mock.Mock
}

func (m *MockService) CreatePatient(ctx context.Context, req *patientv1.CreatePatientRequest) (*patientv1.CreatePatientResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*patientv1.CreatePatientResponse), args.Error(1)
}

func (m *MockService) GetPatient(ctx context.Context, req *patientv1.GetPatientRequest) (*patientv1.GetPatientResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*patientv1.GetPatientResponse), args.Error(1)
}

func (m *MockService) UpdatePatient(ctx context.Context, req *patientv1.UpdatePatientRequest) (*patientv1.UpdatePatientResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*patientv1.UpdatePatientResponse), args.Error(1)
}

func (m *MockService) ListPatients(ctx context.Context, req *patientv1.ListPatientsRequest) (*patientv1.ListPatientsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*patientv1.ListPatientsResponse), args.Error(1)
}

func TestHandlerCreatePatient(t *testing.T) {
	tests := []struct {
		name      string
		req       *patientv1.CreatePatientRequest
		setupMock func(*MockService)
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "successful create",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
					Active:     true,
				},
			},
			setupMock: func(svc *MockService) {
				svc.On("CreatePatient", mock.Anything, mock.Anything).
					Return(&patientv1.CreatePatientResponse{
						Patient: &patientv1.Patient{
							Id:         "patient-123",
							FamilyName: "Doe",
							GivenNames: []string{"John"},
							Active:     true,
						},
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "validation error",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					GivenNames: []string{"John"},
					// Missing family name
				},
			},
			setupMock: func(svc *MockService) {
				svc.On("CreatePatient", mock.Anything, mock.Anything).
					Return(nil, errors.New("validation failed: family name is required"))
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
		{
			name: "internal error",
			req: &patientv1.CreatePatientRequest{
				Patient: &patientv1.Patient{
					FamilyName: "Doe",
					GivenNames: []string{"John"},
					Active:     true,
				},
			},
			setupMock: func(svc *MockService) {
				svc.On("CreatePatient", mock.Anything, mock.Anything).
					Return(nil, errors.New("database connection failed"))
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockService)
			tt.setupMock(svc)

			handler := NewHandler(svc)
			ctx := context.Background()

			resp, err := handler.CreatePatient(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patient)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestHandlerGetPatient(t *testing.T) {
	tests := []struct {
		name      string
		req       *patientv1.GetPatientRequest
		setupMock func(*MockService)
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "successful get",
			req:  &patientv1.GetPatientRequest{Id: "patient-123"},
			setupMock: func(svc *MockService) {
				svc.On("GetPatient", mock.Anything, mock.Anything).
					Return(&patientv1.GetPatientResponse{
						Patient: &patientv1.Patient{
							Id:         "patient-123",
							FamilyName: "Doe",
							GivenNames: []string{"John"},
						},
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "patient not found",
			req:  &patientv1.GetPatientRequest{Id: "non-existent"},
			setupMock: func(svc *MockService) {
				svc.On("GetPatient", mock.Anything, mock.Anything).
					Return(nil, errors.New("patient not found: non-existent"))
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
		{
			name: "missing ID",
			req:  &patientv1.GetPatientRequest{Id: ""},
			setupMock: func(svc *MockService) {
				svc.On("GetPatient", mock.Anything, mock.Anything).
					Return(nil, errors.New("patient ID is required"))
			},
			wantErr:  true,
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockService)
			tt.setupMock(svc)

			handler := NewHandler(svc)
			ctx := context.Background()

			resp, err := handler.GetPatient(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patient)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestHandlerUpdatePatient(t *testing.T) {
	tests := []struct {
		name      string
		req       *patientv1.UpdatePatientRequest
		setupMock func(*MockService)
		wantErr   bool
		wantCode  codes.Code
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
			setupMock: func(svc *MockService) {
				svc.On("UpdatePatient", mock.Anything, mock.Anything).
					Return(&patientv1.UpdatePatientResponse{
						Patient: &patientv1.Patient{
							Id:         "patient-123",
							FamilyName: "UpdatedName",
							GivenNames: []string{"John"},
							Active:     true,
						},
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "patient not found",
			req: &patientv1.UpdatePatientRequest{
				Patient: &patientv1.Patient{
					Id:         "non-existent",
					FamilyName: "Doe",
					GivenNames: []string{"John"},
				},
			},
			setupMock: func(svc *MockService) {
				svc.On("UpdatePatient", mock.Anything, mock.Anything).
					Return(nil, errors.New("patient not found: non-existent"))
			},
			wantErr:  true,
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockService)
			tt.setupMock(svc)

			handler := NewHandler(svc)
			ctx := context.Background()

			resp, err := handler.UpdatePatient(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patient)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestHandlerListPatients(t *testing.T) {
	tests := []struct {
		name      string
		req       *patientv1.ListPatientsRequest
		setupMock func(*MockService)
		wantErr   bool
		wantCode  codes.Code
	}{
		{
			name: "successful list",
			req:  &patientv1.ListPatientsRequest{PageSize: 10},
			setupMock: func(svc *MockService) {
				svc.On("ListPatients", mock.Anything, mock.Anything).
					Return(&patientv1.ListPatientsResponse{
						Patients: []*patientv1.Patient{
							{Id: "1", FamilyName: "Doe", GivenNames: []string{"John"}},
							{Id: "2", FamilyName: "Smith", GivenNames: []string{"Jane"}},
						},
						NextPageToken: "",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "empty list",
			req:  &patientv1.ListPatientsRequest{PageSize: 10},
			setupMock: func(svc *MockService) {
				svc.On("ListPatients", mock.Anything, mock.Anything).
					Return(&patientv1.ListPatientsResponse{
						Patients:      []*patientv1.Patient{},
						NextPageToken: "",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "internal error",
			req:  &patientv1.ListPatientsRequest{PageSize: 10},
			setupMock: func(svc *MockService) {
				svc.On("ListPatients", mock.Anything, mock.Anything).
					Return(nil, errors.New("database error"))
			},
			wantErr:  true,
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockService)
			tt.setupMock(svc)

			handler := NewHandler(svc)
			ctx := context.Background()

			resp, err := handler.ListPatients(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Patients)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{
			name:     "not found error",
			err:      errors.New("patient not found"),
			wantCode: codes.NotFound,
		},
		{
			name:     "validation error",
			err:      errors.New("validation failed: invalid input"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "required field error",
			err:      errors.New("field is required"),
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "already exists error",
			err:      errors.New("patient already exists"),
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "duplicate error",
			err:      errors.New("duplicate key violation"),
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "generic error",
			err:      errors.New("something went wrong"),
			wantCode: codes.Internal,
		},
		{
			name:     "nil error",
			err:      nil,
			wantCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcErr := mapError(tt.err)

			if tt.err == nil {
				assert.Nil(t, grpcErr)
				return
			}

			st, ok := status.FromError(grpcErr)
			require.True(t, ok)
			assert.Equal(t, tt.wantCode, st.Code())
		})
	}
}
