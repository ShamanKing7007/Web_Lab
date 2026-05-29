package queue

import (
	"time"

	"github.com/google/uuid"
)

const UserRegisteredRoutingKey = "user.registered"

type UserRegisteredEvent struct {
	EventID   string                  `json:"eventId"`
	EventType string                  `json:"eventType"`
	Timestamp time.Time               `json:"timestamp"`
	Payload   UserRegisteredPayload   `json:"payload"`
	Metadata  UserRegisteredEventMeta `json:"metadata"`
}

type UserRegisteredPayload struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
}

type UserRegisteredEventMeta struct {
	Attempt       int    `json:"attempt"`
	SourceService string `json:"sourceService"`
}

func NewUserRegisteredEvent(userID uuid.UUID, email string) UserRegisteredEvent {
	return UserRegisteredEvent{
		EventID:   uuid.NewString(),
		EventType: "user.registered",
		Timestamp: time.Now().UTC(),
		Payload: UserRegisteredPayload{
			UserID: userID.String(),
			Email:  email,
		},
		Metadata: UserRegisteredEventMeta{
			Attempt:       1,
			SourceService: "web-labs-api",
		},
	}
}
