package events

import "github.com/PlayEconomy37/Play.Common/permissions"

// UserUpdatedEvent is the event sent whenever an user is created or updated
type UserUpdatedEvent struct {
	ID          int64                   `json:"id"`
	Email       string                  `json:"email"`
	Permissions permissions.Permissions `json:"permissions"`
	Activated   bool                    `json:"activated"`
	Version     int32                   `json:"version"`
}
