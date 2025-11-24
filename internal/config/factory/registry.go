package factory

import (
	"github.com/smykla-labs/klaudiush/internal/validator"
	"github.com/smykla-labs/klaudiush/pkg/config"
	"github.com/smykla-labs/klaudiush/pkg/logger"
)

// RegistryBuilder builds a validator registry from configuration.
type RegistryBuilder struct {
	factory ValidatorFactory
	log     logger.Logger
}

// NewRegistryBuilder creates a new RegistryBuilder.
func NewRegistryBuilder(log logger.Logger) *RegistryBuilder {
	return &RegistryBuilder{
		factory: NewValidatorFactory(log),
		log:     log,
	}
}

// Build creates a validator registry from the provided configuration.
// It creates all enabled validators and registers them with their predicates.
func (b *RegistryBuilder) Build(cfg *config.Config) *validator.Registry {
	registry := validator.NewRegistry()

	// Get all validators with predicates from factory
	validatorsWithPredicates := b.factory.CreateAll(cfg)

	// Register each validator with its predicate
	for _, vp := range validatorsWithPredicates {
		registry.Register(vp.Validator, vp.Predicate)
	}

	b.log.Debug("registry built",
		"validator_count", len(validatorsWithPredicates),
	)

	return registry
}
