package entity

import "errors"

// Domain errors for the currency exchange rate service.
var (
	// ErrInvalidCurrencyCode indicates an invalid currency code format
	ErrInvalidCurrencyCode = errors.New("invalid currency code")

	// ErrInvalidExchangeRate indicates an invalid exchange rate value
	ErrInvalidExchangeRate = errors.New("invalid exchange rate")

	// ErrInvalidTimestamp indicates an invalid timestamp
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// ErrCurrencyCodeMismatch indicates that base and target currencies are the same
	ErrCurrencyCodeMismatch = errors.New("base and target currencies cannot be the same")

	// ErrRateNotFound indicates that an exchange rate was not found
	ErrRateNotFound = errors.New("exchange rate not found")
)

