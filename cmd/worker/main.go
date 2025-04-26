package main

import (
	"log"
	"net/http"

	"github.com/Miku7676/webhook-delivery-service/config"
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

	cfg := config.Load()

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opt)

	dburl := cfg.DBURL
	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting worker + janitor...")

	go helpers.StartWorker(redisClient)
	go helpers.StartLogClean(database)

	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		log.Println("Health server running at :9090/health")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Failed to start health server: %v", err)
		}
	}()

	select {}
}
