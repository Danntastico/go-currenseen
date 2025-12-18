package dto

import (
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// ToRateResponse converts a domain ExchangeRate entity to a RateResponse DTO.
func ToRateResponse(rate *entity.ExchangeRate) RateResponse {
	if rate == nil {
		return RateResponse{}
	}

	return RateResponse{
		Base:      rate.Base.String(),
		Target:    rate.Target.String(),
		Rate:      rate.Rate,
		Timestamp: rate.Timestamp,
		Stale:     rate.Stale,
	}
}

// ToRatesResponse converts a slice of domain ExchangeRate entities to a RatesResponse DTO.
// The base currency is extracted from the first rate (all rates should have the same base).
// If rates is empty, returns a RatesResponse with empty rates map.
func ToRatesResponse(rates []*entity.ExchangeRate) RatesResponse {
	if len(rates) == 0 {
		return RatesResponse{
			Rates:     make(map[string]RateResponse),
			Timestamp: time.Now(),
		}
	}

	ratesMap := make(map[string]RateResponse)
	var latestTimestamp time.Time
	hasStale := false
	base := rates[0].Base.String()

	for _, rate := range rates {
		if rate == nil {
			continue
		}

		rateResponse := ToRateResponse(rate)
		ratesMap[rateResponse.Target] = rateResponse

		// Track the latest timestamp
		if rate.Timestamp.After(latestTimestamp) {
			latestTimestamp = rate.Timestamp
		}

		// Check if any rate is stale
		if rate.Stale {
			hasStale = true
		}
	}

	return RatesResponse{
		Base:      base,
		Rates:     ratesMap,
		Timestamp: latestTimestamp,
		Stale:     hasStale,
	}
}

// ToErrorResponse creates an ErrorResponse from an error.
func ToErrorResponse(err error, code string) ErrorResponse {
	return ErrorResponse{
		Error:     err.Error(),
		Code:      code,
		Timestamp: time.Now(),
	}
}
