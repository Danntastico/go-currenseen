package entity

import (
	"fmt"
	"regexp"
	"strings"
)

// CurrencyCode represents an ISO 4217 currency code (3 uppercase letters).
// It provides validation and type safety for currency codes.
type CurrencyCode string

const (
	// CurrencyCodeLength is the required length for ISO 4217 currency codes
	CurrencyCodeLength = 3

	// currencyCodePattern is the regex pattern for valid currency codes
	currencyCodePattern = `^[A-Z]{3}$`
)

var (
	// currencyCodeRegex is the compiled regex for currency code validation
	currencyCodeRegex = regexp.MustCompile(currencyCodePattern)
)

// NewCurrencyCode creates a new CurrencyCode with validation.
// Returns an error if the code is invalid.
func NewCurrencyCode(code string) (CurrencyCode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("%w: currency code cannot be empty", ErrInvalidCurrencyCode)
	}

	if len(code) != CurrencyCodeLength {
		return "", fmt.Errorf("%w: currency code must be exactly %d characters, got %d", ErrInvalidCurrencyCode, CurrencyCodeLength, len(code))
	}

	upperCode := strings.ToUpper(code)
	if !currencyCodeRegex.MatchString(upperCode) {
		return "", fmt.Errorf("%w: currency code must be 3 uppercase letters, got %q", ErrInvalidCurrencyCode, code)
	}

	return CurrencyCode(upperCode), nil
}

// String returns the string representation of the currency code.
func (c CurrencyCode) String() string {
	return string(c)
}

// IsValid checks if the currency code is valid.
func (c CurrencyCode) IsValid() bool {
	return currencyCodeRegex.MatchString(string(c))
}

// Equal checks if two currency codes are equal (case-insensitive).
func (c CurrencyCode) Equal(other CurrencyCode) bool {
	return strings.EqualFold(string(c), string(other))
}

