package validation

import (
	"fmt"
	"regexp"
)

type brValidator struct{}

func (v *brValidator) Country() string { return "BR" }

var cpfRegex = regexp.MustCompile(`^\d{11}$`)

// ValidateIdentityDocument valida CPF brasileño:
//   - Exactamente 11 dígitos
//   - No todos los dígitos iguales (ex: 11111111111 es inválido)
//   - Dígitos verificadores válidos (algoritmo módulo 11)
func (v *brValidator) ValidateIdentityDocument(doc string) error {
	if !cpfRegex.MatchString(doc) {
		return fmt.Errorf("CPF inválido: deve ter 11 dígitos numéricos")
	}

	// Dígitos iguales son inválidos (11111111111, 22222222222, etc.)
	allSame := true
	for i := 1; i < 11; i++ {
		if doc[i] != doc[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return fmt.Errorf("CPF inválido: dígitos não podem ser todos iguais")
	}

	digits := make([]int, 11)
	for i := 0; i < 11; i++ {
		digits[i] = int(doc[i] - '0')
	}

	// Primer dígito verificador
	sum := 0
	for i := 0; i < 9; i++ {
		sum += digits[i] * (10 - i)
	}
	first := (sum * 10) % 11
	if first == 10 {
		first = 0
	}
	if first != digits[9] {
		return fmt.Errorf("CPF inválido: dígito verificador incorreto")
	}

	// Segundo dígito verificador
	sum = 0
	for i := 0; i < 10; i++ {
		sum += digits[i] * (11 - i)
	}
	second := (sum * 10) % 11
	if second == 10 {
		second = 0
	}
	if second != digits[10] {
		return fmt.Errorf("CPF inválido: dígito verificador incorreto")
	}

	return nil
}

// ValidateAmount verifica capacidad de pago para Brasil:
//   - Monto mínimo: R$500
//   - Monto máximo: 4x el ingreso mensual (regla de capacidad de pago)
func (v *brValidator) ValidateAmount(monthlyIncome, requestedAmount float64) error {
	if requestedAmount < 500 {
		return fmt.Errorf("valor mínimo solicitado é R$500")
	}
	if requestedAmount > monthlyIncome*4 {
		return fmt.Errorf("valor solicitado excede 4x a renda mensal (capacidade de pagamento)")
	}
	return nil
}
