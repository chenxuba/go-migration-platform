package model

import "time"

type MQEventLog struct {
	ID        int64     `json:"id"`
	Topic     string    `json:"topic"`
	Tag       string    `json:"tag,omitempty"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"createdAt"`
}
