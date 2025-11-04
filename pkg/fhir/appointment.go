package fhir

import "time"

// Appointment represents a FHIR R4 Appointment resource (simplified)
type Appointment struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"` // proposed, pending, booked, arrived, fulfilled, cancelled
	ServiceType []CodeableConcept `json:"serviceType,omitempty"`
	Start       time.Time         `json:"start"`
	End         time.Time         `json:"end"`
	Participant []Participant     `json:"participant"`
	Description string            `json:"description,omitempty"`
	Meta        Meta              `json:"meta"`
}

// Participant represents an appointment participant
type Participant struct {
	Actor    Reference `json:"actor"`
	Required string    `json:"required,omitempty"` // required, optional, information-only
	Status   string    `json:"status"`             // accepted, declined, tentative, needs-action
}

// Reference represents a reference to another resource
type Reference struct {
	Reference string `json:"reference"`
	Display   string `json:"display,omitempty"`
}

// CodeableConcept represents a coded value
type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"`
	Text   string   `json:"text,omitempty"`
}

// Coding represents a code defined by a terminology system
type Coding struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}
