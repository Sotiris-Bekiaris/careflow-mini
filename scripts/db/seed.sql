-- CareFlow-Mini Sample Data Seeding Script

-- Seed sample patients
INSERT INTO patient_svc.patients (id, fhir_resource, active) VALUES
(
    '550e8400-e29b-41d4-a716-446655440000',
    '{
        "resourceType": "Patient",
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "active": true,
        "name": {
            "family": "Smith",
            "given": ["John", "Michael"]
        },
        "gender": "male",
        "birthDate": "1980-01-15",
        "telecom": [
            {
                "system": "phone",
                "value": "555-1234",
                "use": "home"
            },
            {
                "system": "email",
                "value": "john.smith@example.com",
                "use": "home"
            }
        ],
        "address": {
            "line": ["123 Main St"],
            "city": "Springfield",
            "state": "IL",
            "postalCode": "62701",
            "country": "USA"
        }
    }'::jsonb,
    true
),
(
    '550e8400-e29b-41d4-a716-446655440001',
    '{
        "resourceType": "Patient",
        "id": "550e8400-e29b-41d4-a716-446655440001",
        "active": true,
        "name": {
            "family": "Johnson",
            "given": ["Jane", "Elizabeth"]
        },
        "gender": "female",
        "birthDate": "1990-05-22",
        "telecom": [
            {
                "system": "phone",
                "value": "555-5678",
                "use": "mobile"
            },
            {
                "system": "email",
                "value": "jane.johnson@example.com",
                "use": "work"
            }
        ],
        "address": {
            "line": ["456 Oak Ave"],
            "city": "Chicago",
            "state": "IL",
            "postalCode": "60601",
            "country": "USA"
        }
    }'::jsonb,
    true
);

-- Seed sample appointments
INSERT INTO appointment_svc.appointments (id, fhir_resource, patient_id, status, start_time, end_time) VALUES
(
    '660e8400-e29b-41d4-a716-446655440000',
    '{
        "resourceType": "Appointment",
        "id": "660e8400-e29b-41d4-a716-446655440000",
        "status": "booked",
        "serviceType": "General Checkup",
        "description": "Annual physical examination",
        "start": "2024-02-15T10:00:00Z",
        "end": "2024-02-15T11:00:00Z"
    }'::jsonb,
    '550e8400-e29b-41d4-a716-446655440000',
    'booked',
    '2024-02-15 10:00:00+00',
    '2024-02-15 11:00:00+00'
),
(
    '660e8400-e29b-41d4-a716-446655440001',
    '{
        "resourceType": "Appointment",
        "id": "660e8400-e29b-41d4-a716-446655440001",
        "status": "booked",
        "serviceType": "Lab Work",
        "description": "Blood work follow-up",
        "start": "2024-02-20T14:00:00Z",
        "end": "2024-02-20T14:30:00Z"
    }'::jsonb,
    '550e8400-e29b-41d4-a716-446655440001',
    'booked',
    '2024-02-20 14:00:00+00',
    '2024-02-20 14:30:00+00'
);

-- Seed sample lab observations
INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime) VALUES
(
    '770e8400-e29b-41d4-a716-446655440000',
    '{
        "resourceType": "Observation",
        "id": "770e8400-e29b-41d4-a716-446655440000",
        "status": "final",
        "code": {
            "coding": [{
                "system": "http://loinc.org",
                "code": "2339-0",
                "display": "Glucose"
            }],
            "text": "Glucose"
        },
        "valueQuantity": {
            "value": 95,
            "unit": "mg/dL",
            "system": "http://unitsofmeasure.org",
            "code": "mg/dL"
        },
        "effectiveDateTime": "2024-02-10T09:30:00Z"
    }'::jsonb,
    '550e8400-e29b-41d4-a716-446655440000',
    'final',
    '2024-02-10 09:30:00+00'
);
