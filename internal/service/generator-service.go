// Package service contains business logic of a project
package service

import (
	"context"
	"fmt"
)

// GeneratorRepository is interface with method for generating prices
type GeneratorRepository interface {
	GeneratePrices(ctx context.Context, initMap map[string]float64) error
}

// GeneratorService contains GeneratorRepository interface
type GeneratorService struct {
	genRep GeneratorRepository
}

// NewGeneratorService accepts GeneratorRepository object and returnes an object of type *GeneratorService
func NewGeneratorService(genRep GeneratorRepository) *GeneratorService {
	return &GeneratorService{genRep: genRep}
}

// GeneratePrices is a method of GeneratorService that calls method of Repository
func (p *GeneratorService) GeneratePrices(ctx context.Context, initMap map[string]float64) error {
	err := p.genRep.GeneratePrices(ctx, initMap)
	if err != nil {
		return fmt.Errorf("GeneratorService-GeneratePrices: error: %w", err)
	}
	return nil
}
