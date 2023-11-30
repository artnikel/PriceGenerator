// Package main of a project
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/artnikel/PriceGenerator/internal/config"
	"github.com/artnikel/PriceGenerator/internal/repository"
	"github.com/go-redis/redis/v8"
)

func connectRedis() (*redis.Client, error) {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("could not parse config: ", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisPriceAddress,
		DB:   0,
	})
	_, err = client.Ping(client.Context()).Result()
	if err != nil {
		return nil, fmt.Errorf("error in method client.Ping(): %v", err)
	}
	return client, nil
}

// nolint gocritic
func main() {
	ctx := context.Background()
	redisClient, err := connectRedis()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer func() {
		errClose := redisClient.Close()
		if errClose != nil {
			log.Fatalf("failed to disconnect from Redis: %v", errClose)
		}
	}()
	repoRedis := repository.NewRedisRepository(redisClient)
	fmt.Println("Generating started")
	for {
		err := repoRedis.GeneratePrices(ctx, repository.InitialMap)
		if err != nil {
			log.Fatalf("failed to run method GeneratePrices: %v", err)
		}
	}
}
