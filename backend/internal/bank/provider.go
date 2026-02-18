package bank

import (
	"context"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/config"
)

// EvaluationRequest es el payload enviado al banco para evaluación de riesgo.
type EvaluationRequest struct {
	ApplicationID    string  `json:"application_id"`
	IdentityDocument string  `json:"identity_document"`
	MonthlyIncome    float64 `json:"monthly_income"`
	RequestedAmount  float64 `json:"requested_amount"`
}

// EvaluationResponse es la respuesta del banco.
type EvaluationResponse struct {
	Approved  bool   `json:"approved"`
	RiskLevel string `json:"risk_level"`
	Reason    string `json:"reason"`
}

// BankProvider define la interfaz para adaptadores de banco por país.
type BankProvider interface {
	Evaluate(ctx context.Context, req EvaluationRequest) (*EvaluationResponse, error)
	Country() string
}

// NewProvider retorna el adaptador de banco correspondiente al país (factory).
func NewProvider(country string, bankURLs config.BankURLsConfig) (BankProvider, error) {
	switch country {
	case "MX":
		return newHTTPProvider("MX", bankURLs.MX), nil
	case "BR":
		return newHTTPProvider("BR", bankURLs.BR), nil
	case "ES":
		return newHTTPProvider("ES", bankURLs.ESP), nil
	case "PT":
		return newHTTPProvider("PT", bankURLs.PT), nil
	case "IT":
		return newHTTPProvider("IT", bankURLs.IT), nil
	case "CO":
		return newHTTPProvider("CO", bankURLs.CO), nil
	default:
		return nil, fmt.Errorf("banco no configurado para país: %s", country)
	}
}
