package entity

import (
	"testing"
)

func TestNewCurrencyCode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CurrencyCode
		wantErr bool
	}{
		{
			name:    "valid uppercase code",
			input:   "USD",
			want:    CurrencyCode("USD"),
			wantErr: false,
		},
		{
			name:    "valid lowercase code (should be uppercased)",
			input:   "usd",
			want:    CurrencyCode("USD"),
			wantErr: false,
		},
		{
			name:    "valid mixed case code (should be uppercased)",
			input:   "UsD",
			want:    CurrencyCode("USD"),
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "US",
			want:    "",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "USDD",
			want:    "",
			wantErr: true,
		},
		{
			name:    "contains numbers",
			input:   "US1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "contains special characters",
			input:   "US$",
			want:    "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "whitespace trimmed",
			input:   "  USD  ",
			want:    CurrencyCode("USD"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCurrencyCode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCurrencyCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewCurrencyCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyCode_String(t *testing.T) {
	code := CurrencyCode("USD")
	if got := code.String(); got != "USD" {
		t.Errorf("CurrencyCode.String() = %v, want USD", got)
	}
}

func TestCurrencyCode_IsValid(t *testing.T) {
	tests := []struct {
		name string
		code CurrencyCode
		want bool
	}{
		{
			name: "valid code",
			code: CurrencyCode("USD"),
			want: true,
		},
		{
			name: "invalid code - lowercase",
			code: CurrencyCode("usd"),
			want: false,
		},
		{
			name: "invalid code - too short",
			code: CurrencyCode("US"),
			want: false,
		},
		{
			name: "invalid code - contains number",
			code: CurrencyCode("US1"),
			want: false,
		},
		{
			name: "empty code",
			code: CurrencyCode(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code.IsValid(); got != tt.want {
				t.Errorf("CurrencyCode.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyCode_Equal(t *testing.T) {
	tests := []struct {
		name  string
		code1 CurrencyCode
		code2 CurrencyCode
		want  bool
	}{
		{
			name:  "same codes uppercase",
			code1: CurrencyCode("USD"),
			code2: CurrencyCode("USD"),
			want:  true,
		},
		{
			name:  "same codes different case",
			code1: CurrencyCode("USD"),
			code2: CurrencyCode("usd"),
			want:  true,
		},
		{
			name:  "different codes",
			code1: CurrencyCode("USD"),
			code2: CurrencyCode("EUR"),
			want:  false,
		},
		{
			name:  "empty string code1",
			code1: CurrencyCode(""),
			code2: CurrencyCode("USD"),
			want:  false,
		},
		{
			name:  "empty string code2",
			code1: CurrencyCode("USD"),
			code2: CurrencyCode(""),
			want:  false,
		},
		{
			name:  "both empty strings",
			code1: CurrencyCode(""),
			code2: CurrencyCode(""),
			want:  true, // Empty strings are equal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.code1.Equal(tt.code2); got != tt.want {
				t.Errorf("CurrencyCode.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
