package service_test

import (
	"context"
	"testing"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/google/uuid"
)

// --- Mock ApplicationRepository ---

type mockAppRepo struct {
	created []*domain.Application
}

func (m *mockAppRepo) Create(ctx context.Context, app *domain.Application) error {
	app.ID = uuid.New()
	m.created = append(m.created, app)
	return nil
}

func (m *mockAppRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	for _, a := range m.created {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAppRepo) FindByUserID(ctx context.Context, userID uuid.UUID, country, status string, limit, offset int) ([]domain.Application, int, error) {
	return nil, 0, nil
}

func (m *mockAppRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, notes *string) error {
	return nil
}

func (m *mockAppRepo) UpdateRiskAndStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, risk domain.RiskLevel) error {
	return nil
}

// newTestAppService creates an ApplicationService with a mock repo and no publisher.
// Safe for tests that only exercise paths that don't reach publishEvent.
func newTestAppService() (*service.ApplicationService, *mockAppRepo) {
	repo := &mockAppRepo{}
	svc := service.NewApplicationService(repo, nil)
	return svc, repo
}

// --- Validation failure tests (no repo/publisher reached) ---

func TestApplicationService_Create_MissingRequiredFields(t *testing.T) {
	svc, _ := newTestAppService()
	userID := uuid.New()

	tests := []struct {
		name string
		req  service.CreateApplicationRequest
	}{
		{"missing country", service.CreateApplicationRequest{FullName: "Juan", IdentityDocument: "LOOE890101HDFPSL09", MonthlyIncome: 5000, RequestedAmount: 3000}},
		{"missing full_name", service.CreateApplicationRequest{Country: "MX", IdentityDocument: "LOOE890101HDFPSL09", MonthlyIncome: 5000, RequestedAmount: 3000}},
		{"missing identity_document", service.CreateApplicationRequest{Country: "MX", FullName: "Juan", MonthlyIncome: 5000, RequestedAmount: 3000}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), userID, tt.req)
			if err == nil {
				t.Errorf("Create() expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestApplicationService_Create_InvalidAmounts(t *testing.T) {
	svc, _ := newTestAppService()
	userID := uuid.New()

	tests := []struct {
		name          string
		monthlyIncome float64
		requested     float64
	}{
		{"zero income", 0, 1000},
		{"zero requested", 5000, 0},
		{"negative income", -1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), userID, service.CreateApplicationRequest{
				Country:          "MX",
				FullName:         "Juan",
				IdentityDocument: "LOOE890101HDFPSL09",
				MonthlyIncome:    tt.monthlyIncome,
				RequestedAmount:  tt.requested,
			})
			if err == nil {
				t.Errorf("Create() expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestApplicationService_Create_InvalidCountry(t *testing.T) {
	svc, _ := newTestAppService()
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateApplicationRequest{
		Country:          "XX",
		FullName:         "Test",
		IdentityDocument: "123",
		MonthlyIncome:    5000,
		RequestedAmount:  1000,
	})
	if err == nil {
		t.Error("expected error for unsupported country, got nil")
	}
}

func TestApplicationService_Create_InvalidMXCURP(t *testing.T) {
	svc, _ := newTestAppService()
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateApplicationRequest{
		Country:          "MX",
		FullName:         "Juan",
		IdentityDocument: "INVALID_CURP",
		MonthlyIncome:    5000,
		RequestedAmount:  3000,
	})
	if err == nil {
		t.Error("expected error for invalid CURP, got nil")
	}
}

func TestApplicationService_Create_MXAmountBelowMinimum(t *testing.T) {
	svc, _ := newTestAppService()
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateApplicationRequest{
		Country:          "MX",
		FullName:         "Juan",
		IdentityDocument: "LOOE890101HDFPSL09",
		MonthlyIncome:    5000,
		RequestedAmount:  500, // below 1000 MXN minimum
	})
	if err == nil {
		t.Error("expected error for amount below MX minimum, got nil")
	}
}

func TestApplicationService_Create_MXAmountExceedsIncome(t *testing.T) {
	svc, _ := newTestAppService()
	_, err := svc.Create(context.Background(), uuid.New(), service.CreateApplicationRequest{
		Country:          "MX",
		FullName:         "Juan",
		IdentityDocument: "LOOE890101HDFPSL09",
		MonthlyIncome:    5000,
		RequestedAmount:  30000, // exceeds 5x income (25000)
	})
	if err == nil {
		t.Error("expected error for amount exceeding 5x income, got nil")
	}
}

// --- UpdateStatus tests ---

func TestApplicationService_UpdateStatus_InvalidStatus(t *testing.T) {
	svc, _ := newTestAppService()
	err := svc.UpdateStatus(context.Background(), uuid.New(), service.UpdateApplicationRequest{
		Status: "INVALID_STATUS",
	})
	if err == nil {
		t.Error("expected error for invalid status, got nil")
	}
}

func TestApplicationService_UpdateStatus_ValidStatuses(t *testing.T) {
	svc, _ := newTestAppService()
	id := uuid.New()

	for _, status := range []string{"PENDING", "VALIDATING", "APPROVED", "DENIED"} {
		err := svc.UpdateStatus(context.Background(), id, service.UpdateApplicationRequest{Status: status})
		if err != nil {
			t.Errorf("UpdateStatus(%q) unexpected error: %v", status, err)
		}
	}
}
