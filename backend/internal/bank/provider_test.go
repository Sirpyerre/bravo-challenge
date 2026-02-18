package bank_test

import (
	"testing"

	"github.com/Sirpyerre/bravo-challenge/internal/bank"
	"github.com/Sirpyerre/bravo-challenge/internal/config"
)

var testBankURLs = config.BankURLsConfig{
	MX:  "http://localhost:8080/mx",
	BR:  "http://localhost:8080/br",
	ESP: "http://localhost:8080/esp",
	PT:  "http://localhost:8080/pt",
	IT:  "http://localhost:8080/it",
	CO:  "http://localhost:8080/co",
}

func TestNewProvider_SupportedCountries(t *testing.T) {
	countries := []string{"MX", "BR", "ES", "PT", "IT", "CO"}

	for _, country := range countries {
		p, err := bank.NewProvider(country, testBankURLs)
		if err != nil {
			t.Errorf("NewProvider(%q) unexpected error: %v", country, err)
			continue
		}
		if p.Country() != country {
			t.Errorf("Country() = %q, want %q", p.Country(), country)
		}
	}
}

func TestNewProvider_UnknownCountry(t *testing.T) {
	_, err := bank.NewProvider("XX", testBankURLs)
	if err == nil {
		t.Error("expected error for unknown country, got nil")
	}
}

func TestNewProvider_EmptyCountry(t *testing.T) {
	_, err := bank.NewProvider("", testBankURLs)
	if err == nil {
		t.Error("expected error for empty country, got nil")
	}
}
