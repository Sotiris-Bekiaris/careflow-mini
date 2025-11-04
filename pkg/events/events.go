package events

import "time"

// EventType represents the type of domain event
type EventType string

const (
	PatientCreated       EventType = "patient.created"
	PatientUpdated       EventType = "patient.updated"
	AppointmentCreated   EventType = "appointment.created"
	AppointmentCancelled EventType = "appointment.cancelled"
	ObservationCreated   EventType = "observation.created"
)

// Event represents a domain event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// Publisher defines the interface for publishing events
type Publisher interface {
	Publish(event Event) error
	Close() error
}

// Subscriber defines the interface for subscribing to events
type Subscriber interface {
	Subscribe(eventType EventType, handler EventHandler) error
	Close() error
}

// EventHandler is a function that handles an event
type EventHandler func(event Event) error

// NATSPublisher implements Publisher using NATS
type NATSPublisher struct {
	// TODO: Add NATS connection
}

// Publish publishes an event to NATS
func (p *NATSPublisher) Publish(event Event) error {
	// TODO: Implement NATS publishing
	return nil
}

// Close closes the NATS connection
func (p *NATSPublisher) Close() error {
	// TODO: Implement connection close
	return nil
}

// NATSSubscriber implements Subscriber using NATS
type NATSSubscriber struct {
	// TODO: Add NATS connection
}

// Subscribe subscribes to events of a specific type
func (s *NATSSubscriber) Subscribe(eventType EventType, handler EventHandler) error {
	// TODO: Implement NATS subscription
	return nil
}

// Close closes the NATS connection
func (s *NATSSubscriber) Close() error {
	// TODO: Implement connection close
	return nil
}
