package observation

import (
	"math/rand"
	"time"
)

// LabTestTemplate defines a template for generating realistic lab observations
type LabTestTemplate struct {
	LoincCode   string
	Display     string
	System      string
	MinValue    float64
	MaxValue    float64
	Unit        string
	UnitSystem  string
	UnitCode    string
	Decimals    int // Number of decimal places for the value
	StatusFinal bool
}

// labTestTemplates contains all available lab test templates with realistic ranges
var labTestTemplates = []LabTestTemplate{
	{
		LoincCode:   "6690-2",
		Display:     "White Blood Cell Count",
		System:      "http://loinc.org",
		MinValue:    4.0,
		MaxValue:    11.0,
		Unit:        "10^3/uL",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "10*3/uL",
		Decimals:    1,
		StatusFinal: true,
	},
	{
		LoincCode:   "718-7",
		Display:     "Hemoglobin",
		System:      "http://loinc.org",
		MinValue:    12.0,
		MaxValue:    18.0,
		Unit:        "g/dL",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "g/dL",
		Decimals:    1,
		StatusFinal: true,
	},
	{
		LoincCode:   "2345-7",
		Display:     "Glucose",
		System:      "http://loinc.org",
		MinValue:    70.0,
		MaxValue:    130.0,
		Unit:        "mg/dL",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "mg/dL",
		Decimals:    0,
		StatusFinal: true,
	},
	{
		LoincCode:   "2093-3",
		Display:     "Total Cholesterol",
		System:      "http://loinc.org",
		MinValue:    150.0,
		MaxValue:    250.0,
		Unit:        "mg/dL",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "mg/dL",
		Decimals:    0,
		StatusFinal: true,
	},
	{
		LoincCode:   "2160-0",
		Display:     "Creatinine",
		System:      "http://loinc.org",
		MinValue:    0.6,
		MaxValue:    1.4,
		Unit:        "mg/dL",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "mg/dL",
		Decimals:    2,
		StatusFinal: true,
	},
	{
		LoincCode:   "777-3",
		Display:     "Platelet Count",
		System:      "http://loinc.org",
		MinValue:    150.0,
		MaxValue:    450.0,
		Unit:        "10^3/uL",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "10*3/uL",
		Decimals:    0,
		StatusFinal: true,
	},
	{
		LoincCode:   "2951-2",
		Display:     "Sodium",
		System:      "http://loinc.org",
		MinValue:    135.0,
		MaxValue:    145.0,
		Unit:        "mmol/L",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "mmol/L",
		Decimals:    0,
		StatusFinal: true,
	},
	{
		LoincCode:   "2823-3",
		Display:     "Potassium",
		System:      "http://loinc.org",
		MinValue:    3.5,
		MaxValue:    5.0,
		Unit:        "mmol/L",
		UnitSystem:  "http://unitsofmeasure.org",
		UnitCode:    "mmol/L",
		Decimals:    1,
		StatusFinal: false, // 90% final, 10% preliminary
	},
}

// generateRandomValue generates a random value within the template's range
func (t *LabTestTemplate) generateRandomValue() float64 {
	value := t.MinValue + rand.Float64()*(t.MaxValue-t.MinValue)

	// Round to specified decimal places
	multiplier := 1.0
	for i := 0; i < t.Decimals; i++ {
		multiplier *= 10.0
	}

	return float64(int(value*multiplier+0.5)) / multiplier
}

// generateStatus generates a status based on the template's settings
func (t *LabTestTemplate) generateStatus() string {
	if t.StatusFinal {
		// 80% final, 20% preliminary
		if rand.Float64() < 0.8 {
			return "final"
		}
		return "preliminary"
	}

	// 90% final, 10% preliminary for tests marked as not always final
	if rand.Float64() < 0.9 {
		return "final"
	}
	return "preliminary"
}

// generateObservationDateTimes generates a list of datetime stamps
// spanning the past 30 days, spaced 4 days apart
func generateObservationDateTimes(count int) []time.Time {
	now := time.Now().UTC()
	dates := make([]time.Time, count)

	for i := 0; i < count; i++ {
		daysAgo := i * 4
		dates[i] = now.AddDate(0, 0, -daysAgo)
	}

	return dates
}
