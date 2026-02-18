package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationRepository interface {
	Create(ctx context.Context, app *domain.Application) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, country, status string, limit, offset int) ([]domain.Application, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, notes *string) error
	UpdateRiskAndStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, riskLevel domain.RiskLevel) error
}

type applicationRepository struct {
	db *pgxpool.Pool
}

func NewApplicationRepository(db *pgxpool.Pool) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(ctx context.Context, app *domain.Application) error {
	query := `
		INSERT INTO applications (user_id, country, full_name, identity_document, monthly_income, requested_amount, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		app.UserID, app.Country, app.FullName, app.IdentityDocument,
		app.MonthlyIncome, app.RequestedAmount, app.Status,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}
	return nil
}

func (r *applicationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	query := `
		SELECT id, user_id, country, full_name, identity_document, monthly_income,
		       requested_amount, status, risk_level, notes, created_at, updated_at
		FROM applications WHERE id = $1`

	var app domain.Application
	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.UserID, &app.Country, &app.FullName, &app.IdentityDocument,
		&app.MonthlyIncome, &app.RequestedAmount, &app.Status, &app.RiskLevel,
		&app.Notes, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find application by id: %w", err)
	}
	return &app, nil
}

func (r *applicationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, country, status string, limit, offset int) ([]domain.Application, int, error) {
	// Consulta de conteo
	countQuery := `SELECT COUNT(*) FROM applications WHERE user_id = $1`
	args := []any{userID}
	argIdx := 2

	if country != "" {
		countQuery += fmt.Sprintf(" AND country = $%d", argIdx)
		args = append(args, country)
		argIdx++
	}
	if status != "" {
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}

	// Consulta de datos
	dataQuery := `
		SELECT id, user_id, country, full_name, identity_document, monthly_income,
		       requested_amount, status, risk_level, notes, created_at, updated_at
		FROM applications WHERE user_id = $1`

	dataArgs := []any{userID}
	dataIdx := 2

	if country != "" {
		dataQuery += fmt.Sprintf(" AND country = $%d", dataIdx)
		dataArgs = append(dataArgs, country)
		dataIdx++
	}
	if status != "" {
		dataQuery += fmt.Sprintf(" AND status = $%d", dataIdx)
		dataArgs = append(dataArgs, status)
		dataIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", dataIdx, dataIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query applications: %w", err)
	}
	defer rows.Close()

	var apps []domain.Application
	for rows.Next() {
		var app domain.Application
		if err := rows.Scan(
			&app.ID, &app.UserID, &app.Country, &app.FullName, &app.IdentityDocument,
			&app.MonthlyIncome, &app.RequestedAmount, &app.Status, &app.RiskLevel,
			&app.Notes, &app.CreatedAt, &app.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan application: %w", err)
		}
		apps = append(apps, app)
	}

	return apps, total, nil
}

func (r *applicationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, notes *string) error {
	query := `
		UPDATE applications
		SET status = $1, notes = $2, updated_at = NOW()
		WHERE id = $3`

	result, err := r.db.Exec(ctx, query, status, notes, id)
	if err != nil {
		return fmt.Errorf("update application status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}
	return nil
}

func (r *applicationRepository) UpdateRiskAndStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, riskLevel domain.RiskLevel) error {
	query := `
		UPDATE applications
		SET status = $1, risk_level = $2, updated_at = NOW()
		WHERE id = $3`

	result, err := r.db.Exec(ctx, query, status, riskLevel, id)
	if err != nil {
		return fmt.Errorf("update application risk and status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}
	return nil
}
