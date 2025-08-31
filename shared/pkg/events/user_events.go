package events

import (
	"encoding/json"
	"time"
)

// UserCreated represents the event published when a user is created
type UserCreated struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	EventID   string    `json:"event_id"` // For idempotency
}

// ToJSON serializes the event to JSON
func (u *UserCreated) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

// FromJSON deserializes the event from JSON
func (u *UserCreated) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}