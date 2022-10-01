package events

// UserUpdatedEvent is the event sent whenever an user is created or updated
type UserUpdatedEvent struct {
	ID          int64    `json:"id"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	Activated   bool     `json:"activated"`
	Version     int32    `json:"version"`
}
