package models

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeSMS      NotificationType = "sms"
	NotificationTypePush     NotificationType = "push"
	NotificationTypeWebhook  NotificationType = "webhook"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusRetrying  NotificationStatus = "retrying"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityLow      NotificationPriority = "low"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityUrgent   NotificationPriority = "urgent"
)

// NotificationTemplate represents a notification template
type NotificationTemplate struct {
	ID           string           `json:"id" db:"id"`
	Name         string           `json:"name" db:"name"`
	Type         NotificationType `json:"type" db:"type"`
	Subject      string           `json:"subject" db:"subject"`
	Body         string           `json:"body" db:"body"`
	Variables    []string         `json:"variables" db:"variables"`
	IsActive     bool             `json:"is_active" db:"is_active"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at" db:"updated_at"`
}

// Notification represents a notification in the system
type Notification struct {
	ID           string                `json:"id" db:"id"`
	UserID       string                `json:"user_id" db:"user_id"`
	Type         NotificationType      `json:"type" db:"type"`
	Status       NotificationStatus    `json:"status" db:"status"`
	Priority     NotificationPriority  `json:"priority" db:"priority"`
	Subject      string                `json:"subject" db:"subject"`
	Body         string                `json:"body" db:"body"`
	Recipient    string                `json:"recipient" db:"recipient"`
	TemplateID   string                `json:"template_id,omitempty" db:"template_id"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ScheduledAt  *time.Time            `json:"scheduled_at,omitempty" db:"scheduled_at"`
	SentAt       *time.Time            `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt  *time.Time            `json:"delivered_at,omitempty" db:"delivered_at"`
	FailedAt     *time.Time            `json:"failed_at,omitempty" db:"failed_at"`
	FailureReason string               `json:"failure_reason,omitempty" db:"failure_reason"`
	RetryCount   int                   `json:"retry_count" db:"retry_count"`
	MaxRetries   int                   `json:"max_retries" db:"max_retries"`
	CreatedAt    time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at" db:"updated_at"`
}

// UserPreferences represents a user's notification preferences
type UserPreferences struct {
	UserID              string `json:"user_id" db:"user_id"`
	EmailEnabled        bool   `json:"email_enabled" db:"email_enabled"`
	SMSEnabled          bool   `json:"sms_enabled" db:"sms_enabled"`
	PushEnabled         bool   `json:"push_enabled" db:"push_enabled"`
	MarketingEnabled    bool   `json:"marketing_enabled" db:"marketing_enabled"`
	OrderUpdatesEnabled bool   `json:"order_updates_enabled" db:"order_updates_enabled"`
	PhoneNumber         string `json:"phone_number,omitempty" db:"phone_number"`
	Timezone            string `json:"timezone" db:"timezone"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// SendNotificationRequest represents the request to send a notification
type SendNotificationRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Type        NotificationType       `json:"type" binding:"required"`
	Priority    NotificationPriority   `json:"priority"`
	Subject     string                 `json:"subject" binding:"required"`
	Body        string                 `json:"body" binding:"required"`
	Recipient   string                 `json:"recipient" binding:"required"`
	TemplateID  string                 `json:"template_id,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
}

// SendNotificationResponse represents the response after sending a notification
type SendNotificationResponse struct {
	ID          string               `json:"id"`
	Status      NotificationStatus   `json:"status"`
	ScheduledAt *time.Time           `json:"scheduled_at,omitempty"`
	SentAt      *time.Time           `json:"sent_at,omitempty"`
}

// WebhookNotification represents a webhook notification payload
type WebhookNotification struct {
	Event       string                 `json:"event"`
	Timestamp   time.Time              `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	Data        map[string]interface{} `json:"data"`
	Signature   string                 `json:"signature,omitempty"`
}

// SMSNotification represents an SMS notification
type SMSNotification struct {
	PhoneNumber string `json:"phone_number"`
	Message     string `json:"message"`
	CountryCode string `json:"country_code,omitempty"`
}

// PushNotification represents a push notification
type PushNotification struct {
	DeviceToken string                 `json:"device_token"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Sound       string                 `json:"sound,omitempty"`
	Badge       int                    `json:"badge,omitempty"`
}