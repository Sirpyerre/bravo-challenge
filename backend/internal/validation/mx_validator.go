package validation

import (
	"fmt"
	"regexp"
	"strconv"
)

type mxValidator struct{}

func (v *mxValidator) Country() string { return "MX" }

var curpRegex = regexp.MustCompile(`^[A-Z]{4}\d{6}[HM][A-Z]{5}[A-Z0-9]\d$`)

// Claves de entidad federativa válidas en CURP (RENAPO).
var mxStateCodes = map[string]bool{
	"AS": true, "BC": true, "BS": true, "CC": true, "CL": true,
	"CS": true, "CH": true, "DF": true, "GT": true, "GR": true,
	"HG": true, "JC": true, "MC": true, "MN": true, "MS": true,
	"NT": true, "NL": true, "OC": true, "PL": true, "QT": true,
	"QR": true, "SP": true, "SL": true, "SR": true, "TC": true,
	"TS": true, "TL": true, "VZ": true, "YN": true, "ZS": true,
	"NE": true, // Nacido en el Extranjero
}

// maxDaysInMonth permite validar el día sin importar el año exacto (febrero: 29 días máx).
var maxDaysInMonth = [13]int{0, 31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// ValidateIdentityDocument valida CURP mexicano:
//   - 18 caracteres con formato oficial (regex)
//   - Clave de entidad federativa reconocida por RENAPO
//   - Fecha de nacimiento embebida con mes y día válidos
func (v *mxValidator) ValidateIdentityDocument(doc string) error {
	if !curpRegex.MatchString(doc) {
		return fmt.Errorf("CURP inválido: debe tener 18 caracteres con formato válido")
	}

	// Posiciones 11-12 (base 0): clave de entidad federativa
	stateCode := doc[11:13]
	if !mxStateCodes[stateCode] {
		return fmt.Errorf("CURP inválido: clave de entidad federativa '%s' no reconocida", stateCode)
	}

	// Posiciones 6-7: mes de nacimiento (MM)
	month, _ := strconv.Atoi(doc[6:8])
	if month < 1 || month > 12 {
		return fmt.Errorf("CURP inválido: mes de nacimiento inválido (%02d)", month)
	}

	// Posiciones 8-9: día de nacimiento (DD)
	day, _ := strconv.Atoi(doc[8:10])
	if day < 1 || day > maxDaysInMonth[month] {
		return fmt.Errorf("CURP inválido: día de nacimiento inválido (%02d) para el mes %02d", day, month)
	}

	return nil
}

// ValidateAmount verifica capacidad de pago para México:
//   - Monto mínimo: 1,000 MXN
//   - Monto máximo: 5x el ingreso mensual
func (v *mxValidator) ValidateAmount(monthlyIncome, requestedAmount float64) error {
	if requestedAmount < 1000 {
		return fmt.Errorf("monto mínimo solicitado es 1,000 MXN")
	}
	if requestedAmount > monthlyIncome*5 {
		return fmt.Errorf("monto solicitado excede 5x el ingreso mensual")
	}
	return nil
}
