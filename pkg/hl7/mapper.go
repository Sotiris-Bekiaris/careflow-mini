package hl7

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sotiris-Bekiaris/careflow-mini/pkg/fhir"
	"github.com/google/uuid"
)

// MapToFHIRObservation converts the first OBX segment from an HL7 ORU^R01 message to a FHIR Observation
func MapToFHIRObservation(msg *Message) (*fhir.Observation, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	obs := &fhir.Observation{
		ID:     uuid.New().String(),
		Status: "final",
		Meta: fhir.Meta{
			LastUpdated: time.Now().UTC(),
		},
	}

	// Extract patient ID from PID segment
	patientID := ""
	if pid := msg.GetSegment("PID"); pid != nil {
		pidField := pid.GetField(3) // PID-3: Patient Identifier (complex field)
		if pidField == "" {
			return nil, errors.New("patient ID not found in PID segment")
		}
		// PID-3 has components separated by ^, extract first component (the ID)
		components := strings.Split(pidField, "^")
		patientID = components[0]
		if patientID == "" {
			return nil, errors.New("patient ID not found in PID segment")
		}
		obs.Subject = fhir.Reference{
			Reference: "Patient/" + patientID,
		}
	} else {
		return nil, errors.New("PID segment not found")
	}

	// Extract effective date/time from OBR segment
	if obr := msg.GetSegment("OBR"); obr != nil {
		// OBR-7: Observation Date/Time (observation start)
		dateTime := obr.GetField(7)
		if dateTime != "" {
			if parsedTime, err := parseHL7DateTime(dateTime); err == nil {
				obs.EffectiveDateTime = parsedTime
			}
		}
		// OBR-12: Ordering Provider (issuing authority)
		// OBR-14: Specimen Received Date/Time (issued)
		issuedStr := obr.GetField(14)
		if issuedStr != "" {
			if parsedTime, err := parseHL7DateTime(issuedStr); err == nil {
				obs.Issued = parsedTime
			}
		}
	}

	// Extract observation data from first OBX segment
	obxSegments := msg.GetSegments("OBX")
	if len(obxSegments) == 0 {
		return nil, errors.New("no OBX segments found")
	}

	// Process first OBX segment for this observation
	obx := obxSegments[0]

	// OBX-3: Observation Identifier (e.g., "WBC^White Blood Cell Count")
	obsIdentifier := obx.GetField(3)
	if obsIdentifier == "" {
		return nil, errors.New("observation identifier (OBX-3) not found")
	}

	// Parse observation identifier (may contain subcomponents)
	parts := strings.Split(obsIdentifier, "^")
	obs.Code = fhir.CodeableConcept{
		Text: obsIdentifier,
	}
	if len(parts) > 0 {
		obs.Code.Coding = []fhir.Coding{
			{
				Code:    parts[0],
				Display: parts[0],
			},
		}
		if len(parts) > 1 {
			obs.Code.Coding[0].Display = parts[1]
		}
	}

	// OBX-2: Value Type (NM=Numeric, ST=String, DT=Date, etc.)
	valueType := obx.GetField(2)

	// OBX-5: Observation Value
	value := obx.GetField(5)
	if value != "" {
		switch valueType {
		case "NM": // Numeric
			if quantity, err := parseQuantity(value); err == nil {
				obs.ValueQuantity = quantity
			}
		case "ST", "TX": // String/Text
			obs.ValueString = value
		case "CWE": // Coded With Exceptions
			obs.ValueString = value
		}
	}

	// OBX-6: Units of Measure
	if obs.ValueQuantity != nil {
		units := obx.GetField(6)
		if units != "" {
			obs.ValueQuantity.Unit = units
			// Parse UCUM code if available (e.g., "10^3/uL" is the unit with UCUM code)
			obs.ValueQuantity.Code = units
		}
	}

	// OBX-7: Reference Range
	refRange := obx.GetField(7)
	if refRange != "" {
		if rr, err := parseReferenceRange(refRange); err == nil {
			obs.ReferenceRange = []fhir.ReferenceRange{rr}
		}
	}

	// OBX-8: Abnormal Flags (L, H, LL, HH, or empty for normal)
	abnormalFlags := obx.GetField(8)
	if abnormalFlags != "" {
		interpretation := fhir.CodeableConcept{
			Coding: []fhir.Coding{
				{
					Code:    abnormalFlags,
					Display: interpretationDisplay(abnormalFlags),
				},
			},
			Text: interpretationDisplay(abnormalFlags),
		}
		obs.Interpretation = []fhir.CodeableConcept{interpretation}
	}

	// Set category to laboratory
	obs.Category = []fhir.CodeableConcept{
		{
			Coding: []fhir.Coding{
				{
					System:  "http://terminology.hl7.org/CodeSystem/observation-category",
					Code:    "laboratory",
					Display: "Laboratory",
				},
			},
			Text: "Laboratory",
		},
	}

	return obs, nil
}

// parseQuantity parses a numeric value into a Quantity
func parseQuantity(value string) (*fhir.Quantity, error) {
	if value == "" {
		return nil, errors.New("empty value")
	}

	// Handle comparators: <, <=, >=, >
	comparator := ""
	cleanValue := value
	if strings.HasPrefix(value, "<=") {
		comparator = "<="
		cleanValue = strings.TrimPrefix(value, "<=")
	} else if strings.HasPrefix(value, ">=") {
		comparator = ">="
		cleanValue = strings.TrimPrefix(value, ">=")
	} else if strings.HasPrefix(value, "<") {
		comparator = "<"
		cleanValue = strings.TrimPrefix(value, "<")
	} else if strings.HasPrefix(value, ">") {
		comparator = ">"
		cleanValue = strings.TrimPrefix(value, ">")
	}

	floatVal, err := strconv.ParseFloat(strings.TrimSpace(cleanValue), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse numeric value: %w", err)
	}

	quantity := &fhir.Quantity{
		Value:      floatVal,
		Comparator: comparator,
	}

	return quantity, nil
}

// parseReferenceRange parses HL7 reference range format
// Formats: "4.0-11.0", "4.0^11.0", "4.0-" or "^11.0", "<200", ">=100"
func parseReferenceRange(rangeStr string) (fhir.ReferenceRange, error) {
	rr := fhir.ReferenceRange{}

	// Handle comparators at the start of the range string
	if strings.HasPrefix(rangeStr, "<=") || strings.HasPrefix(rangeStr, ">=") ||
		strings.HasPrefix(rangeStr, "<") || strings.HasPrefix(rangeStr, ">") {
		// Extract comparator and numeric value
		var comparator string
		var numStr string

		switch {
		case strings.HasPrefix(rangeStr, "<="):
			comparator = "<="
			numStr = strings.TrimPrefix(rangeStr, "<=")
		case strings.HasPrefix(rangeStr, ">="):
			comparator = ">="
			numStr = strings.TrimPrefix(rangeStr, ">=")
		case strings.HasPrefix(rangeStr, "<"):
			comparator = "<"
			numStr = strings.TrimPrefix(rangeStr, "<")
		case strings.HasPrefix(rangeStr, ">"):
			comparator = ">"
			numStr = strings.TrimPrefix(rangeStr, ">")
		}

		if val, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64); err == nil {
			// Determine if this is a high or low bound based on comparator
			switch comparator {
			case "<", "<=":
				rr.High = &fhir.Quantity{Value: val}
			case ">", ">=":
				rr.Low = &fhir.Quantity{Value: val}
			}
		}
		rr.Text = rangeStr
		return rr, nil
	}

	// Handle component separator (^ or -)
	var separator string
	if strings.Contains(rangeStr, "^") {
		separator = "^"
	} else if strings.Contains(rangeStr, "-") {
		separator = "-"
	} else {
		rr.Text = rangeStr
		return rr, nil
	}

	parts := strings.Split(rangeStr, separator)
	rr.Text = rangeStr

	if len(parts) >= 1 && parts[0] != "" {
		if val, err := strconv.ParseFloat(parts[0], 64); err == nil {
			rr.Low = &fhir.Quantity{Value: val}
		}
	}

	if len(parts) >= 2 && parts[1] != "" {
		if val, err := strconv.ParseFloat(parts[1], 64); err == nil {
			rr.High = &fhir.Quantity{Value: val}
		}
	}

	return rr, nil
}

// interpretationDisplay returns human-readable interpretation text
func interpretationDisplay(flags string) string {
	flagMap := map[string]string{
		"N":  "Normal",
		"L":  "Low",
		"H":  "High",
		"LL": "Critically Low",
		"HH": "Critically High",
		"A":  "Abnormal",
		"U":  "Unknown",
	}
	if display, ok := flagMap[flags]; ok {
		return display
	}
	return flags
}

// parseHL7DateTime parses HL7 datetime format (YYYYMMDD[HHMMSS[.SSS[Z]]])
func parseHL7DateTime(hl7Date string) (time.Time, error) {
	if hl7Date == "" {
		return time.Time{}, errors.New("empty datetime")
	}

	// Remove timezone indicator if present
	hl7Date = strings.TrimSuffix(hl7Date, "Z")

	var layout string
	switch len(hl7Date) {
	case 8: // YYYYMMDD
		layout = "20060102"
	case 14: // YYYYMMDDHHMMSS
		layout = "20060102150405"
	case 17: // YYYYMMDDHHMMSS.SSS
		layout = "20060102150405.000"
	default:
		return time.Time{}, fmt.Errorf("unsupported datetime format: %s", hl7Date)
	}

	return time.Parse(layout, hl7Date)
}
