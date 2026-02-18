package validation

import "fmt"

// Validator define la interfaz de validación por país.
type Validator interface {
	ValidateIdentityDocument(doc string) error
	ValidateAmount(monthlyIncome, requestedAmount float64) error
	Country() string
}

// NewValidator retorna el validador correspondiente al país (factory).
func NewValidator(country string) (Validator, error) {
	switch country {
	case "MX":
		return &mxValidator{}, nil
	case "BR":
		return &brValidator{}, nil
	default:
		return nil, fmt.Errorf("país no soportado: %s", country)
	}
}
