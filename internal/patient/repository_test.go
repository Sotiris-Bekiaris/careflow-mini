package patient

import (
	"context"
	"testing"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts("../../scripts/db/init.sql"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		_ = postgresContainer.Terminate(ctx)
	}

	return pool, cleanup
}

func TestRepositoryCreate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(pool)
	ctx := context.Background()

	tests := []struct {
		name    string
		patient *fhir.Patient
		wantErr bool
	}{
		{
			name: "valid patient",
			patient: &fhir.Patient{
				Active: true,
				Name: fhir.HumanName{
					Family: "Doe",
					Given:  []string{"John"},
				},
				Gender:    "male",
				BirthDate: "1990-01-01",
			},
			wantErr: false,
		},
		{
			name: "patient with ID",
			patient: &fhir.Patient{
				ID:     "test-id-123",
				Active: true,
				Name: fhir.HumanName{
					Family: "Smith",
					Given:  []string{"Jane"},
				},
			},
			wantErr: false,
		},
		{
			name: "patient with contact info",
			patient: &fhir.Patient{
				Active: true,
				Name: fhir.HumanName{
					Family: "Johnson",
					Given:  []string{"Bob"},
				},
				Telecom: []fhir.Contact{
					{System: "phone", Value: "+1234567890", Use: "mobile"},
					{System: "email", Value: "bob@example.com", Use: "home"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.patient)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, tt.patient.ID)
			assert.False(t, tt.patient.Meta.LastUpdated.IsZero())
		})
	}
}

func TestRepositoryGetByID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create a patient first
	patient := &fhir.Patient{
		Active: true,
		Name: fhir.HumanName{
			Family: "Test",
			Given:  []string{"User"},
		},
		Gender:    "female",
		BirthDate: "1985-05-15",
	}
	require.NoError(t, repo.Create(ctx, patient))

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "existing patient",
			id:      patient.ID,
			wantErr: false,
		},
		{
			name:    "non-existent patient",
			id:      "non-existent-id",
			wantErr: true,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, patient.ID, got.ID)
			assert.Equal(t, patient.Name.Family, got.Name.Family)
			assert.Equal(t, patient.Gender, got.Gender)
			assert.Equal(t, patient.BirthDate, got.BirthDate)
		})
	}
}

func TestRepositoryUpdate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create a patient first
	patient := &fhir.Patient{
		Active: true,
		Name: fhir.HumanName{
			Family: "Original",
			Given:  []string{"Name"},
		},
		Gender: "male",
	}
	require.NoError(t, repo.Create(ctx, patient))
	originalUpdated := patient.Meta.LastUpdated

	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	tests := []struct {
		name     string
		patient  *fhir.Patient
		wantErr  bool
		validate func(t *testing.T)
	}{
		{
			name: "update name",
			patient: &fhir.Patient{
				ID:     patient.ID,
				Active: true,
				Name: fhir.HumanName{
					Family: "Updated",
					Given:  []string{"Name"},
				},
				Gender: "male",
			},
			wantErr: false,
			validate: func(t *testing.T) {
				updated, err := repo.GetByID(ctx, patient.ID)
				require.NoError(t, err)
				assert.Equal(t, "Updated", updated.Name.Family)
				assert.True(t, updated.Meta.LastUpdated.After(originalUpdated))
			},
		},
		{
			name: "update without ID",
			patient: &fhir.Patient{
				Active: true,
				Name: fhir.HumanName{
					Family: "Test",
					Given:  []string{"User"},
				},
			},
			wantErr: true,
		},
		{
			name: "update non-existent patient",
			patient: &fhir.Patient{
				ID:     "non-existent",
				Active: true,
				Name: fhir.HumanName{
					Family: "Test",
					Given:  []string{"User"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.patient)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestRepositoryList(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create multiple patients
	for i := 0; i < 5; i++ {
		patient := &fhir.Patient{
			Active: true,
			Name: fhir.HumanName{
				Family: "Patient",
				Given:  []string{string(rune('A' + i))},
			},
		}
		require.NoError(t, repo.Create(ctx, patient))
		time.Sleep(5 * time.Millisecond) // Ensure different timestamps
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
		wantErr   bool
	}{
		{
			name:      "default limit",
			limit:     0,
			offset:    0,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name:      "limit 2",
			limit:     2,
			offset:    0,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "with offset",
			limit:     2,
			offset:    2,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "offset beyond results",
			limit:     10,
			offset:    10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "large limit capped at 100",
			limit:     200,
			offset:    0,
			wantCount: 5,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patients, err := repo.List(ctx, tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(patients))

			// Verify patients are sorted by created_at DESC (newest first)
			for i := 0; i < len(patients)-1; i++ {
				assert.True(t, patients[i].Meta.LastUpdated.After(patients[i+1].Meta.LastUpdated) ||
					patients[i].Meta.LastUpdated.Equal(patients[i+1].Meta.LastUpdated))
			}
		})
	}
}

func TestRepositoryDelete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Create a patient first
	patient := &fhir.Patient{
		Active: true,
		Name: fhir.HumanName{
			Family: "ToDelete",
			Given:  []string{"Patient"},
		},
	}
	require.NoError(t, repo.Create(ctx, patient))

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "delete existing patient",
			id:      patient.ID,
			wantErr: false,
		},
		{
			name:    "delete already deleted patient",
			id:      patient.ID,
			wantErr: true,
		},
		{
			name:    "delete non-existent patient",
			id:      "non-existent-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify patient is soft-deleted (can't be retrieved)
			_, err = repo.GetByID(ctx, tt.id)
			assert.Error(t, err)
		})
	}
}

func TestRepositoryConcurrency(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(pool)
	ctx := context.Background()

	// Test concurrent creates
	numGoroutines := 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			patient := &fhir.Patient{
				Active: true,
				Name: fhir.HumanName{
					Family: "Concurrent",
					Given:  []string{string(rune('A' + index))},
				},
			}
			if err := repo.Create(ctx, patient); err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent create failed: %v", err)
	}

	// Verify all patients were created
	patients, err := repo.List(ctx, 100, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(patients), numGoroutines)
}
