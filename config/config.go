package config

import (
	"log"
	"os"
)

type Config struct {
	DBURL    string
	RedisURL string
	Port     string
}

func Load() *Config {
	c := &Config{
		DBURL:    os.Getenv("DB_URL"),
		RedisURL: os.Getenv("REDIS_URL"),
		Port:     os.Getenv("PORT"),
	}

	if c.DBURL == "" || c.RedisURL == "" {
		log.Fatal("Missing required environment variables")
	}

	if c.Port == "" {
		c.Port = "8080"
	}

	return c
}
