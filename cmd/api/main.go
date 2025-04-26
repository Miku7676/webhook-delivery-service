// @title Webhook Delivery System - Gyanaranjan Bal
// @version 1.0
// @description API Documentation for Webhook Delivery Assignment - Developed by Gyanaranjan Bal
// @contact.name Biswajit Bal
// @host webhook-api-wwhi.onrender.com
// @BasePath /
package main

import (
	"log"

	"github.com/Miku7676/webhook-delivery-service/config"
	_ "github.com/Miku7676/webhook-delivery-service/docs"
	"github.com/Miku7676/webhook-delivery-service/handlers"
	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// var redisClient *redis.Client

// type HandlerDependencies struct {
// 	DB          *gorm.DB
// 	RedisClient *redis.Client
// }

func main() {

	gin.SetMode(gin.ReleaseMode)

	// Load env vars
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found.")
	}

	// Load centralized configuration
	cfg := config.Load()

	// Connect to Database
	dburl := cfg.DBURL
	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// AutoMigrate models
	if err := database.AutoMigrate(&models.Subscription{}, &models.WebhookTask{}, &models.DeliveryLog{}); err != nil {
		log.Fatalf("Failed to automigrate models: %v", err)
	}

	// Parse Redis URL properly (handles Upstash rediss://)
	redisOpt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(redisOpt)

	// Setup Handler Dependencies
	dependencyHandler := &handlers.HandlerDependencies{
		DB:          database,
		RedisClient: redisClient,
	}

	// Setup Gin Router
	r := gin.Default()

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all for now (can restrict later)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Hub-Signature-256"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // Swagger UI

	// Health Check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up"})
	})

	r.POST("/subscriptions", dependencyHandler.CreateSubscription())
	r.GET("/subscriptions/:id", dependencyHandler.GetSubscription())
	r.PUT("/subscriptions/:id", dependencyHandler.UpdateSubscription())
	r.DELETE("/subscriptions/:id", dependencyHandler.DeleteSubscription())

	r.POST("/ingest/:subscription_id", handlers.IngestWebhook(database))
	r.GET("/status/:webhook_id", handlers.GetDeliveryStatusByWebhook(database))
	r.GET("/subscriptions/:id/logs", handlers.GetRecentLogsBySubscription(database))

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running at http://localhost:%s\n", port)
	r.Run("0.0.0.0:" + port)
}
