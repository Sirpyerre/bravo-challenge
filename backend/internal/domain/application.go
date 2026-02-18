package domain

import (
	"time"

	"github.com/google/uuid"
)

type ApplicationStatus string

const (
	StatusPending    ApplicationStatus = "PENDING"
	StatusValidating ApplicationStatus = "VALIDATING"
	StatusApproved   ApplicationStatus = "APPROVED"
	StatusDenied     ApplicationStatus = "DENIED"
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"
	RiskMedium RiskLevel = "MEDIUM"
	RiskHigh   RiskLevel = "HIGH"
)

type Application struct {
	ID               uuid.UUID         `json:"id" db:"id"`
	UserID           uuid.UUID         `json:"user_id" db:"user_id"`
	Country          string            `json:"country" db:"country"`
	FullName         string            `json:"full_name" db:"full_name"`
	IdentityDocument string            `json:"identity_document" db:"identity_document"`
	MonthlyIncome    float64           `json:"monthly_income" db:"monthly_income"`
	RequestedAmount  float64           `json:"requested_amount" db:"requested_amount"`
	Status           ApplicationStatus `json:"status" db:"status"`
	RiskLevel        *RiskLevel        `json:"risk_level,omitempty" db:"risk_level"`
	Notes            *string           `json:"notes,omitempty" db:"notes"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at" db:"updated_at"`
}
