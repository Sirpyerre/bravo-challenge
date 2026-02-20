package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/metrics"
	"github.com/Sirpyerre/bravo-challenge/internal/repository"
	"github.com/Sirpyerre/bravo-challenge/internal/validation"
	"github.com/Sirpyerre/bravo-challenge/pkg/rabbitmq"
	"github.com/google/uuid"
)

type ApplicationService struct {
	appRepo   repository.ApplicationRepository
	publisher *rabbitmq.Publisher
}

type CreateApplicationRequest struct {
	Country          string  `json:"country"           example:"MX"`
	FullName         string  `json:"full_name"         example:"Juan Pérez"`
	IdentityDocument string  `json:"identity_document" example:"CURP123456"`
	MonthlyIncome    float64 `json:"monthly_income"    example:"15000"`
	RequestedAmount  float64 `json:"requested_amount"  example:"50000"`
}

type UpdateApplicationRequest struct {
	Status string  `json:"status" example:"APPROVED"`
	Notes  *string `json:"notes"  example:"Aprobado por evaluación de riesgo"`
}

type ListApplicationsRequest struct {
	Country  string
	Status   string
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
}

type ListApplicationsResponse struct {
	Applications []domain.Application `json:"applications"`
	Total        int                  `json:"total"`
}

func NewApplicationService(appRepo repository.ApplicationRepository, publisher *rabbitmq.Publisher) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, publisher: publisher}
}

func (s *ApplicationService) Create(ctx context.Context, userID uuid.UUID, req CreateApplicationRequest) (*domain.Application, error) {
	if req.Country == "" || req.FullName == "" || req.IdentityDocument == "" {
		return nil, fmt.Errorf("country, full_name e identity_document son requeridos")
	}
	if req.MonthlyIncome <= 0 || req.RequestedAmount <= 0 {
		return nil, fmt.Errorf("monthly_income y requested_amount deben ser mayores a 0")
	}

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

	metrics.ApplicationsCreated.WithLabelValues(app.Country).Inc()

	// Publicar evento para procesamiento asíncrono
	s.publishEvent("application.created", domain.Event{
		ID:            uuid.New().String(),
		Type:          "application.created",
		ApplicationID: app.ID,
		UserID:        app.UserID,
		Data:          app,
		CreatedAt:     time.Now(),
	})

	return app, nil
}

func (s *ApplicationService) publishEvent(routingKey string, event domain.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	s.publisher.Publish(routingKey, data)
}

func (s *ApplicationService) GetByID(ctx context.Context, requesterID uuid.UUID, role domain.Role, id uuid.UUID) (*domain.Application, error) {
	app, err := s.appRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("obtener solicitud: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("solicitud no encontrada")
	}
	if !role.IsPrivileged() && app.UserID != requesterID {
		return nil, fmt.Errorf("acceso no permitido")
	}
	return app, nil
}

func (s *ApplicationService) List(ctx context.Context, userID uuid.UUID, role domain.Role, req ListApplicationsRequest) (*ListApplicationsResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	if req.FromDate != nil && req.ToDate != nil && req.ToDate.Before(*req.FromDate) {
		return nil, fmt.Errorf("to_date no puede ser anterior a from_date")
	}

	var (
		apps  []domain.Application
		total int
		err   error
	)

	if role.IsPrivileged() {
		apps, total, err = s.appRepo.FindAll(ctx, req.Country, req.Status, req.FromDate, req.ToDate, req.Limit, req.Offset)
	} else {
		apps, total, err = s.appRepo.FindByUserID(ctx, userID, req.Country, req.Status, req.FromDate, req.ToDate, req.Limit, req.Offset)
	}
	if err != nil {
		return nil, fmt.Errorf("listar solicitudes: %w", err)
	}

	return &ListApplicationsResponse{
		Applications: apps,
		Total:        total,
	}, nil
}

func (s *ApplicationService) UpdateStatus(ctx context.Context, requesterRole domain.Role, id uuid.UUID, req UpdateApplicationRequest) error {
	if !requesterRole.IsPrivileged() {
		return fmt.Errorf("acceso no permitido")
	}
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
