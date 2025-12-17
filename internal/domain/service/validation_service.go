package service

import (
	"fmt"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// ValidationService provides currency code validation utilities.
// This is a domain service that encapsulates validation logic.
type ValidationService struct{}

// NewValidationService creates a new ValidationService.
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidateCurrencyCode validates a currency code string and returns a CurrencyCode.
// This is a convenience method that wraps entity.NewCurrencyCode.
func (s *ValidationService) ValidateCurrencyCode(code string) (entity.CurrencyCode, error) {
	return entity.NewCurrencyCode(code)
}

// ValidateCurrencyPair validates both base and target currency codes.
// Returns an error if either code is invalid or if they are the same.
func (s *ValidationService) ValidateCurrencyPair(baseCode, targetCode string) (base, target entity.CurrencyCode, err error) {
	base, err = entity.NewCurrencyCode(baseCode)
	if err != nil {
		return "", "", fmt.Errorf("invalid base currency: %w", err)
	}

	target, err = entity.NewCurrencyCode(targetCode)
	if err != nil {
		return "", "", fmt.Errorf("invalid target currency: %w", err)
	}

	if base.Equal(target) {
		return "", "", entity.ErrCurrencyCodeMismatch
	}

	return base, target, nil
}

// IsValidCurrencyCode checks if a string is a valid currency code.
func (s *ValidationService) IsValidCurrencyCode(code string) bool {
	_, err := entity.NewCurrencyCode(code)
	return err == nil
}


