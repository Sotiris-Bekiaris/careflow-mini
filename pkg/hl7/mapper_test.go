package hl7

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test parsing of HL7 ORU^R01 message with full OBX mapping
func TestMapToFHIRObservation_Complete(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||12345^^^MRN||Doe^John||19800101|M
OBR|1|ORDER001|RESULT001|CBC^Complete Blood Count|||20240101120000|
OBX|1|NM|WBC^White Blood Cell Count||7.5|10^3/uL|4.0-11.0|N|||F`

	msg, err := Parse(hl7Message)
	require.NoError(t, err)

	obs, err := MapToFHIRObservation(msg)
	require.NoError(t, err)

	// Verify observation structure
	assert.NotEmpty(t, obs.ID)
	assert.Equal(t, "final", obs.Status)
	assert.Equal(t, "Patient/12345", obs.Subject.Reference)

	// Verify code
	assert.Equal(t, "WBC^White Blood Cell Count", obs.Code.Text)
	assert.Equal(t, "WBC", obs.Code.Coding[0].Code)
	assert.Equal(t, "White Blood Cell Count", obs.Code.Coding[0].Display)

	// Verify value
	assert.NotNil(t, obs.ValueQuantity)
	assert.Equal(t, 7.5, obs.ValueQuantity.Value)
	assert.Equal(t, "10^3/uL", obs.ValueQuantity.Unit)

	// Verify interpretation
	assert.Len(t, obs.Interpretation, 1)
	assert.Equal(t, "N", obs.Interpretation[0].Coding[0].Code)
	assert.Equal(t, "Normal", obs.Interpretation[0].Coding[0].Display)

	// Verify category
	assert.Len(t, obs.Category, 1)
	assert.Equal(t, "laboratory", obs.Category[0].Coding[0].Code)

	// Verify reference range
	assert.Len(t, obs.ReferenceRange, 1)
	assert.NotNil(t, obs.ReferenceRange[0].Low)
	assert.Equal(t, 4.0, obs.ReferenceRange[0].Low.Value)
	assert.NotNil(t, obs.ReferenceRange[0].High)
	assert.Equal(t, 11.0, obs.ReferenceRange[0].High.Value)
}

func TestMapToFHIRObservation_HighValue(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||54321^^^MRN||Smith^Jane||19900115|F
OBR|1|ORDER002|RESULT002|CHOL^Cholesterol|||20240101120000|
OBX|1|NM|CHOL^Total Cholesterol||250|mg/dL|<200|H|||F`

	msg, err := Parse(hl7Message)
	require.NoError(t, err)

	obs, err := MapToFHIRObservation(msg)
	require.NoError(t, err)

	// Verify high value interpretation
	assert.Len(t, obs.Interpretation, 1)
	assert.Equal(t, "H", obs.Interpretation[0].Coding[0].Code)
	assert.Equal(t, "High", obs.Interpretation[0].Coding[0].Display)

	// Verify reference range high
	assert.NotNil(t, obs.ReferenceRange[0].High)
	assert.Equal(t, 200.0, obs.ReferenceRange[0].High.Value)
}

func TestMapToFHIRObservation_LowValue(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||99999^^^MRN||Taylor^Bob||19850520|M
OBR|1|ORDER003|RESULT003|HGB^Hemoglobin|||20240101120000|
OBX|1|NM|HGB^Hemoglobin||11.0|g/dL|13.5-17.5|L|||F`

	msg, err := Parse(hl7Message)
	require.NoError(t, err)

	obs, err := MapToFHIRObservation(msg)
	require.NoError(t, err)

	// Verify low value interpretation
	assert.Len(t, obs.Interpretation, 1)
	assert.Equal(t, "L", obs.Interpretation[0].Coding[0].Code)
	assert.Equal(t, "Low", obs.Interpretation[0].Coding[0].Display)
}

func TestMapToFHIRObservation_StringValue(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||77777^^^MRN||Brown^Alice||19750315|F
OBR|1|ORDER004|RESULT004|UA^Urinalysis|||20240101120000|
OBX|1|ST|UA-GLUC^Glucose in Urine||NEGATIVE|||||N|||F`

	msg, err := Parse(hl7Message)
	require.NoError(t, err)

	obs, err := MapToFHIRObservation(msg)
	require.NoError(t, err)

	// Verify string value
	assert.Equal(t, "NEGATIVE", obs.ValueString)
	assert.Nil(t, obs.ValueQuantity)
}

func TestMapToFHIRObservation_MissingPID(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
OBR|1|ORDER001|RESULT001|CBC||||20240101120000|
OBX|1|NM|WBC||7.5|10^3/uL|4.0-11.0||||F`

	msg, err := Parse(hl7Message)
	require.NoError(t, err)

	_, err = MapToFHIRObservation(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PID segment")
}

func TestMapToFHIRObservation_MissingOBX(t *testing.T) {
	hl7Message := `MSH|^~\&|LAB|Hospital|CAREFLOW|Hospital|20240101120000||ORU^R01|MSG001|P|2.5
PID|1||12345^^^MRN||Doe^John||19800101|M
OBR|1|ORDER001|RESULT001|CBC||||20240101120000|`

	msg, err := Parse(hl7Message)
	require.NoError(t, err)

	_, err = MapToFHIRObservation(msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OBX segment")
}

func TestParseHL7DateTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result time.Time)
	}{
		{
			name:    "date only",
			input:   "20240115",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, time.January, result.Month())
				assert.Equal(t, 15, result.Day())
			},
		},
		{
			name:    "date and time",
			input:   "20240115143025",
			wantErr: false,
			check: func(t *testing.T, result time.Time) {
				assert.Equal(t, 2024, result.Year())
				assert.Equal(t, 14, result.Hour())
				assert.Equal(t, 30, result.Minute())
				assert.Equal(t, 25, result.Second())
			},
		},
		{
			name:    "with timezone",
			input:   "20240115143025Z",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseHL7DateTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, result)
				}
			}
		})
	}
}

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantVal float64
		wantCmp string
		wantErr bool
	}{
		{
			name:    "plain number",
			input:   "7.5",
			wantVal: 7.5,
			wantCmp: "",
		},
		{
			name:    "with less than",
			input:   "<100",
			wantVal: 100,
			wantCmp: "<",
		},
		{
			name:    "with less than or equal",
			input:   "<=200",
			wantVal: 200,
			wantCmp: "<=",
		},
		{
			name:    "with greater than",
			input:   ">50",
			wantVal: 50,
			wantCmp: ">",
		},
		{
			name:    "with greater than or equal",
			input:   ">=10",
			wantVal: 10,
			wantCmp: ">=",
		},
		{
			name:    "invalid number",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseQuantity(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, result.Value)
				assert.Equal(t, tt.wantCmp, result.Comparator)
			}
		})
	}
}

func TestParseReferenceRange(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLow  float64
		wantHigh float64
		hasBoth  bool
	}{
		{
			name:     "dash separator",
			input:    "4.0-11.0",
			wantLow:  4.0,
			wantHigh: 11.0,
			hasBoth:  true,
		},
		{
			name:     "caret separator",
			input:    "13.5^17.5",
			wantLow:  13.5,
			wantHigh: 17.5,
			hasBoth:  true,
		},
		{
			name:     "low only",
			input:    "100-",
			wantLow:  100,
			wantHigh: 0,
			hasBoth:  false,
		},
		{
			name:     "high only",
			input:    "^200",
			wantLow:  0,
			wantHigh: 200,
			hasBoth:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseReferenceRange(tt.input)
			assert.NoError(t, err)

			if tt.hasBoth || tt.wantLow > 0 {
				assert.NotNil(t, result.Low)
				assert.Equal(t, tt.wantLow, result.Low.Value)
			}
			if tt.hasBoth || tt.wantHigh > 0 {
				assert.NotNil(t, result.High)
				assert.Equal(t, tt.wantHigh, result.High.Value)
			}
		})
	}
}

func TestInterpretationDisplay(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"N", "Normal"},
		{"L", "Low"},
		{"H", "High"},
		{"LL", "Critically Low"},
		{"HH", "Critically High"},
		{"A", "Abnormal"},
		{"U", "Unknown"},
		{"X", "X"}, // Unknown codes return as-is
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := interpretationDisplay(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}
