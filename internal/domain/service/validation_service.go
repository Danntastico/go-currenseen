package service

import (
	"fmt"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// ValidationService provides currency code validation utilities.
// This is a domain service that encapsulates validation logic for currency pairs.
//
// Note: For single currency code validation, use entity.NewCurrencyCode() directly.
type ValidationService struct{}

// NewValidationService creates a new ValidationService.
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidateCurrencyPair validates both base and target currency codes.
// Returns an error if either code is invalid or if they are the same.
// This method adds value by validating the relationship between two currency codes,
// which is domain logic that doesn't belong in the entity.
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
