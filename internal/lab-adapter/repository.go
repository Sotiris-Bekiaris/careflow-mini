package labadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
)

// Repository defines observation data access operations
type Repository interface {
	CreateObservation(ctx context.Context, observation *fhir.Observation) (*fhir.Observation, error)
	GetObservation(ctx context.Context, id string) (*fhir.Observation, error)
	ListObservationsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Observation, string, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL-backed repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateObservation creates a new observation in the database
func (r *PostgresRepository) CreateObservation(ctx context.Context, observation *fhir.Observation) (*fhir.Observation, error) {
	if observation == nil {
		return nil, errors.New("observation is nil")
	}

	if observation.ID == "" {
		return nil, errors.New("observation ID is required")
	}

	// Extract patient ID from subject reference
	patientID := ""
	if observation.Subject.Reference != "" {
		// Format: "Patient/123"
		if len(observation.Subject.Reference) > 8 && observation.Subject.Reference[:8] == "Patient/" {
			patientID = observation.Subject.Reference[8:]
		}
	}

	if patientID == "" {
		return nil, errors.New("patient ID is required in subject reference")
	}

	observationJSON, err := json.Marshal(observation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal observation: %w", err)
	}

	query := `
		INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING fhir_resource
	`

	var result []byte
	err = r.pool.QueryRow(ctx, query,
		observation.ID,
		observationJSON,
		patientID,
		observation.Status,
		observation.EffectiveDateTime,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to insert observation: %w", err)
	}

	var createdObservation fhir.Observation
	if err := json.Unmarshal(result, &createdObservation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &createdObservation, nil
}

// GetObservation retrieves an observation by ID
func (r *PostgresRepository) GetObservation(ctx context.Context, id string) (*fhir.Observation, error) {
	if id == "" {
		return nil, errors.New("observation ID is required")
	}

	query := `
		SELECT fhir_resource
		FROM lab_svc.observations
		WHERE id = $1
	`

	var observationData []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(&observationData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("observation not found: %w", err)
		}
		return nil, fmt.Errorf("failed to query observation: %w", err)
	}

	var observation fhir.Observation
	if err := json.Unmarshal(observationData, &observation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal observation: %w", err)
	}

	return &observation, nil
}

// ListObservationsByPatient lists observations for a patient with pagination
func (r *PostgresRepository) ListObservationsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Observation, string, error) {
	if patientID == "" {
		return nil, "", errors.New("patient ID is required")
	}

	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Fetch limit + 1 to determine if there's a next page
	query := `
		SELECT fhir_resource, id
		FROM lab_svc.observations
		WHERE patient_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, patientID, limit+1, offset)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query observations: %w", err)
	}
	defer rows.Close()

	var observations []*fhir.Observation
	for i := int32(0); rows.Next() && i < limit; i++ {
		var observationData []byte
		var id string
		if err := rows.Scan(&observationData, &id); err != nil {
			return nil, "", fmt.Errorf("failed to scan observation: %w", err)
		}

		var observation fhir.Observation
		if err := json.Unmarshal(observationData, &observation); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal observation: %w", err)
		}

		observations = append(observations, &observation)
		_ = id // id used for pagination
	}

	// Check if there's a next page
	var nextPageToken string
	if len(observations) > 0 && rows.Next() {
		nextPageToken = fmt.Sprintf("offset:%d", offset+limit)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate observations: %w", err)
	}

	return observations, nextPageToken, nil
}
