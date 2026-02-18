package validation

import (
	"fmt"
	"regexp"
)

type brValidator struct{}

func (v *brValidator) Country() string { return "BR" }

// ValidateIdentityDocument valida CPF brasileño (11 dígitos).
func (v *brValidator) ValidateIdentityDocument(doc string) error {
	cpfRegex := regexp.MustCompile(`^\d{11}$`)
	if !cpfRegex.MatchString(doc) {
		return fmt.Errorf("CPF inválido: debe tener 11 dígitos")
	}
	return nil
}

// ValidateAmount verifica reglas de monto para Brasil.
// Monto máximo: 4x el ingreso mensual. Mínimo solicitado: 500 BRL.
func (v *brValidator) ValidateAmount(monthlyIncome, requestedAmount float64) error {
	if requestedAmount < 500 {
		return fmt.Errorf("valor mínimo solicitado é R$500")
	}
	if requestedAmount > monthlyIncome*4 {
		return fmt.Errorf("valor solicitado excede 4x a renda mensal")
	}
	return nil
}
