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

type CreateSubscriptionRequest struct { // this struct is only to modify the request body of swagger.
	TargetURL string `json:"target_url" binding:"required"`
	Secret    string `json:"secret"`
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
	TargetURL      string    `json:"target_url"`
	AttemptNumber  int       `json:"attempt_number"`
	Status         string    `json:"status"`
	HTTPStatus     int       `json:"http_status"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}
