package service

import (
	"errors"
	"testing"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

func TestValidationService_ValidateCurrencyCode(t *testing.T) {
	service := NewValidationService()

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid code",
			code:    "USD",
			wantErr: false,
		},
		{
			name:    "invalid code",
			code:    "XX",
			wantErr: true,
		},
		{
			name:    "empty code",
			code:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ValidateCurrencyCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrencyCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationService_ValidateCurrencyPair(t *testing.T) {
	service := NewValidationService()

	tests := []struct {
		name      string
		baseCode  string
		targetCode string
		wantErr   bool
		errType   error
	}{
		{
			name:      "valid pair",
			baseCode:  "USD",
			targetCode: "EUR",
			wantErr:   false,
		},
		{
			name:      "invalid base",
			baseCode:  "XX",
			targetCode: "EUR",
			wantErr:   true,
		},
		{
			name:      "invalid target",
			baseCode:  "USD",
			targetCode: "YY",
			wantErr:   true,
		},
		{
			name:      "same currencies",
			baseCode:  "USD",
			targetCode: "USD",
			wantErr:   true,
			errType:   entity.ErrCurrencyCodeMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, target, err := service.ValidateCurrencyPair(tt.baseCode, tt.targetCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrencyPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if base.String() != tt.baseCode {
					t.Errorf("ValidateCurrencyPair() base = %v, want %v", base, tt.baseCode)
				}
				if target.String() != tt.targetCode {
					t.Errorf("ValidateCurrencyPair() target = %v, want %v", target, tt.targetCode)
				}
			}
			if tt.wantErr && tt.errType != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("ValidateCurrencyPair() error = %v, want error type %v", err, tt.errType)
				}
			}
		})
	}
}

func TestValidationService_IsValidCurrencyCode(t *testing.T) {
	service := NewValidationService()

	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "valid code",
			code: "USD",
			want: true,
		},
		{
			name: "invalid code",
			code: "XX",
			want: false,
		},
		{
			name: "empty code",
			code: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := service.IsValidCurrencyCode(tt.code); got != tt.want {
				t.Errorf("IsValidCurrencyCode() = %v, want %v", got, tt.want)
			}
		})
	}
}


