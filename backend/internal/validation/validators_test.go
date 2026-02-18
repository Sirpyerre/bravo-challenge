package validation_test

import (
	"testing"

	"github.com/Sirpyerre/bravo-challenge/internal/validation"
)

// --- MX Validator ---

func TestMXValidator_ValidateIdentityDocument(t *testing.T) {
	v, _ := validation.NewValidator("MX")

	tests := []struct {
		name    string
		doc     string
		wantErr bool
	}{
		{"valid CURP", "LOOE890101HDFPSL09", false},
		{"too short", "LOOE890101", true},
		{"lowercase", "looe890101hdfpsl09", true},
		{"empty", "", true},
		{"wrong format - no sex indicator", "LOOE890101XDFPSL09", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateIdentityDocument(tt.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIdentityDocument(%q) error = %v, wantErr %v", tt.doc, err, tt.wantErr)
			}
		})
	}
}

func TestMXValidator_ValidateAmount(t *testing.T) {
	v, _ := validation.NewValidator("MX")

	tests := []struct {
		name          string
		monthlyIncome float64
		requested     float64
		wantErr       bool
	}{
		{"valid amount", 10000, 5000, false},
		{"exact max (5x)", 10000, 50000, false},
		{"below minimum (999)", 10000, 999, true},
		{"exceeds 5x income", 10000, 50001, true},
		{"minimum edge (1000)", 10000, 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateAmount(tt.monthlyIncome, tt.requested)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%v, %v) error = %v, wantErr %v", tt.monthlyIncome, tt.requested, err, tt.wantErr)
			}
		})
	}
}

// --- BR Validator ---

func TestBRValidator_ValidateIdentityDocument(t *testing.T) {
	v, _ := validation.NewValidator("BR")

	tests := []struct {
		name    string
		doc     string
		wantErr bool
	}{
		{"valid CPF 11 digits", "12345678901", false},
		{"too short", "1234567890", true},
		{"too long", "123456789012", true},
		{"contains letters", "1234567890A", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateIdentityDocument(tt.doc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIdentityDocument(%q) error = %v, wantErr %v", tt.doc, err, tt.wantErr)
			}
		})
	}
}

func TestBRValidator_ValidateAmount(t *testing.T) {
	v, _ := validation.NewValidator("BR")

	tests := []struct {
		name          string
		monthlyIncome float64
		requested     float64
		wantErr       bool
	}{
		{"valid amount", 5000, 2000, false},
		{"exact max (4x)", 5000, 20000, false},
		{"below minimum (499)", 5000, 499, true},
		{"exceeds 4x income", 5000, 20001, true},
		{"minimum edge (500)", 5000, 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateAmount(tt.monthlyIncome, tt.requested)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount(%v, %v) error = %v, wantErr %v", tt.monthlyIncome, tt.requested, err, tt.wantErr)
			}
		})
	}
}

// --- Factory ---

func TestNewValidator_UnknownCountry(t *testing.T) {
	_, err := validation.NewValidator("XX")
	if err == nil {
		t.Error("expected error for unknown country, got nil")
	}
}

func TestNewValidator_SupportedCountries(t *testing.T) {
	for _, country := range []string{"MX", "BR"} {
		v, err := validation.NewValidator(country)
		if err != nil {
			t.Errorf("NewValidator(%q) unexpected error: %v", country, err)
		}
		if v.Country() != country {
			t.Errorf("Country() = %q, want %q", v.Country(), country)
		}
	}
}
