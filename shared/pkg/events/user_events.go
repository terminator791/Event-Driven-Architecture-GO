package events

import (
	"encoding/json"
	"time"
)

// UserCreated represents the event published when a user is created
type UserCreated struct {
	BaseEvent
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(userID, email string, createdAt time.Time) *UserCreated {
	return &UserCreated{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("user-api", "1.0"),
		},
		UserID:    userID,
		Email:     email,
		CreatedAt: createdAt,
	}
}

// GetMetadata returns the event metadata
func (u *UserCreated) GetMetadata() EventMetadata {
	return u.Metadata
}

// GetEventType returns the event type
func (u *UserCreated) GetEventType() string {
	return "user.created"
}

// ToJSON serializes the event to JSON
func (u *UserCreated) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// FromJSON deserializes the event from JSON
func (u *UserCreated) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

// UserUpdated represents the event published when a user is updated
type UserUpdated struct {
	BaseEvent
	UserID      string            `json:"user_id"`
	Email       string            `json:"email"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Changes     map[string]string `json:"changes"`
	PreviousEmail string          `json:"previous_email,omitempty"`
}

// NewUserUpdated creates a new UserUpdated event
func NewUserUpdated(userID, email string, updatedAt time.Time, changes map[string]string) *UserUpdated {
	return &UserUpdated{
		BaseEvent: BaseEvent{
			Metadata: NewEventMetadata("user-api", "1.0"),
		},
		UserID:    userID,
		Email:     email,
		UpdatedAt: updatedAt,
		Changes:   changes,
	}
}

// GetMetadata returns the event metadata
func (u *UserUpdated) GetMetadata() EventMetadata {
	return u.Metadata
}

// GetEventType returns the event type
func (u *UserUpdated) GetEventType() string {
	return "user.updated"
}

// ToJSON serializes the event to JSON
func (u *UserUpdated) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// FromJSON deserializes the event from JSON
func (u *UserUpdated) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}