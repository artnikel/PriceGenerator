// Package repository is a lower level of project
package repository

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
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

// nolint gochecknoglobals
// InitialMap is a map with name of company and start price of actions
var InitialMap = map[string]decimal.Decimal{
	"Logitech":  initRandomPrice(),
	"Apple":     initRandomPrice(),
	"Microsoft": initRandomPrice(),
	"Samsung":   initRandomPrice(),
	"Xerox":     initRandomPrice(),
}

func initRandomPrice() decimal.Decimal {
	min, max := 100.0, 1000.0
	// nolint gosec
	return decimal.NewFromFloat(min + rand.Float64()*(max-min))
}

// nolint gomnd
// GeneratePrices is a method that generates prices for 5 companies 2 times per second and put them into a Redis stream
func (r *RedisRepository) GeneratePrices(ctx context.Context, initMap map[string]decimal.Decimal) error {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))
	currentPrices := make(map[string]decimal.Decimal, len(initMap))
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	msgs, err := r.client.XRead(ctxTimeout, &redis.XReadArgs{
		Streams: []string{"shares", "0"},
		Count:   10,
		Block:   0,
	}).Result()
	if err == nil && len(msgs) > 0 && len(msgs[0].Messages) > 0 {
		for _, message := range msgs[0].Messages {
			if rawMessage, ok := message.Values["message"].(string); ok {
				parts := strings.Split(rawMessage, ":")
				if len(parts) != 2 {
					return fmt.Errorf("incorrect message format: %s", rawMessage)
				} else {
					company := strings.TrimSpace(parts[0])
					priceStr := strings.TrimSpace(parts[1])
					price, err := decimal.NewFromString(priceStr)
					if err != nil {
						return fmt.Errorf("error when converting price to number: %v", err)
					} else {
						currentPrices[company] = price
					}
				}
			} else {
				return fmt.Errorf("missing 'message' field in Redis stream")
			}
		}
	} else {
		for company, price := range initMap {
			currentPrices[company] = price
		}
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
				return fmt.Errorf("error when writing a message to Redis Stream: %v", err)
			}
		}
		time.Sleep(time.Second / 2)
	}
}
