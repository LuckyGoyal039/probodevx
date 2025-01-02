package data

import "time"

type UserEvent struct {
	UserId    string    `json:"userId"`
	EventType string    `json:"eventType"`
	Timestamp time.Time `json:"timestamp"`
}
