package validation

import (
	"fmt"
	"regexp"
)

type mxValidator struct{}

func (v *mxValidator) Country() string { return "MX" }

// ValidateIdentityDocument valida CURP mexicano (18 caracteres alfanuméricos).
func (v *mxValidator) ValidateIdentityDocument(doc string) error {
	curpRegex := regexp.MustCompile(`^[A-Z]{4}\d{6}[HM][A-Z]{5}[A-Z0-9]\d$`)
	if !curpRegex.MatchString(doc) {
		return fmt.Errorf("CURP inválido: debe tener 18 caracteres con formato válido")
	}
	return nil
}

// ValidateAmount verifica reglas de monto para México.
// Monto máximo: 5x el ingreso mensual. Mínimo solicitado: 1,000 MXN.
func (v *mxValidator) ValidateAmount(monthlyIncome, requestedAmount float64) error {
	if requestedAmount < 1000 {
		return fmt.Errorf("monto mínimo solicitado es 1,000 MXN")
	}
	if requestedAmount > monthlyIncome*5 {
		return fmt.Errorf("monto solicitado excede 5x el ingreso mensual")
	}
	return nil
}
