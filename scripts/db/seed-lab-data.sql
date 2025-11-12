-- Seed realistic lab observation data for CareFlow demo
-- This script creates sample laboratory observations for existing patients
-- Uses LOINC codes (standard lab test identifiers) and realistic value ranges

-- Function to generate observations for a patient
CREATE OR REPLACE FUNCTION seed_patient_observations(p_patient_id UUID) RETURNS VOID AS $$
DECLARE
    v_obs_id UUID;
    v_date_offset INT;
BEGIN
    -- Loop to create 8 observations over the past 30 days
    FOR i IN 1..8 LOOP
        v_obs_id := gen_random_uuid();
        v_date_offset := (i - 1) * 4; -- Space observations 4 days apart

        -- Choose observation based on loop counter for variety
        CASE (i % 8)
            -- White Blood Cell Count
            WHEN 0 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', CASE WHEN random() < 0.8 THEN 'final' ELSE 'preliminary' END,
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '6690-2',
                                    'display', 'White Blood Cell Count'
                                )
                            ),
                            'text', 'White Blood Cell Count'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((4.0 + random() * 7.0)::numeric, 1), -- 4.0-11.0
                            'unit', '10^3/uL',
                            'system', 'http://unitsofmeasure.org',
                            'code', '10*3/uL'
                        )
                    ),
                    p_patient_id,
                    CASE WHEN random() < 0.8 THEN 'final' ELSE 'preliminary' END,
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Hemoglobin
            WHEN 1 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', 'final',
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '718-7',
                                    'display', 'Hemoglobin'
                                )
                            ),
                            'text', 'Hemoglobin'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((12.0 + random() * 6.0)::numeric, 1), -- 12.0-18.0
                            'unit', 'g/dL',
                            'system', 'http://unitsofmeasure.org',
                            'code', 'g/dL'
                        )
                    ),
                    p_patient_id,
                    'final',
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Glucose
            WHEN 2 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', 'final',
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '2345-7',
                                    'display', 'Glucose'
                                )
                            ),
                            'text', 'Glucose'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((70.0 + random() * 60.0)::numeric, 0), -- 70-130
                            'unit', 'mg/dL',
                            'system', 'http://unitsofmeasure.org',
                            'code', 'mg/dL'
                        )
                    ),
                    p_patient_id,
                    'final',
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Total Cholesterol
            WHEN 3 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', 'final',
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '2093-3',
                                    'display', 'Total Cholesterol'
                                )
                            ),
                            'text', 'Total Cholesterol'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((150.0 + random() * 100.0)::numeric, 0), -- 150-250
                            'unit', 'mg/dL',
                            'system', 'http://unitsofmeasure.org',
                            'code', 'mg/dL'
                        )
                    ),
                    p_patient_id,
                    'final',
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Creatinine
            WHEN 4 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', 'final',
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '2160-0',
                                    'display', 'Creatinine'
                                )
                            ),
                            'text', 'Creatinine'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((0.6 + random() * 0.8)::numeric, 2), -- 0.6-1.4
                            'unit', 'mg/dL',
                            'system', 'http://unitsofmeasure.org',
                            'code', 'mg/dL'
                        )
                    ),
                    p_patient_id,
                    'final',
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Platelet Count
            WHEN 5 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', 'final',
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '777-3',
                                    'display', 'Platelet Count'
                                )
                            ),
                            'text', 'Platelet Count'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((150.0 + random() * 300.0)::numeric, 0), -- 150-450
                            'unit', '10^3/uL',
                            'system', 'http://unitsofmeasure.org',
                            'code', '10*3/uL'
                        )
                    ),
                    p_patient_id,
                    'final',
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Sodium
            WHEN 6 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', 'final',
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '2951-2',
                                    'display', 'Sodium'
                                )
                            ),
                            'text', 'Sodium'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((135.0 + random() * 10.0)::numeric, 0), -- 135-145
                            'unit', 'mmol/L',
                            'system', 'http://unitsofmeasure.org',
                            'code', 'mmol/L'
                        )
                    ),
                    p_patient_id,
                    'final',
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );

            -- Potassium
            WHEN 7 THEN
                INSERT INTO lab_svc.observations (id, fhir_resource, patient_id, status, effective_datetime)
                VALUES (
                    v_obs_id,
                    jsonb_build_object(
                        'resourceType', 'Observation',
                        'id', v_obs_id::text,
                        'status', CASE WHEN random() < 0.9 THEN 'final' ELSE 'preliminary' END,
                        'code', jsonb_build_object(
                            'coding', jsonb_build_array(
                                jsonb_build_object(
                                    'system', 'http://loinc.org',
                                    'code', '2823-3',
                                    'display', 'Potassium'
                                )
                            ),
                            'text', 'Potassium'
                        ),
                        'subject', jsonb_build_object(
                            'reference', 'Patient/' || p_patient_id::text
                        ),
                        'effectiveDateTime', (NOW() - (v_date_offset || ' days')::INTERVAL)::text,
                        'valueQuantity', jsonb_build_object(
                            'value', ROUND((3.5 + random() * 1.5)::numeric, 1), -- 3.5-5.0
                            'unit', 'mmol/L',
                            'system', 'http://unitsofmeasure.org',
                            'code', 'mmol/L'
                        )
                    ),
                    p_patient_id,
                    CASE WHEN random() < 0.9 THEN 'final' ELSE 'preliminary' END,
                    NOW() - (v_date_offset || ' days')::INTERVAL
                );
        END CASE;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Seed observations for all existing patients
DO $$
DECLARE
    patient_rec RECORD;
    patient_count INT := 0;
BEGIN
    FOR patient_rec IN
        SELECT id FROM patient_svc.patients
    LOOP
        PERFORM seed_patient_observations(patient_rec.id);
        patient_count := patient_count + 1;
    END LOOP;

    RAISE NOTICE 'Seeded lab observations for % patient(s)', patient_count;
END $$;

-- Clean up function
DROP FUNCTION IF EXISTS seed_patient_observations(UUID);

-- Display summary
SELECT
    COUNT(*) as total_observations,
    COUNT(DISTINCT patient_id) as patients_with_obs,
    COUNT(CASE WHEN status = 'final' THEN 1 END) as final_status,
    COUNT(CASE WHEN status = 'preliminary' THEN 1 END) as preliminary_status,
    MIN(effective_datetime)::date as oldest_observation,
    MAX(effective_datetime)::date as newest_observation
FROM lab_svc.observations;
