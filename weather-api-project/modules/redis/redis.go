package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func StartRedisServer() error {
	// Redis server
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

	ctx := context.Background()

	setErr := redisClient.Set(ctx, "testing", "true", 0).Err()

	if setErr != nil {
		return fmt.Errorf("unable to save data in Redis %v", setErr)
	}

	result, getErr := redisClient.Get(ctx, "testing").Result()

	if getErr != nil {
		return fmt.Errorf("unable to get data from Redis %v", getErr)
	}

	log.Print("redis test -> " + result)

	return nil
}

func GetRedisClient() (*redis.Client, error) {

	if redisClient == nil {
		return nil, fmt.Errorf("redisClient is invalid")
	}

	return redisClient, nil
}
