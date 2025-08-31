package events

import (
	"encoding/json"
	"time"

	"github.com/terminator791/Event-Driven-Architecture-GO/shared/pkg/models"
)

// NotificationRequested represents the event published when a notification is requested
type NotificationRequested struct {
	BaseEvent
	NotificationID string                         `json:"notification_id"`
	UserID         string                         `json:"user_id"`
	Type           models.NotificationType        `json:"type"`
	Priority       models.NotificationPriority    `json:"priority"`
	Subject        string                         `json:"subject"`
	Body           string                         `json:"body"`
	Recipient      string                         `json:"recipient"`
	TemplateID     string                         `json:"template_id,omitempty"`
	Variables      map[string]interface{}         `json:"variables,omitempty"`
	ScheduledAt    *time.Time                     `json:"scheduled_at,omitempty"`
	RequestedAt    time.Time                      `json:"requested_at"`
}

// NewNotificationRequested creates a new NotificationRequested event
func NewNotificationRequested(notification *models.Notification) *NotificationRequested {
	return &NotificationRequested{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("notification-service", "1.0"),
		},
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Priority:       notification.Priority,
		Subject:        notification.Subject,
		Body:           notification.Body,
		Recipient:      notification.Recipient,
		TemplateID:     notification.TemplateID,
		Variables:      notification.Metadata,
		ScheduledAt:    notification.ScheduledAt,
		RequestedAt:    notification.CreatedAt,
	}
}

// GetMetadata returns the event metadata
func (n *NotificationRequested) GetMetadata() EventMetadata {
	return n.Metadata
}

// GetEventType returns the event type
func (n *NotificationRequested) GetEventType() string {
	return "notification.requested"
}

// ToJSON serializes the event to JSON
func (n *NotificationRequested) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON deserializes the event from JSON
func (n *NotificationRequested) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}

// NotificationSent represents the event published when a notification is sent
type NotificationSent struct {
	BaseEvent
	NotificationID string                      `json:"notification_id"`
	UserID         string                      `json:"user_id"`
	Type           models.NotificationType     `json:"type"`
	Recipient      string                      `json:"recipient"`
	Subject        string                      `json:"subject"`
	SentAt         time.Time                   `json:"sent_at"`
	Provider       string                      `json:"provider,omitempty"`
	ProviderID     string                      `json:"provider_id,omitempty"`
}

// NewNotificationSent creates a new NotificationSent event
func NewNotificationSent(notification *models.Notification, provider, providerID string) *NotificationSent {
	return &NotificationSent{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("notification-service", "1.0"),
		},
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Recipient:      notification.Recipient,
		Subject:        notification.Subject,
		SentAt:         *notification.SentAt,
		Provider:       provider,
		ProviderID:     providerID,
	}
}

// GetMetadata returns the event metadata
func (n *NotificationSent) GetMetadata() EventMetadata {
	return n.Metadata
}

// GetEventType returns the event type
func (n *NotificationSent) GetEventType() string {
	return "notification.sent"
}

// ToJSON serializes the event to JSON
func (n *NotificationSent) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON deserializes the event from JSON
func (n *NotificationSent) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}

// NotificationDelivered represents the event published when a notification is delivered
type NotificationDelivered struct {
	BaseEvent
	NotificationID string                  `json:"notification_id"`
	UserID         string                  `json:"user_id"`
	Type           models.NotificationType `json:"type"`
	Recipient      string                  `json:"recipient"`
	DeliveredAt    time.Time               `json:"delivered_at"`
	Provider       string                  `json:"provider,omitempty"`
	ProviderID     string                  `json:"provider_id,omitempty"`
}

// NewNotificationDelivered creates a new NotificationDelivered event
func NewNotificationDelivered(notification *models.Notification, provider, providerID string) *NotificationDelivered {
	return &NotificationDelivered{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("notification-service", "1.0"),
		},
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Recipient:      notification.Recipient,
		DeliveredAt:    *notification.DeliveredAt,
		Provider:       provider,
		ProviderID:     providerID,
	}
}

// GetMetadata returns the event metadata
func (n *NotificationDelivered) GetMetadata() EventMetadata {
	return n.Metadata
}

// GetEventType returns the event type
func (n *NotificationDelivered) GetEventType() string {
	return "notification.delivered"
}

// ToJSON serializes the event to JSON
func (n *NotificationDelivered) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON deserializes the event from JSON
func (n *NotificationDelivered) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}

// NotificationFailed represents the event published when a notification fails
type NotificationFailed struct {
	BaseEvent
	NotificationID string                  `json:"notification_id"`
	UserID         string                  `json:"user_id"`
	Type           models.NotificationType `json:"type"`
	Recipient      string                  `json:"recipient"`
	FailureReason  string                  `json:"failure_reason"`
	FailedAt       time.Time               `json:"failed_at"`
	RetryCount     int                     `json:"retry_count"`
	MaxRetries     int                     `json:"max_retries"`
	WillRetry      bool                    `json:"will_retry"`
}

// NewNotificationFailed creates a new NotificationFailed event
func NewNotificationFailed(notification *models.Notification, willRetry bool) *NotificationFailed {
	return &NotificationFailed{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("notification-service", "1.0"),
		},
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Recipient:      notification.Recipient,
		FailureReason:  notification.FailureReason,
		FailedAt:       *notification.FailedAt,
		RetryCount:     notification.RetryCount,
		MaxRetries:     notification.MaxRetries,
		WillRetry:      willRetry,
	}
}

// GetMetadata returns the event metadata
func (n *NotificationFailed) GetMetadata() EventMetadata {
	return n.Metadata
}

// GetEventType returns the event type
func (n *NotificationFailed) GetEventType() string {
	return "notification.failed"
}

// ToJSON serializes the event to JSON
func (n *NotificationFailed) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON deserializes the event from JSON
func (n *NotificationFailed) FromJSON(data []byte) error {
	return json.Unmarshal(data, n)
}