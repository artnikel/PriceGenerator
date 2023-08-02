// Package config with environment variables
package config

// Variables is a struct with environment variables
type Variables struct {
	RedisPriceAddress  string `env:"REDIS_PRICE_ADDRESS"`
	RedisPricePassword string `env:"REDIS_PRICE_PASSWORD"`
}
