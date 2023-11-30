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
	"github.com/shopspring/decimal"
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
var InitialMap = map[string]decimal.Decimal{
	"Logitech":  decimal.NewFromFloat(172.3),
	"Apple":     decimal.NewFromFloat(930.6),
	"Microsoft": decimal.NewFromFloat(859.5),
	"Samsung":   decimal.NewFromFloat(565.3),
	"Xerox":     decimal.NewFromFloat(415.7),
}

// nolint gomnd
// GeneratePrices is a method that generates prices for 5 companies 2 times per second and inserts them into a Redis stream
func (r *RedisRepository) GeneratePrices(ctx context.Context, initMap map[string]decimal.Decimal) error {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	currentPrices := make(map[string]decimal.Decimal, len(initMap))
	for company, price := range initMap {
		currentPrices[company] = price
	}
	for {
		for company := range initMap {
			change := decimal.NewFromFloat(rng.Float64() * 20.0).Sub(decimal.NewFromFloat(10.0))
			currentPrices[company] = currentPrices[company].Add(change)
			if currentPrices[company].LessThan(decimal.NewFromFloat(0)) {
				currentPrices[company] = decimal.NewFromFloat(0.1)
			}
			r.stockData.Store(company, currentPrices[company])
			_, err := r.client.XAdd(ctx, &redis.XAddArgs{
				Stream: "shares",
				Values: map[string]interface{}{
					"message": fmt.Sprintf("%s: %.2f", company, currentPrices[company].InexactFloat64()),
				},
				MaxLen: 5,
			}).Result()
			if err != nil {
				log.Fatalf("error when writing a message to Redis Stream: %v", err)
			}
		}
		time.Sleep(time.Second / 2)
	}
}
