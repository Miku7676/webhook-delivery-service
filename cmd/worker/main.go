package main

import (
	"log"
	"net/http"

	"github.com/Miku7676/webhook-delivery-service/config"
	"github.com/Miku7676/webhook-delivery-service/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {


	// Load centralized configuration
	cfg := config.Load()

	// Parse Redis URL from config and create a Redis client
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opt)
	redisOpt := asynq.RedisClientOpt{
		Addr:     opt.Addr,
		Username: opt.Username,
		Password: opt.Password,
		TLSConfig: opt.TLSConfig,
		DB: opt.DB,
	}

	//connect to database
	dburl := cfg.DBURL
	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting worker + janitor...")

	// Start the background Worker to process webhook tasks
	go helpers.StartWorker(redisClient,redisOpt)

	// Start the background Log Cleaner
	go helpers.StartLogClean(database)

	// healthcheck endpoint
	go func() {
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		log.Println("Health server running at :9090/healthz")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Failed to start health server: %v", err)
		}
	}()

	// Block forever
	select {}
}
