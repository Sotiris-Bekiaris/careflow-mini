package hl7

import (
	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
)

// MapToFHIRObservation converts an HL7 ORU^R01 message to a FHIR Observation
func MapToFHIRObservation(msg *Message) (*fhir.Observation, error) {
	// TODO: Implement full HL7 -> FHIR mapping
	// This is a placeholder that shows the mapping structure

	obs := &fhir.Observation{
		Status: "final",
	}

	// Extract patient ID from PID segment
	if pid := msg.GetSegment("PID"); pid != nil {
		patientID := pid.GetField(3) // PID-3: Patient Identifier
		obs.Subject = fhir.Reference{
			Reference: "Patient/" + patientID,
		}
	}

	// Extract observation data from OBX segments
	// TODO: Map OBX fields to FHIR Observation when implementing full HL7 mapping
	// OBX-3: Observation Identifier
	// OBX-5: Observation Value
	// OBX-6: Units
	// OBX-7: Reference Range
	_ = msg.GetSegments("OBX") // Reserved for future OBX segment mapping

	return obs, nil
}
