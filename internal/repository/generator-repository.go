// Package repository is a lower level of project
package repository

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisRepository contains objects of type *redis.Client and *sync.Map
type RedisRepository struct {
	client    *redis.Client
	stockData *sync.Map
}

// NewRedisRepository accepts an object of *redis.Client and returns an object of type *Redis
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client:    client,
		stockData: &sync.Map{},
	}
}

// nolint gomnd
// InitialMap is a map with name of company and start price of actions
var InitialMap = map[string]float64{
	"Logitech":  172.3,
	"Apple":     930.6,
	"Microsoft": 859.5,
	"Samsung":   565.3,
	"Xerox":     415.7,
}

// nolint gomnd
// GeneratePrices is a method that generating price of 5 company`s 2 times per second and insert them in redis stream
func (r *RedisRepository) GeneratePrices(ctx context.Context, initMap map[string]float64) error {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	for company, price := range initMap {
		r.stockData.Store(company, price)
	}

	r.stockData.Range(func(key, value interface{}) bool {
		company := key.(string)
		price := value.(float64)

		change := rng.Float64()*40.0 - 20.0
		price += change

		if price < 0 {
			price = 0.1
		}

		r.stockData.Store(company, price)

		_, err := r.client.XAdd(ctx, &redis.XAddArgs{
			Stream: "messagestream",
			Values: map[string]interface{}{
				"message": fmt.Sprintf("%s: %.2f", company, price),
			},
			MaxLen: 5,
		}).Result()
		if err != nil {
			log.Fatalf("Error when writing a message to Redis Stream: %v", err)
		}

		return true
	})
	time.Sleep(time.Second / 2)
	return nil
}
