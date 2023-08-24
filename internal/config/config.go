// Package config with environment variables
package config

import "github.com/caarlos0/env"

// Variables is a struct with environment variables
type Variables struct {
	RedisPriceAddress  string `env:"REDIS_PRICE_ADDRESS"`
	RedisPricePassword string `env:"REDIS_PRICE_PASSWORD"`
}

// New returns parsed object of config
func New() (*Variables, error) {
	cfg := &Variables{}
	err := env.Parse(cfg)
	return cfg, err
}
