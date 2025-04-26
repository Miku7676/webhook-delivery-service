package main

import (
	"log"
	"os"

	"github.com/Miku7676/webhook-delivery-service/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found.")
	}

	redisUrl := os.Getenv("REDIS_URL")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	dburl := os.Getenv("DB_URL")
	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting worker + janitor...")

	go helpers.StartWorker(redisClient)
	go helpers.StartLogClean(database)

	select {}
}
