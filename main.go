package main

import (
	"log"
	"os"

	"github.com/Miku7676/webhook-delivery-service/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
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
	database.AutoMigrate(&handlers.Subscription{}, &handlers.WebhookTask{})

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "up"})
	})

	r.POST("/subscriptions", handlers.CreateSubscription(database))
	r.GET("/subscriptions/:id", handlers.GetSubscription(database))
	r.PUT("/subscriptions/:id", handlers.UpdateSubscription(database))
	r.DELETE("/subscriptions/:id", handlers.DeleteSubscription(database))

	r.POST("/ingest/:subscription_id", handlers.IngestWebhook(database))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running at http://localhost:%s\n", port)
	r.Run(":" + port)
}
