package patient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for patient data access operations.
type Repository interface {
	Create(ctx context.Context, patient *fhir.Patient) error
	GetByID(ctx context.Context, id string) (*fhir.Patient, error)
	Update(ctx context.Context, patient *fhir.Patient) error
	List(ctx context.Context, limit, offset int, nameFilter string) ([]*fhir.Patient, error)
	Delete(ctx context.Context, id string) error
}

// pgRepository implements Repository using PostgreSQL.
type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgreSQL-backed patient repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{
		pool: pool,
	}
}

// Create inserts a new patient into the database.
func (r *pgRepository) Create(ctx context.Context, patient *fhir.Patient) error {
	// Generate UUID if not provided
	if patient.ID == "" {
		patient.ID = uuid.New().String()
	}

	// Set metadata
	now := time.Now()
	patient.Meta.LastUpdated = now

	// Serialize FHIR resource to JSONB
	fhirJSON, err := json.Marshal(patient)
	if err != nil {
		return fmt.Errorf("failed to marshal patient to JSON: %w", err)
	}

	query := `
		INSERT INTO patient_svc.patients (id, fhir_resource, active, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = r.pool.Exec(
		ctx,
		query,
		patient.ID,
		fhirJSON,
		patient.Active,
		now,
		now,
		1,
	)
	if err != nil {
		return fmt.Errorf("failed to create patient: %w", err)
	}

	return nil
}

// GetByID retrieves a patient by their ID.
func (r *pgRepository) GetByID(ctx context.Context, id string) (*fhir.Patient, error) {
	query := `
		SELECT fhir_resource
		FROM patient_svc.patients
		WHERE id = $1 AND active = true
	`

	var fhirJSON []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(&fhirJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("patient not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get patient: %w", err)
	}

	var patient fhir.Patient
	if err := json.Unmarshal(fhirJSON, &patient); err != nil {
		return nil, fmt.Errorf("failed to unmarshal patient: %w", err)
	}

	return &patient, nil
}

// Update updates an existing patient in the database.
func (r *pgRepository) Update(ctx context.Context, patient *fhir.Patient) error {
	if patient.ID == "" {
		return fmt.Errorf("patient ID is required for update")
	}

	// Update metadata
	patient.Meta.LastUpdated = time.Now()

	// Serialize FHIR resource to JSONB
	fhirJSON, err := json.Marshal(patient)
	if err != nil {
		return fmt.Errorf("failed to marshal patient to JSON: %w", err)
	}

	query := `
		UPDATE patient_svc.patients
		SET fhir_resource = $1, active = $2, version = version + 1
		WHERE id = $3 AND active = true
	`

	commandTag, err := r.pool.Exec(ctx, query, fhirJSON, patient.Active, patient.ID)
	if err != nil {
		return fmt.Errorf("failed to update patient: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("patient not found or already inactive: %s", patient.ID)
	}

	return nil
}

// List retrieves a paginated list of patients.
func (r *pgRepository) List(ctx context.Context, limit, offset int, nameFilter string) ([]*fhir.Patient, error) {
	// Set default limit if not provided or invalid
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Cap at 100 for performance
	}

	var query string
	var args []interface{}

	if nameFilter != "" {
		// Filter by name using ILIKE for case-insensitive search
		query = `
			SELECT fhir_resource
			FROM patient_svc.patients
			WHERE active = true
			  AND (family_name ILIKE $1 OR given_names::text ILIKE $1)
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{"%" + nameFilter + "%", limit, offset}
	} else {
		query = `
			SELECT fhir_resource
			FROM patient_svc.patients
			WHERE active = true
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list patients: %w", err)
	}
	defer rows.Close()

	patients := make([]*fhir.Patient, 0, limit)
	for rows.Next() {
		var fhirJSON []byte
		if err := rows.Scan(&fhirJSON); err != nil {
			return nil, fmt.Errorf("failed to scan patient row: %w", err)
		}

		var patient fhir.Patient
		if err := json.Unmarshal(fhirJSON, &patient); err != nil {
			return nil, fmt.Errorf("failed to unmarshal patient: %w", err)
		}

		patients = append(patients, &patient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating patient rows: %w", err)
	}

	return patients, nil
}

// Delete performs a soft delete on a patient by setting active to false.
func (r *pgRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE patient_svc.patients
		SET active = false
		WHERE id = $1 AND active = true
	`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete patient: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("patient not found or already inactive: %s", id)
	}

	return nil
}
