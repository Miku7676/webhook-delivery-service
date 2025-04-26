// @title Webhook Delivery Service API
// @version 1.0
// @description API Documentation for Webhook Delivery System
// @host localhost:8080
// @BasePath /
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/Miku7676/webhook-delivery-service/docs"
	"github.com/Miku7676/webhook-delivery-service/handlers"
	"github.com/Miku7676/webhook-delivery-service/models"
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Run a goroutine to handle shutdown
	go func() {
		<-c
		fmt.Println("Shutting down...")
		os.Exit(0)
	}()

	// Load env vars
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found.")
	}

	dburl := os.Getenv("DB_URL")
	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	database.AutoMigrate(&models.Subscription{}, &models.WebhookTask{}, &models.DeliveryLog{})

	//redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Local Redis
	})

	dependencyHandler := &handlers.HandlerDependencies{
		DB:          database,
		RedisClient: redisClient,
	}

	// go helpers.StartWorker(redisClient) //go routine
	// go helpers.StartLogClean(database)

	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up"})
	})

	r.POST("/subscriptions", dependencyHandler.CreateSubscription())
	r.GET("/subscriptions/:id", dependencyHandler.GetSubscription())
	r.PUT("/subscriptions/:id", dependencyHandler.UpdateSubscription())
	r.DELETE("/subscriptions/:id", dependencyHandler.DeleteSubscription())

	r.POST("/ingest/:subscription_id", handlers.IngestWebhook(database))
	r.GET("/status/:webhook_id", handlers.GetDeliveryStatusByWebhook(database))
	r.GET("/subscriptions/:id/logs", handlers.GetRecentLogsBySubscription(database))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running at http://localhost:%s\n", port)
	r.Run(":" + port)
}
