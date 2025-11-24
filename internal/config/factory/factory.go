// Package factory provides factories for creating validators from configuration.
package factory

import (
	"github.com/smykla-labs/klaudiush/internal/validator"
	"github.com/smykla-labs/klaudiush/pkg/config"
	"github.com/smykla-labs/klaudiush/pkg/logger"
)

// ValidatorWithPredicate pairs a validator with its registration predicate.
type ValidatorWithPredicate struct {
	Validator validator.Validator
	Predicate validator.Predicate
}

// ValidatorFactory creates validators from configuration.
type ValidatorFactory interface {
	// CreateGitValidators creates all git validators from config.
	CreateGitValidators(cfg *config.Config) []ValidatorWithPredicate

	// CreateFileValidators creates all file validators from config.
	CreateFileValidators(cfg *config.Config) []ValidatorWithPredicate

	// CreateNotificationValidators creates all notification validators from config.
	CreateNotificationValidators(cfg *config.Config) []ValidatorWithPredicate

	// CreateAll creates all validators from config.
	CreateAll(cfg *config.Config) []ValidatorWithPredicate
}

// DefaultValidatorFactory is the default implementation of ValidatorFactory.
type DefaultValidatorFactory struct {
	gitFactory          *GitValidatorFactory
	fileFactory         *FileValidatorFactory
	notificationFactory *NotificationValidatorFactory
}

// NewValidatorFactory creates a new DefaultValidatorFactory.
func NewValidatorFactory(log logger.Logger) *DefaultValidatorFactory {
	return &DefaultValidatorFactory{
		gitFactory:          NewGitValidatorFactory(log),
		fileFactory:         NewFileValidatorFactory(log),
		notificationFactory: NewNotificationValidatorFactory(log),
	}
}

// CreateGitValidators creates all git validators from config.
func (f *DefaultValidatorFactory) CreateGitValidators(cfg *config.Config) []ValidatorWithPredicate {
	return f.gitFactory.CreateValidators(cfg)
}

// CreateFileValidators creates all file validators from config.
func (f *DefaultValidatorFactory) CreateFileValidators(
	cfg *config.Config,
) []ValidatorWithPredicate {
	return f.fileFactory.CreateValidators(cfg)
}

// CreateNotificationValidators creates all notification validators from config.
func (f *DefaultValidatorFactory) CreateNotificationValidators(
	cfg *config.Config,
) []ValidatorWithPredicate {
	return f.notificationFactory.CreateValidators(cfg)
}

// CreateAll creates all validators from config.
func (f *DefaultValidatorFactory) CreateAll(cfg *config.Config) []ValidatorWithPredicate {
	var all []ValidatorWithPredicate

	all = append(all, f.CreateGitValidators(cfg)...)
	all = append(all, f.CreateFileValidators(cfg)...)
	all = append(all, f.CreateNotificationValidators(cfg)...)

	return all
}
