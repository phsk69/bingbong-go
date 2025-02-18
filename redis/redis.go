package redis

import (
	"fmt"
	"log"
	"os"
	"time"

	"git.ssy.dk/noob/bingbong-go/handlers"
)

func InitRedis() (*handlers.DistributedHub, error) {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisUser := os.Getenv("REDIS_USER")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Validate if we have all the required environment variables
	if redisHost == "" || redisPort == "" || redisPassword == "" || redisUser == "" {
		log.Println("missing required environment variables for Redis")
		return nil, fmt.Errorf("missing required environment variables for Redis")
	}

	redisConfig := handlers.HubConfig{
		RedisURL:        fmt.Sprintf("redis://%s:%s@%s:%s/0", redisUser, redisPassword, redisHost, redisPort),
		MaxRetries:      3,
		SessionDuration: 24 * time.Hour,
		BufferSize:      256,
	}

	hub, err := handlers.NewDistributedHub(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis hub: %v", err)
	}

	go hub.Run()
	log.Printf("Redis hub initialized and connected to %s:%s", redisHost, redisPort)
	return hub, nil
}
