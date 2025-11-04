package fhir

import "time"

// Observation represents a FHIR R4 Observation resource (simplified)
// Used for lab results and clinical observations
type Observation struct {
	ID                 string            `json:"id"`
	Status             string            `json:"status"` // registered, preliminary, final, amended
	Category           []CodeableConcept `json:"category,omitempty"`
	Code               CodeableConcept   `json:"code"`
	Subject            Reference         `json:"subject"`
	EffectiveDateTime  time.Time         `json:"effectiveDateTime,omitempty"`
	Issued             time.Time         `json:"issued,omitempty"`
	ValueQuantity      *Quantity         `json:"valueQuantity,omitempty"`
	ValueString        string            `json:"valueString,omitempty"`
	Interpretation     []CodeableConcept `json:"interpretation,omitempty"`
	ReferenceRange     []ReferenceRange  `json:"referenceRange,omitempty"`
	Meta               Meta              `json:"meta"`
}

// Quantity represents a measured amount
type Quantity struct {
	Value      float64 `json:"value"`
	Unit       string  `json:"unit,omitempty"`
	System     string  `json:"system,omitempty"`
	Code       string  `json:"code,omitempty"`
	Comparator string  `json:"comparator,omitempty"` // <, <=, >=, >
}

// ReferenceRange represents the reference range for an observation
type ReferenceRange struct {
	Low  *Quantity       `json:"low,omitempty"`
	High *Quantity       `json:"high,omitempty"`
	Type CodeableConcept `json:"type,omitempty"`
	Text string          `json:"text,omitempty"`
}
