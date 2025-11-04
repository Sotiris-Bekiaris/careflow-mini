package hl7

import (
	"errors"
	"strings"
)

// Message represents a parsed HL7 message
type Message struct {
	Type     string              // Message type (e.g., "ORU^R01")
	Segments []Segment
	Raw      string
}

// Segment represents an HL7 segment
type Segment struct {
	Name   string   // Segment identifier (MSH, PID, OBR, OBX, etc.)
	Fields []string // Segment fields
}

// Parse parses an HL7 message string
func Parse(message string) (*Message, error) {
	if message == "" {
		return nil, errors.New("empty HL7 message")
	}

	lines := strings.Split(strings.ReplaceAll(message, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return nil, errors.New("no segments found")
	}

	msg := &Message{
		Raw:      message,
		Segments: make([]Segment, 0),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, "|")

		segment := Segment{
			Name:   fields[0],
			Fields: fields,
		}

		msg.Segments = append(msg.Segments, segment)

		// Extract message type from MSH segment
		if segment.Name == "MSH" && len(fields) > 8 {
			msg.Type = fields[8]
		}
	}

	return msg, nil
}

// GetSegment returns the first segment with the given name
func (m *Message) GetSegment(name string) *Segment {
	for _, seg := range m.Segments {
		if seg.Name == name {
			return &seg
		}
	}
	return nil
}

// GetSegments returns all segments with the given name
func (m *Message) GetSegments(name string) []Segment {
	segments := make([]Segment, 0)
	for _, seg := range m.Segments {
		if seg.Name == name {
			segments = append(segments, seg)
		}
	}
	return segments
}

// GetField returns the field value at the given index (0-based)
func (s *Segment) GetField(index int) string {
	if index < 0 || index >= len(s.Fields) {
		return ""
	}
	return s.Fields[index]
}
