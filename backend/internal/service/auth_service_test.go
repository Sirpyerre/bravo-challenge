package service_test

import (
	"context"
	"testing"

	"github.com/Sirpyerre/bravo-challenge/internal/domain"
	"github.com/Sirpyerre/bravo-challenge/internal/service"
	"github.com/google/uuid"
)

const testSecret = "test-secret-key"

// --- Mock UserRepository ---

type mockUserRepo struct {
	users  map[string]*domain.User
	createErr error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	user.ID = uuid.New()
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

// --- Tests ---

func TestAuthService_Register_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	resp, err := svc.Register(context.Background(), service.RegisterRequest{
		Email:    "user@test.com",
		Password: "secret123",
		Country:  "MX",
	})
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("Register() returned empty token")
	}
	if resp.UserID == uuid.Nil {
		t.Error("Register() returned nil UserID")
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	req := service.RegisterRequest{Email: "dup@test.com", Password: "pass", Country: "BR"}
	if _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("first Register() unexpected error: %v", err)
	}

	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Error("expected error on duplicate email, got nil")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	// Register first
	if _, err := svc.Register(context.Background(), service.RegisterRequest{
		Email: "login@test.com", Password: "mypassword", Country: "MX",
	}); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	resp, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "login@test.com", Password: "mypassword",
	})
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if resp.Token == "" {
		t.Error("Login() returned empty token")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	if _, err := svc.Register(context.Background(), service.RegisterRequest{
		Email: "pwd@test.com", Password: "correct", Country: "MX",
	}); err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	_, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "pwd@test.com", Password: "wrong",
	})
	if err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	_, err := svc.Login(context.Background(), service.LoginRequest{
		Email: "ghost@test.com", Password: "any",
	})
	if err == nil {
		t.Error("expected error for unknown user, got nil")
	}
}

func TestAuthService_ValidateToken_Valid(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	resp, err := svc.Register(context.Background(), service.RegisterRequest{
		Email: "vt@test.com", Password: "pass", Country: "BR",
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	claims, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("ValidateToken() unexpected error: %v", err)
	}
	if claims.Email != "vt@test.com" {
		t.Errorf("claims.Email = %q, want %q", claims.Email, "vt@test.com")
	}
	if claims.Country != "BR" {
		t.Errorf("claims.Country = %q, want %q", claims.Country, "BR")
	}
}

func TestAuthService_ValidateToken_Tampered(t *testing.T) {
	repo := newMockUserRepo()
	svc := service.NewAuthService(repo, testSecret)

	_, err := svc.ValidateToken("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJoYWNrZXIifQ.invalid_sig")
	if err == nil {
		t.Error("expected error for tampered token, got nil")
	}
}

func TestAuthService_ValidateToken_WrongSecret(t *testing.T) {
	repo := newMockUserRepo()
	svc1 := service.NewAuthService(repo, "secret-a")
	svc2 := service.NewAuthService(repo, "secret-b")

	resp, err := svc1.Register(context.Background(), service.RegisterRequest{
		Email: "cross@test.com", Password: "pass", Country: "MX",
	})
	if err != nil {
		t.Fatalf("Register() error: %v", err)
	}

	_, err = svc2.ValidateToken(resp.Token)
	if err == nil {
		t.Error("expected error when validating token with wrong secret, got nil")
	}
}
