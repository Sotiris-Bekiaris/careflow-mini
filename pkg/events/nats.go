package events

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// NATSConfig holds configuration for NATS connection.
type NATSConfig struct {
	URL           string
	StreamName    string
	MaxReconnects int
}

// DefaultNATSConfig returns default NATS configuration.
func DefaultNATSConfig() NATSConfig {
	return NATSConfig{
		URL:           nats.DefaultURL,
		StreamName:    "CAREFLOW_EVENTS",
		MaxReconnects: 10,
	}
}

// natsPublisher implements Publisher using NATS JetStream.
type natsPublisher struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	stream string
}

// NewNATSPublisher creates a new NATS-backed event publisher.
// It establishes a connection to NATS and initializes JetStream.
func NewNATSPublisher(cfg NATSConfig) (Publisher, error) {
	// Connect to NATS with options
	opts := []nats.Option{
		nats.Name("careflow-publisher"),
		nats.MaxReconnects(cfg.MaxReconnects),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Get JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// Ensure stream exists
	if err := ensureStream(js, cfg.StreamName); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ensure stream: %w", err)
	}

	return &natsPublisher{
		conn:   conn,
		js:     js,
		stream: cfg.StreamName,
	}, nil
}

// Publish publishes an event to NATS JetStream.
func (p *natsPublisher) Publish(event Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to JetStream with subject based on event type
	subject := fmt.Sprintf("%s.%s", p.stream, event.Type)

	_, err = p.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish event to %s: %w", subject, err)
	}

	return nil
}

// Close closes the NATS connection.
func (p *natsPublisher) Close() error {
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

// natsSubscriber implements Subscriber using NATS JetStream.
type natsSubscriber struct {
	conn          *nats.Conn
	js            nats.JetStreamContext
	stream        string
	subscriptions []*nats.Subscription
}

// NewNATSSubscriber creates a new NATS-backed event subscriber.
func NewNATSSubscriber(cfg NATSConfig) (Subscriber, error) {
	// Connect to NATS with options
	opts := []nats.Option{
		nats.Name("careflow-subscriber"),
		nats.MaxReconnects(cfg.MaxReconnects),
	}

	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Get JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// Ensure stream exists
	if err := ensureStream(js, cfg.StreamName); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ensure stream: %w", err)
	}

	return &natsSubscriber{
		conn:          conn,
		js:            js,
		stream:        cfg.StreamName,
		subscriptions: make([]*nats.Subscription, 0),
	}, nil
}

// Subscribe subscribes to events of a specific type and processes them with the handler.
func (s *natsSubscriber) Subscribe(eventType EventType, handler EventHandler) error {
	subject := fmt.Sprintf("%s.%s", s.stream, eventType)

	// Create durable consumer for the subscription
	durableName := fmt.Sprintf("consumer-%s", eventType)

	sub, err := s.js.Subscribe(subject, func(msg *nats.Msg) {
		var event Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			// Log error but acknowledge message to prevent redelivery
			_ = msg.Ack()
			return
		}

		// Process the event with the handler
		if err := handler(event); err != nil {
			// NACK the message for redelivery
			_ = msg.Nak()
			return
		}

		// Acknowledge successful processing
		_ = msg.Ack()
	}, nats.Durable(durableName), nats.ManualAck())

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	s.subscriptions = append(s.subscriptions, sub)
	return nil
}

// Close unsubscribes from all subscriptions and closes the NATS connection.
func (s *natsSubscriber) Close() error {
	for _, sub := range s.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}

	if s.conn != nil {
		s.conn.Close()
	}

	return nil
}

// ensureStream creates the JetStream stream if it doesn't exist.
func ensureStream(js nats.JetStreamContext, streamName string) error {
	// Check if stream already exists
	_, err := js.StreamInfo(streamName)
	if err == nil {
		// Stream exists
		return nil
	}

	// Create the stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{fmt.Sprintf("%s.>", streamName)},
		Storage:  nats.FileStorage,
		MaxAge:   0, // No expiration
		Replicas: 1,
	})

	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	return nil
}
