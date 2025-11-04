-- CareFlow-Mini Database Initialization Script

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For text search

-- Create schemas for each service
CREATE SCHEMA IF NOT EXISTS patient_svc;
CREATE SCHEMA IF NOT EXISTS appointment_svc;
CREATE SCHEMA IF NOT EXISTS lab_svc;

-- Patient Service Tables
CREATE TABLE IF NOT EXISTS patient_svc.patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fhir_resource JSONB NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

-- Indexes for patient queries
CREATE INDEX IF NOT EXISTS idx_patients_fhir_gin ON patient_svc.patients USING GIN (fhir_resource);
CREATE INDEX IF NOT EXISTS idx_patients_active ON patient_svc.patients (active);
CREATE INDEX IF NOT EXISTS idx_patients_created_at ON patient_svc.patients (created_at);

-- Appointment Service Tables
CREATE TABLE IF NOT EXISTS appointment_svc.appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fhir_resource JSONB NOT NULL,
    patient_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    version INTEGER DEFAULT 1
);

-- Indexes for appointment queries
CREATE INDEX IF NOT EXISTS idx_appointments_patient_id ON appointment_svc.appointments (patient_id);
CREATE INDEX IF NOT EXISTS idx_appointments_status ON appointment_svc.appointments (status);
CREATE INDEX IF NOT EXISTS idx_appointments_start_time ON appointment_svc.appointments (start_time);
CREATE INDEX IF NOT EXISTS idx_appointments_fhir_gin ON appointment_svc.appointments USING GIN (fhir_resource);

-- Lab Service Tables
CREATE TABLE IF NOT EXISTS lab_svc.observations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fhir_resource JSONB NOT NULL,
    patient_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    effective_datetime TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for observation queries
CREATE INDEX IF NOT EXISTS idx_observations_patient_id ON lab_svc.observations (patient_id);
CREATE INDEX IF NOT EXISTS idx_observations_status ON lab_svc.observations (status);
CREATE INDEX IF NOT EXISTS idx_observations_effective_datetime ON lab_svc.observations (effective_datetime);
CREATE INDEX IF NOT EXISTS idx_observations_fhir_gin ON lab_svc.observations USING GIN (fhir_resource);

-- Event log table (for event sourcing/auditing)
CREATE TABLE IF NOT EXISTS public.event_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL,
    aggregate_id UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_log_aggregate ON public.event_log (aggregate_id, aggregate_type);
CREATE INDEX IF NOT EXISTS idx_event_log_event_type ON public.event_log (event_type);
CREATE INDEX IF NOT EXISTS idx_event_log_created_at ON public.event_log (created_at);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at triggers
CREATE TRIGGER update_patients_updated_at BEFORE UPDATE ON patient_svc.patients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_appointments_updated_at BEFORE UPDATE ON appointment_svc.appointments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_observations_updated_at BEFORE UPDATE ON lab_svc.observations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
