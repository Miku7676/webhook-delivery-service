package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TargetURL string    `json:"target_url"`
	Secret    string    `json:"secret"`
}

type WebhookTask struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SubscriptionID uuid.UUID `gorm:"type:uuid" json:"subscription_id"`
	Payload        string    `json:"payload"`
	CreatedAt      time.Time `json:"created_at"`
}

type DeliveryLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	WebhookTaskID  uuid.UUID `gorm:"type:uuid"`
	SubscriptionID uuid.UUID `gorm:"type:uuid"`
	TargetURL      string
	AttemptNumber  int
	Status         string // "Success" or "Failed"
	HTTPStatus     int
	ErrorMessage   string
	CreatedAt      time.Time
}
