package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// type WebhookTask struct {
// 	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
// 	SubscriptionID uuid.UUID `gorm:"type:uuid" json:"subscription_id"`
// 	Payload        string    `json:"payload"`
// 	CreatedAt      time.Time `json:"created_at"`
// }

func IngestWebhook(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		subID := c.Param("subscription_id")
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		parsedID, err := uuid.Parse(subID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		body, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payload marshal failed"})
			return
		}

		task := models.WebhookTask{
			ID:             uuid.New(),
			SubscriptionID: parsedID,
			Payload:        string(body),
			CreatedAt:      time.Now(),
		}
		db.Create(&task)

		client := asynq.NewClient(asynq.RedisClientOpt{Addr: "localhost:6379"}) //change to the docker container name
		defer client.Close()

		jobPayload, _ := json.Marshal(task)
		job := asynq.NewTask("webhook:deliver", jobPayload)

		_, err = client.Enqueue(job,
			asynq.Queue("default"),
			asynq.MaxRetry(5),
			asynq.Timeout(10*time.Second),
		) //asynq.Retention(24*time.Hour),
		if err != nil {
			log.Printf("Failed to enqueue task: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"status": "queued", "task_id": task.ID})
	}
}
