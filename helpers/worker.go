package helpers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type WorkerDependencies struct {
	RedisClient *redis.Client
}

func StartWorker(rdb *redis.Client) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("REDIS_URL environment variable not set")
	}

	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	// Configure Asynq for redis
	redisOpt := asynq.RedisClientOpt{
		Addr:     opt.Addr,
		Username: opt.Username,
		Password: opt.Password,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		DB: opt.DB,
	}

	// Create Worker dependency instance
	deps := &WorkerDependencies{
		RedisClient: rdb,
	}

	// Setup Asynq server
	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 5, // 5 concurrent workers
			Queues: map[string]int{
				"default": 10, // Queue configuration
			},
			RetryDelayFunc: asynq.DefaultRetryDelayFunc, // Exponential backoff for retries - builtin function
		},
	)

	// setup taskhandler multiplexer
	mux := asynq.NewServeMux()
	mux.HandleFunc("webhook:deliver", deps.processWebhookTask)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run worker server: %v", err)
	}
}

func (wd *WorkerDependencies) processWebhookTask(ctx context.Context, task *asynq.Task) error {
	var webhookTask models.WebhookTask
	if err := json.Unmarshal(task.Payload(), &webhookTask); err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return err
	}

	dburl := os.Getenv("DB_URL")
	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}

	var sub models.Subscription
	cacheKey := fmt.Sprintf("subscription:%s", webhookTask.SubscriptionID)

	// fetch subscription from redis
	val, err := wd.RedisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
		// Cache miss - fetch from DB
		if err := database.First(&sub, "id = ?", webhookTask.SubscriptionID).Error; err != nil {
			log.Printf("Subscription not found: %v", err)
			return err
		}

		// Cache subscription
		subBytes, _ := json.Marshal(sub)
		wd.RedisClient.Set(ctx, cacheKey, subBytes, time.Hour)
	} else if err != nil {
		//cache error
		log.Printf("Failed to get subscription from cache: %v", err)
		return err
	} else {
		// Cache hit
		json.Unmarshal([]byte(val), &sub)
	}

	// Prepare HTTP client and request
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", sub.TargetURL, bytes.NewBuffer([]byte(webhookTask.Payload)))
	if err != nil {
		log.Printf("Request creation failed: %v", err)
		retryCount, _ := asynq.GetRetryCount(ctx)
		logAttempt(database, webhookTask, sub, retryCount+1, "Failed", 500, err.Error()) // server error, failed to create request
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request to subscription url
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Delivery failed: %v", err)
		retryCount, _ := asynq.GetRetryCount(ctx)
		logAttempt(database, webhookTask, sub, retryCount+1, "Failed", 404, err.Error()) // assumed that the url does not exist
		return err
	}
	defer resp.Body.Close()

	// Handle HTTP response
	status := "Success"
	errorMessage := ""
	if resp.StatusCode >= 300 {
		status = "Failed"

		// Read body (optional, but useful for debugging why it failed)
		bodyBytes, _ := io.ReadAll(resp.Body)
		errorMessage = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))

		log.Printf("Received non-2xx response: %s", errorMessage)
	}

	retryCount, _ := asynq.GetRetryCount(ctx)
	logAttempt(database, webhookTask, sub, retryCount+1, status, resp.StatusCode, errorMessage)
	if status == "Failed" {
		return fmt.Errorf("Non-2xx status code: %d", resp.StatusCode)
	}

	log.Printf("Delivery successful for %s", webhookTask.ID)
	return nil
}

// logAttempt records a delivery attempt into the DeliveryLog table
func logAttempt(db *gorm.DB, task models.WebhookTask, sub models.Subscription, attempt int, status string, httpStatus int, errMsg string) {
	logEntry := models.DeliveryLog{
		ID:             uuid.New(),
		WebhookTaskID:  task.ID,
		SubscriptionID: task.SubscriptionID,
		TargetURL:      sub.TargetURL,
		AttemptNumber:  attempt,
		Status:         status,
		HTTPStatus:     httpStatus,
		ErrorMessage:   errMsg,
		CreatedAt:      time.Now(),
	}
	if err := db.Create(&logEntry).Error; err != nil {
		log.Printf("Failed to log delivery attempt: %v", err)
	}
}
