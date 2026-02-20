package credit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type handlerMockRepo struct {
	listResp    []domain.Application
	listTotal   int
	lastFrom    *time.Time
	lastTo      *time.Time
	lastCountry string
	lastStatus  string
	listAll     bool
}

func (m *handlerMockRepo) Create(ctx context.Context, app *domain.Application) error { return nil }

func (m *handlerMockRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	return nil, nil
}

func (m *handlerMockRepo) FindByUserID(ctx context.Context, userID uuid.UUID, country, status string, fromDate, toDate *time.Time, limit, offset int) ([]domain.Application, int, error) {
	m.lastFrom = fromDate
	m.lastTo = toDate
	m.lastCountry = country
	m.lastStatus = status
	m.listAll = false

	if m.listResp == nil {
		return []domain.Application{}, m.listTotal, nil
	}
	return m.listResp, m.listTotal, nil
}

func (m *handlerMockRepo) FindAll(ctx context.Context, country, status string, fromDate, toDate *time.Time, limit, offset int) ([]domain.Application, int, error) {
	m.lastFrom = fromDate
	m.lastTo = toDate
	m.lastCountry = country
	m.lastStatus = status
	m.listAll = true

	if m.listResp == nil {
		return []domain.Application{}, m.listTotal, nil
	}
	return m.listResp, m.listTotal, nil
}

func (m *handlerMockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, notes *string) error {
	return nil
}

func (m *handlerMockRepo) UpdateRiskAndStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, riskLevel domain.RiskLevel, bankReason string) error {
	return nil
}

func TestHandler_List_ParsesDateFilters(t *testing.T) {
	repo := &handlerMockRepo{}
	svc := service.NewApplicationService(repo, nil)
	h := NewHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/applications?from_date=2024-01-01&to_date=2024-02-01", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("user_id", uuid.New())
	ctx.Set("role", domain.RoleUser)

	if err := h.List(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if repo.lastFrom == nil || repo.lastTo == nil {
		t.Fatalf("expected date filters to be parsed")
	}
	if !repo.lastFrom.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("from_date parsed incorrectly: %v", repo.lastFrom)
	}
	if !repo.lastTo.Equal(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("to_date parsed incorrectly: %v", repo.lastTo)
	}
}

func TestHandler_List_InvalidDateFormat(t *testing.T) {
	repo := &handlerMockRepo{}
	svc := service.NewApplicationService(repo, nil)
	h := NewHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/applications?from_date=invalid-date", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set("user_id", uuid.New())
	ctx.Set("role", domain.RoleUser)

	_ = h.List(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
