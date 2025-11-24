package factory

import (
	"github.com/smykla-labs/klaudiush/internal/validator"
	notificationvalidators "github.com/smykla-labs/klaudiush/internal/validators/notification"
	"github.com/smykla-labs/klaudiush/pkg/config"
	"github.com/smykla-labs/klaudiush/pkg/hook"
	"github.com/smykla-labs/klaudiush/pkg/logger"
)

// NotificationValidatorFactory creates notification validators from configuration.
type NotificationValidatorFactory struct {
	log logger.Logger
}

// NewNotificationValidatorFactory creates a new NotificationValidatorFactory.
func NewNotificationValidatorFactory(log logger.Logger) *NotificationValidatorFactory {
	return &NotificationValidatorFactory{log: log}
}

// CreateValidators creates all notification validators based on configuration.
func (f *NotificationValidatorFactory) CreateValidators(
	cfg *config.Config,
) []ValidatorWithPredicate {
	var validators []ValidatorWithPredicate

	if cfg.Validators.Notification.Bell != nil && cfg.Validators.Notification.Bell.IsEnabled() {
		validators = append(validators, f.createBellValidator())
	}

	return validators
}

func (f *NotificationValidatorFactory) createBellValidator() ValidatorWithPredicate {
	return ValidatorWithPredicate{
		Validator: notificationvalidators.NewBellValidator(f.log),
		Predicate: validator.EventTypeIs(hook.Notification),
	}
}
