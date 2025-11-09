package appointment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines appointment data access operations
type Repository interface {
	CreateAppointment(ctx context.Context, appointment *fhir.Appointment) (*fhir.Appointment, error)
	GetAppointment(ctx context.Context, id string) (*fhir.Appointment, error)
	UpdateAppointment(ctx context.Context, appointment *fhir.Appointment) (*fhir.Appointment, error)
	ListAppointmentsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Appointment, string, error)
	DeleteAppointment(ctx context.Context, id string) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL-backed repository
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateAppointment creates a new appointment in the database
func (r *PostgresRepository) CreateAppointment(ctx context.Context, appointment *fhir.Appointment) (*fhir.Appointment, error) {
	if appointment.ID == "" {
		appointment.ID = uuid.New().String()
	}

	if appointment.Meta.LastUpdated.IsZero() {
		appointment.Meta.LastUpdated = time.Now().UTC()
	}

	// Extract patient ID from participant
	var patientID string
	for _, p := range appointment.Participant {
		if p.Actor.Reference != "" && len(p.Actor.Reference) > 8 && p.Actor.Reference[:8] == "Patient/" {
			patientID = p.Actor.Reference[8:]
			break
		}
	}

	if patientID == "" {
		return nil, errors.New("patient ID is required in participant")
	}

	appointmentJSON, err := json.Marshal(appointment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal appointment: %w", err)
	}

	query := `
		INSERT INTO appointment_svc.appointments (id, fhir_resource, status, patient_id, start_time, end_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING fhir_resource
	`

	var result []byte
	err = r.pool.QueryRow(ctx, query,
		appointment.ID,
		appointmentJSON,
		appointment.Status,
		patientID,
		appointment.Start,
		appointment.End,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&result)

	if err != nil {
		return nil, fmt.Errorf("failed to insert appointment: %w", err)
	}

	var createdAppointment fhir.Appointment
	if err := json.Unmarshal(result, &createdAppointment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &createdAppointment, nil
}

// GetAppointment retrieves an appointment by ID
func (r *PostgresRepository) GetAppointment(ctx context.Context, id string) (*fhir.Appointment, error) {
	query := `
		SELECT fhir_resource
		FROM appointment_svc.appointments
		WHERE id = $1
	`

	var appointmentData []byte
	err := r.pool.QueryRow(ctx, query, id).Scan(&appointmentData)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("appointment not found: %w", err)
		}
		return nil, fmt.Errorf("failed to query appointment: %w", err)
	}

	var appointment fhir.Appointment
	if err := json.Unmarshal(appointmentData, &appointment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal appointment: %w", err)
	}

	return &appointment, nil
}

// UpdateAppointment updates an existing appointment
func (r *PostgresRepository) UpdateAppointment(ctx context.Context, appointment *fhir.Appointment) (*fhir.Appointment, error) {
	if appointment.ID == "" {
		return nil, errors.New("appointment ID is required for update")
	}

	appointment.Meta.LastUpdated = time.Now().UTC()

	// Extract patient ID from participant
	var patientID string
	for _, p := range appointment.Participant {
		if p.Actor.Reference != "" && len(p.Actor.Reference) > 8 && p.Actor.Reference[:8] == "Patient/" {
			patientID = p.Actor.Reference[8:]
			break
		}
	}

	if patientID == "" {
		return nil, errors.New("patient ID is required in participant")
	}

	appointmentJSON, err := json.Marshal(appointment)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal appointment: %w", err)
	}

	query := `
		UPDATE appointment_svc.appointments
		SET fhir_resource = $2, status = $3, patient_id = $4, start_time = $5, end_time = $6, updated_at = $7
		WHERE id = $1
		RETURNING fhir_resource
	`

	var result []byte
	err = r.pool.QueryRow(ctx, query,
		appointment.ID,
		appointmentJSON,
		appointment.Status,
		patientID,
		appointment.Start,
		appointment.End,
		time.Now().UTC(),
	).Scan(&result)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("appointment not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update appointment: %w", err)
	}

	var updatedAppointment fhir.Appointment
	if err := json.Unmarshal(result, &updatedAppointment); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &updatedAppointment, nil
}

// ListAppointmentsByPatient lists appointments for a patient with pagination
func (r *PostgresRepository) ListAppointmentsByPatient(ctx context.Context, patientID string, limit int32, offset int32) ([]*fhir.Appointment, string, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Fetch limit + 1 to determine if there's a next page
	query := `
		SELECT fhir_resource, id
		FROM appointment_svc.appointments
		WHERE patient_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, patientID, limit+1, offset)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query appointments: %w", err)
	}
	defer rows.Close()

	var appointments []*fhir.Appointment
	for i := int32(0); rows.Next() && i < limit; i++ {
		var appointmentData []byte
		var id string
		if err := rows.Scan(&appointmentData, &id); err != nil {
			return nil, "", fmt.Errorf("failed to scan appointment: %w", err)
		}

		var appointment fhir.Appointment
		if err := json.Unmarshal(appointmentData, &appointment); err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal appointment: %w", err)
		}

		appointments = append(appointments, &appointment)
		_ = id // id is used for pagination tracking
	}

	// Check if there's a next page
	var nextPageToken string
	if len(appointments) > 0 && rows.Next() {
		// There's another row, so there's a next page
		nextPageToken = fmt.Sprintf("offset:%d", offset+limit)
	}

	if err = rows.Err(); err != nil {
		return nil, "", fmt.Errorf("failed to iterate appointments: %w", err)
	}

	return appointments, nextPageToken, nil
}

// DeleteAppointment deletes an appointment
func (r *PostgresRepository) DeleteAppointment(ctx context.Context, id string) error {
	query := `
		DELETE FROM appointment_svc.appointments
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete appointment: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("appointment not found")
	}

	return nil
}
