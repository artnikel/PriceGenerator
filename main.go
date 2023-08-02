// Package main of a project
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/artnikel/PriceGenerator/internal/config"
	"github.com/artnikel/PriceGenerator/internal/repository"
	"github.com/artnikel/PriceGenerator/internal/service"
	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v8"
)

func connectRedis() (*redis.Client, error) {
	cfg := config.Variables{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisPriceAddress,
		Password: cfg.RedisPricePassword,
		DB:       0,
	})
	_, err := client.Ping(client.Context()).Result()
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
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer func() {
		errClose := redisClient.Close()
		if errClose != nil {
			log.Fatalf("Failed to disconnect from Redis: %v", errClose)
		}
	}()
	repoRedis := repository.NewRedisRepository(redisClient)
	servRedis := service.NewGeneratorService(repoRedis)
	fmt.Println("Generating started")
	for {
		err := servRedis.GeneratePrices(ctx,repository.InitialMap)
		if err != nil {
			log.Fatalf("Failed to run method GeneratePrices: %v", err)
		}
	}
}
