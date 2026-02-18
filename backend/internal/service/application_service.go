package service

import (
	"context"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/Sirpyerre/bravo-challenge/internal/validation"
	"github.com/google/uuid"
)

type ApplicationService struct {
	appRepo repository.ApplicationRepository
}

type CreateApplicationRequest struct {
	Country          string  `json:"country"`
	FullName         string  `json:"full_name"`
	IdentityDocument string  `json:"identity_document"`
	MonthlyIncome    float64 `json:"monthly_income"`
	RequestedAmount  float64 `json:"requested_amount"`
}

type UpdateApplicationRequest struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes"`
}

type ListApplicationsRequest struct {
	Country string
	Status  string
	Limit   int
	Offset  int
}

type ListApplicationsResponse struct {
	Applications []domain.Application `json:"applications"`
	Total        int                  `json:"total"`
}

func NewApplicationService(appRepo repository.ApplicationRepository) *ApplicationService {
	return &ApplicationService{appRepo: appRepo}
}

func (s *ApplicationService) Create(ctx context.Context, userID uuid.UUID, req CreateApplicationRequest) (*domain.Application, error) {
	// Validar campos requeridos
	if req.Country == "" || req.FullName == "" || req.IdentityDocument == "" {
		return nil, fmt.Errorf("country, full_name e identity_document son requeridos")
	}
	if req.MonthlyIncome <= 0 || req.RequestedAmount <= 0 {
		return nil, fmt.Errorf("monthly_income y requested_amount deben ser mayores a 0")
	}

	// Validación por país
	validator, err := validation.NewValidator(req.Country)
	if err != nil {
		return nil, err
	}

	if err := validator.ValidateIdentityDocument(req.IdentityDocument); err != nil {
		return nil, err
	}

	if err := validator.ValidateAmount(req.MonthlyIncome, req.RequestedAmount); err != nil {
		return nil, err
	}

	app := &domain.Application{
		UserID:           userID,
		Country:          req.Country,
		FullName:         req.FullName,
		IdentityDocument: req.IdentityDocument,
		MonthlyIncome:    req.MonthlyIncome,
		RequestedAmount:  req.RequestedAmount,
		Status:           domain.StatusPending,
	}

	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("crear solicitud: %w", err)
	}

	return app, nil
}

func (s *ApplicationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	app, err := s.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obtener solicitud: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("solicitud no encontrada")
	}
	return app, nil
}

func (s *ApplicationService) List(ctx context.Context, userID uuid.UUID, req ListApplicationsRequest) (*ListApplicationsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	apps, total, err := s.appRepo.FindByUserID(ctx, userID, req.Country, req.Status, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("listar solicitudes: %w", err)
	}

	return &ListApplicationsResponse{
		Applications: apps,
		Total:        total,
	}, nil
}

func (s *ApplicationService) UpdateStatus(ctx context.Context, id uuid.UUID, req UpdateApplicationRequest) error {
	// Validar que el status sea válido
	status := domain.ApplicationStatus(req.Status)
	switch status {
	case domain.StatusPending, domain.StatusValidating, domain.StatusApproved, domain.StatusDenied:
		// ok
	default:
		return fmt.Errorf("status inválido: %s", req.Status)
	}

	if err := s.appRepo.UpdateStatus(ctx, id, status, req.Notes); err != nil {
		return fmt.Errorf("actualizar solicitud: %w", err)
	}

	return nil
}
