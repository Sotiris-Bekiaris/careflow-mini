package fhir

import "time"

// Patient represents a FHIR R4 Patient resource (simplified)
type Patient struct {
	ID         string    `json:"id"`
	Identifier []string  `json:"identifier,omitempty"`
	Active     bool      `json:"active"`
	Name       HumanName `json:"name"`
	Gender     string    `json:"gender,omitempty"`
	BirthDate  string    `json:"birthDate,omitempty"`
	Telecom    []Contact `json:"telecom,omitempty"`
	Address    []Address `json:"address,omitempty"`
	Meta       Meta      `json:"meta"`
}

// HumanName represents a person's name
type HumanName struct {
	Use    string   `json:"use,omitempty"`
	Family string   `json:"family"`
	Given  []string `json:"given,omitempty"`
}

// Contact represents contact information
type Contact struct {
	System string `json:"system"` // phone, email, fax
	Value  string `json:"value"`
	Use    string `json:"use,omitempty"` // home, work, mobile
}

// Address represents a physical address
type Address struct {
	Use        string   `json:"use,omitempty"`
	Line       []string `json:"line,omitempty"`
	City       string   `json:"city,omitempty"`
	State      string   `json:"state,omitempty"`
	PostalCode string   `json:"postalCode,omitempty"`
	Country    string   `json:"country,omitempty"`
}

// Meta represents resource metadata
type Meta struct {
	VersionID   string    `json:"versionId,omitempty"`
	LastUpdated time.Time `json:"lastUpdated"`
}
