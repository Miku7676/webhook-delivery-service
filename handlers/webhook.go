package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// IngestWebhook godoc
// @Summary Ingest a webhook
// @Description Accepts a webhook payload and queues it for delivery. If the subscription has a secret, you must provide the correct X-Hub-Signature-256 header (HMAC-SHA256 of the payload using the secret).
// @Tags Webhook
// @Accept json
// @Produce json
// @Param subscription_id path string true "Subscription ID"
// @Param X-Hub-Signature-256 header string false "HMAC SHA256 signature of payload using secret"
// @Param payload body object true "Webhook payload JSON"
// @Success 202 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ingest/{subscription_id} [post]
func IngestWebhook(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Load Redis URL from environment
		redisUrl := os.Getenv("REDIS_URL")
		if redisUrl == "" {
			log.Println("REDIS_URL not set")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Parse Redis URL properly
		redisOpt, err := redis.ParseURL(redisUrl)
		if err != nil {
			log.Printf("Failed to parse Redis URL: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Subscription ID
		subID := c.Param("subscription_id")
		parsedID, err := uuid.Parse(subID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		// Parse Payload
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		// Marshal payload
		body, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Payload marshal failed"})
			return
		}

		// Fetch subscription from DB
		var sub models.Subscription
		if err := db.First(&sub, "id = ?", parsedID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}

		// Verify Signature if secret is present
		if sub.Secret != "" {
			expectedSignature := computeHMACSHA256(body, sub.Secret)
			receivedSignature := c.GetHeader("X-Hub-Signature-256")

			if receivedSignature == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing X-Hub-Signature-256 header"})
				return
			}

			if !hmac.Equal([]byte(receivedSignature), []byte(expectedSignature)) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
				return
			}
		}

		// Create Task
		task := models.WebhookTask{
			ID:             uuid.New(),
			SubscriptionID: parsedID,
			Payload:        string(body),
			CreatedAt:      time.Now(),
		}
		if err := db.Create(&task).Error; err != nil {
			log.Printf("Failed to create Webhook Task: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Create Asynq client
		client := asynq.NewClient(asynq.RedisClientOpt{
			Addr:      redisOpt.Addr,
			Username:  redisOpt.Username,
			Password:  redisOpt.Password,
			TLSConfig: redisOpt.TLSConfig, // if present, will be included
			DB:        redisOpt.DB,
		})
		defer client.Close()

		// Enqueue Job
		jobPayload, _ := json.Marshal(task)
		job := asynq.NewTask("webhook:deliver", jobPayload)

		_, err = client.Enqueue(job,
			asynq.Queue("default"),
			asynq.MaxRetry(5),
			asynq.Timeout(10*time.Second),
		)
		if err != nil {
			log.Printf("Failed to enqueue task: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue task"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"status": "queued", "task_id": task.ID})
	}
}

func computeHMACSHA256(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
