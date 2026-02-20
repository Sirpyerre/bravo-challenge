package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrAlreadyTerminal se devuelve cuando se intenta actualizar una aplicación
// que ya tiene un estado terminal (APPROVED o DENIED).
var ErrAlreadyTerminal = errors.New("application already in terminal state")

type ApplicationRepository interface {
	Create(ctx context.Context, app *domain.Application) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, country, status string, fromDate, toDate *time.Time, limit, offset int) ([]domain.Application, int, error)
	FindAll(ctx context.Context, country, status string, fromDate, toDate *time.Time, limit, offset int) ([]domain.Application, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, notes *string) error
	// SetValidating transiciona de PENDING → VALIDATING solo si aún está en PENDING.
	// Devuelve ErrAlreadyTerminal si el webhook ya actualizó la app antes que el worker.
	SetValidating(ctx context.Context, id uuid.UUID) error
	UpdateRiskAndStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, riskLevel domain.RiskLevel, bankReason string) error
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

func (r *applicationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, country, status string, fromDate, toDate *time.Time, limit, offset int) ([]domain.Application, int, error) {
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
	if fromDate != nil {
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *fromDate)
		argIdx++
	}
	if toDate != nil {
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *toDate)
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
	if fromDate != nil {
		dataQuery += fmt.Sprintf(" AND created_at >= $%d", dataIdx)
		dataArgs = append(dataArgs, *fromDate)
		dataIdx++
	}
	if toDate != nil {
		dataQuery += fmt.Sprintf(" AND created_at <= $%d", dataIdx)
		dataArgs = append(dataArgs, *toDate)
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

func (r *applicationRepository) FindAll(ctx context.Context, country, status string, fromDate, toDate *time.Time, limit, offset int) ([]domain.Application, int, error) {
	countQuery := `SELECT COUNT(*) FROM applications WHERE 1=1`
	args := []any{}
	argIdx := 1

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
	if fromDate != nil {
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *fromDate)
		argIdx++
	}
	if toDate != nil {
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *toDate)
		argIdx++
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count applications: %w", err)
	}

	dataQuery := `
		SELECT id, user_id, country, full_name, identity_document, monthly_income,
		       requested_amount, status, risk_level, notes, created_at, updated_at
		FROM applications WHERE 1=1`

	dataArgs := []any{}
	dataIdx := 1

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
	if fromDate != nil {
		dataQuery += fmt.Sprintf(" AND created_at >= $%d", dataIdx)
		dataArgs = append(dataArgs, *fromDate)
		dataIdx++
	}
	if toDate != nil {
		dataQuery += fmt.Sprintf(" AND created_at <= $%d", dataIdx)
		dataArgs = append(dataArgs, *toDate)
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

func (r *applicationRepository) SetValidating(ctx context.Context, id uuid.UUID) error {
	// Solo transiciona si la app está en PENDING. Si el webhook llegó primero
	// (y ya la puso en APPROVED/DENIED), no sobreescribimos el estado terminal.
	result, err := r.db.Exec(ctx,
		`UPDATE applications SET status = 'VALIDATING', updated_at = NOW()
		 WHERE id = $1 AND status = 'PENDING'`, id)
	if err != nil {
		return fmt.Errorf("set validating: %w", err)
	}
	if result.RowsAffected() == 0 {
		var exists bool
		_ = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM applications WHERE id = $1)`, id).Scan(&exists)
		if !exists {
			return fmt.Errorf("application not found")
		}
		return ErrAlreadyTerminal
	}
	return nil
}

func (r *applicationRepository) UpdateRiskAndStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus, riskLevel domain.RiskLevel, bankReason string) error {
	// La condición `status NOT IN (...)` hace la guarda atómica:
	// si otro proceso ya actualizó la aplicación a un estado terminal,
	// este UPDATE no afecta ninguna fila y devuelve ErrAlreadyTerminal.
	query := `
		UPDATE applications
		SET status = $1, risk_level = $2, notes = $3, updated_at = NOW()
		WHERE id = $4
		  AND status NOT IN ('APPROVED', 'DENIED')`

	result, err := r.db.Exec(ctx, query, status, riskLevel, bankReason, id)
	if err != nil {
		return fmt.Errorf("update application risk and status: %w", err)
	}
	if result.RowsAffected() == 0 {
		// Puede ser que no exista o que ya esté en estado terminal.
		// Consultamos para distinguir ambos casos.
		var exists bool
		_ = r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM applications WHERE id = $1)`, id).Scan(&exists)
		if !exists {
			return fmt.Errorf("application not found")
		}
		return ErrAlreadyTerminal
	}
	return nil
}
