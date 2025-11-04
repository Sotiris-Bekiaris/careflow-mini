#!/bin/bash
# CareFlow-Mini Demo Script

set -e

echo "========================================="
echo "CareFlow-Mini Demo"
echo "========================================="
echo ""

API_GATEWAY="http://localhost:8080"

# Check if API Gateway is running
echo "Checking if API Gateway is running..."
if ! curl -s "${API_GATEWAY}/health" > /dev/null; then
    echo "Error: API Gateway is not running at ${API_GATEWAY}"
    echo "Please start the services first with 'make run-dev'"
    exit 1
fi

echo "✓ API Gateway is healthy"
echo ""

# Demo 1: Create a patient
echo "========================================="
echo "Demo 1: Creating a new patient"
echo "========================================="

PATIENT_PAYLOAD='{
  "resourceType": "Patient",
  "active": true,
  "name": {
    "family": "Demo",
    "given": ["Test", "User"]
  },
  "gender": "male",
  "birthDate": "1985-03-20",
  "telecom": [
    {
      "system": "phone",
      "value": "555-9999",
      "use": "mobile"
    }
  ],
  "address": {
    "line": ["789 Demo Street"],
    "city": "DemoCity",
    "state": "DC",
    "postalCode": "12345",
    "country": "USA"
  }
}'

echo "POST /fhir/Patient"
echo "Request: ${PATIENT_PAYLOAD}"
echo ""

# TODO: Implement actual API call when endpoints are ready
# PATIENT_RESPONSE=$(curl -s -X POST "${API_GATEWAY}/fhir/Patient" \
#   -H "Content-Type: application/json" \
#   -d "${PATIENT_PAYLOAD}")
# echo "Response: ${PATIENT_RESPONSE}"

echo "[Demo] Patient creation - not yet implemented"
echo ""

# Demo 2: Create an appointment
echo "========================================="
echo "Demo 2: Creating an appointment"
echo "========================================="

APPOINTMENT_PAYLOAD='{
  "resourceType": "Appointment",
  "status": "booked",
  "serviceType": "Consultation",
  "description": "Demo consultation",
  "start": "2024-03-01T10:00:00Z",
  "end": "2024-03-01T11:00:00Z"
}'

echo "POST /fhir/Appointment"
echo "Request: ${APPOINTMENT_PAYLOAD}"
echo ""

# TODO: Implement actual API call when endpoints are ready
echo "[Demo] Appointment creation - not yet implemented"
echo ""

# Demo 3: Submit HL7 lab result
echo "========================================="
echo "Demo 3: Submitting HL7 lab result"
echo "========================================="

HL7_MESSAGE="MSH|^~\&|LAB|FACILITY|CAREFLOW|SYSTEM|20240215120000||ORU^R01|MSG001|P|2.5
PID|1||12345||Demo^Test||19850320|M
OBR|1||LAB001|CBC^Complete Blood Count
OBX|1|NM|WBC^White Blood Cells||7.5|10^9/L|4.0-11.0|N|||F"

echo "HL7 ORU^R01 Message:"
echo "${HL7_MESSAGE}"
echo ""

# TODO: Implement HL7 message submission
echo "[Demo] HL7 message submission - not yet implemented"
echo ""

# Demo 4: Check observability
echo "========================================="
echo "Demo 4: Observability Stack"
echo "========================================="
echo "Jaeger UI:      http://localhost:16686"
echo "Prometheus:     http://localhost:9090"
echo "Grafana:        http://localhost:3000 (admin/admin)"
echo "NATS Monitor:   http://localhost:8222"
echo ""

echo "========================================="
echo "Demo complete!"
echo "========================================="
