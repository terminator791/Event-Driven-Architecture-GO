package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventMetadata contains common metadata for all events
type EventMetadata struct {
	EventID       string            `json:"event_id"`
	CorrelationID string            `json:"correlation_id"`
	CausationID   string            `json:"causation_id,omitempty"` // ID of the event that caused this event
	Timestamp     time.Time         `json:"timestamp"`
	Version       string            `json:"version"`
	Source        string            `json:"source"`
	TraceID       string            `json:"trace_id,omitempty"`
	UserID        string            `json:"user_id,omitempty"`
	SessionID     string            `json:"session_id,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// BaseEvent provides common functionality for all events
type BaseEvent struct {
	Metadata EventMetadata `json:"metadata"`
}

// NewEventMetadata creates new event metadata with required fields
func NewEventMetadata(source, version string) EventMetadata {
	return EventMetadata{
		EventID:     uuid.New().String(),
		Timestamp:   time.Now().UTC(),
		Version:     version,
		Source:      source,
		Attributes:  make(map[string]string),
	}
}

// WithCorrelationID sets the correlation ID
func (m *EventMetadata) WithCorrelationID(correlationID string) *EventMetadata {
	m.CorrelationID = correlationID
	return m
}

// WithCausationID sets the causation ID
func (m *EventMetadata) WithCausationID(causationID string) *EventMetadata {
	m.CausationID = causationID
	return m
}

// WithTraceID sets the trace ID for distributed tracing
func (m *EventMetadata) WithTraceID(traceID string) *EventMetadata {
	m.TraceID = traceID
	return m
}

// WithUserID sets the user ID
func (m *EventMetadata) WithUserID(userID string) *EventMetadata {
	m.UserID = userID
	return m
}

// WithSessionID sets the session ID
func (m *EventMetadata) WithSessionID(sessionID string) *EventMetadata {
	m.SessionID = sessionID
	return m
}

// WithAttribute adds a custom attribute
func (m *EventMetadata) WithAttribute(key, value string) *EventMetadata {
	if m.Attributes == nil {
		m.Attributes = make(map[string]string)
	}
	m.Attributes[key] = value
	return m
}

// Event interface that all events must implement
type Event interface {
	GetMetadata() EventMetadata
	GetEventType() string
	ToJSON() ([]byte, error)
}

// EventEnvelope wraps events with routing information
type EventEnvelope struct {
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
	Metadata  EventMetadata   `json:"metadata"`
}

// NewEventEnvelope creates a new event envelope
func NewEventEnvelope(eventType string, event Event) (*EventEnvelope, error) {
	data, err := event.ToJSON()
	if err != nil {
		return nil, err
	}

	return &EventEnvelope{
		EventType: eventType,
		Data:      data,
		Metadata:  event.GetMetadata(),
	}, nil
}

// ToJSON serializes the envelope to JSON
func (e *EventEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes the envelope from JSON
func (e *EventEnvelope) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}