package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"server/internal/domain"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	createUserWithCredentialsFunc func(ctx context.Context, credentials domain.Credentials) error
	getUserByEmailFunc            func(ctx context.Context, email string) (*domain.User, error)
	getUserByOAuthInfoFunc        func(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.User, error)
	createUserWithOAuthInfoFunc   func(ctx context.Context, oauthInfo *domain.OAuthUserInfo) error
}

func (m *mockUserRepository) CreateUserWithCredentials(ctx context.Context, credentials domain.Credentials) error {
	if m.createUserWithCredentialsFunc != nil {
		return m.createUserWithCredentialsFunc(ctx, credentials)
	}
	return nil
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) GetUserByOAuthInfo(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.User, error) {
	if m.getUserByOAuthInfoFunc != nil {
		return m.getUserByOAuthInfoFunc(ctx, oauthInfo)
	}
	return nil, nil
}

func (m *mockUserRepository) CreateUserWithOAuthInfo(ctx context.Context, oauthInfo *domain.OAuthUserInfo) error {
	if m.createUserWithOAuthInfoFunc != nil {
		return m.createUserWithOAuthInfoFunc(ctx, oauthInfo)
	}
	return nil
}

type mockSessionRepository struct {
	storeSessionFunc  func(ctx context.Context, session *domain.Session) error
	getSessionFunc    func(ctx context.Context, token string) (*domain.Session, error)
	deleteSessionFunc func(ctx context.Context, token string) error
}

func (m *mockSessionRepository) StoreSession(ctx context.Context, session *domain.Session) error {
	if m.storeSessionFunc != nil {
		return m.storeSessionFunc(ctx, session)
	}
	return nil
}

func (m *mockSessionRepository) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	if m.getSessionFunc != nil {
		return m.getSessionFunc(ctx, token)
	}
	return nil, nil
}

func (m *mockSessionRepository) DeleteSession(ctx context.Context, token string) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(ctx, token)
	}
	return nil
}

type mockOAuthGateway struct {
	getOAuthUserInfoFunc func(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error)
	getGoogleAuthURLFunc func(ctx context.Context, purpose, state string) string
}

func (m *mockOAuthGateway) GetOAuthUserInfo(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error) {
	if m.getOAuthUserInfoFunc != nil {
		return m.getOAuthUserInfoFunc(ctx, code, purpose)
	}
	return nil, nil
}

func (m *mockOAuthGateway) GetGoogleAuthURL(ctx context.Context, purpose, state string) string {
	if m.getGoogleAuthURLFunc != nil {
		return m.getGoogleAuthURLFunc(ctx, purpose, state)
	}
	return ""
}

type mockCSRFTokenGenerator struct {
	getCSRFTokenFunc func(ctx context.Context) (string, error)
}

func (m *mockCSRFTokenGenerator) GetCSRFToken(ctx context.Context) (string, error) {
	if m.getCSRFTokenFunc != nil {
		return m.getCSRFTokenFunc(ctx)
	}
	return "", nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func TestUseCase_SignUpWithEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mockUserRepository)
		expectedError error
	}{
		{
			name:     "successful signup",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(m *mockUserRepository) {
				m.createUserWithCredentialsFunc = func(ctx context.Context, credentials domain.Credentials) error {
					if credentials.Email != "test@example.com" {
						t.Errorf("expected email test@example.com, got %s", credentials.Email)
					}
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:     "invalid email",
			email:    "invalid-email",
			password: "password123",
			setupMocks: func(m *mockUserRepository) {
			},
			expectedError: domain.ErrNotValidEmail,
		},
		{
			name:     "user already exists",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(m *mockUserRepository) {
				m.createUserWithCredentialsFunc = func(ctx context.Context, credentials domain.Credentials) error {
					return domain.ErrUserAlreadyExists
				}
			},
			expectedError: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepository{}
			mockSessionRepo := &mockSessionRepository{}
			mockOAuthGateway := &mockOAuthGateway{}
			mockCSRF := &mockCSRFTokenGenerator{}

			tt.setupMocks(mockUserRepo)

			uc := NewUseCase(logger, mockUserRepo, mockSessionRepo, mockOAuthGateway, mockCSRF)
			err := uc.SignUpWithEmail(ctx, tt.email, tt.password)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestUseCase_LogInWithEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mockUserRepository, *mockSessionRepository)
		expectedError error
		expectSession bool
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(mu *mockUserRepository, ms *mockSessionRepository) {
				hashedPassword, _ := hashPassword("password123")
				mu.getUserByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{
						ID:       1,
						Email:    "test@example.com",
						Password: hashedPassword,
					}, nil
				}
				ms.storeSessionFunc = func(ctx context.Context, session *domain.Session) error {
					if session.UserID != 1 {
						t.Errorf("expected userID 1, got %d", session.UserID)
					}
					return nil
				}
			},
			expectedError: nil,
			expectSession: true,
		},
		{
			name:     "user not found",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(mu *mockUserRepository, ms *mockSessionRepository) {
				mu.getUserByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return nil, domain.ErrUserNotExists
				}
			},
			expectedError: domain.ErrUserNotExists,
			expectSession: false,
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(mu *mockUserRepository, ms *mockSessionRepository) {
				hashedPassword, _ := hashPassword("password123")
				mu.getUserByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{
						ID:       1,
						Email:    "test@example.com",
						Password: hashedPassword,
					}, nil
				}
			},
			expectedError: domain.ErrInvalidPassword,
			expectSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepository{}
			mockSessionRepo := &mockSessionRepository{}
			mockOAuthGateway := &mockOAuthGateway{}
			mockCSRF := &mockCSRFTokenGenerator{}

			tt.setupMocks(mockUserRepo, mockSessionRepo)

			uc := NewUseCase(logger, mockUserRepo, mockSessionRepo, mockOAuthGateway, mockCSRF)
			session, err := uc.LogInWithEmail(ctx, tt.email, tt.password)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				if session != nil {
					t.Errorf("expected nil session, got %v", session)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if session == nil {
					t.Error("expected session, got nil")
				} else {
					if session.UserID != 1 {
						t.Errorf("expected userID 1, got %d", session.UserID)
					}
					if session.Token == "" {
						t.Error("expected non-empty token")
					}
				}
			}
		})
	}
}

func TestUseCase_LogInWithGoogle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		code          string
		setupMocks    func(*mockOAuthGateway, *mockUserRepository, *mockSessionRepository)
		expectedError error
		expectSession bool
	}{
		{
			name: "successful login",
			code: "valid_code",
			setupMocks: func(mo *mockOAuthGateway, mu *mockUserRepository, ms *mockSessionRepository) {
				mo.getOAuthUserInfoFunc = func(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error) {
					return &domain.OAuthUserInfo{
						Email:        "test@example.com",
						ProviderName: "google",
						Sub:          "123456",
					}, nil
				}
				mu.getUserByOAuthInfoFunc = func(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.User, error) {
					return &domain.User{
						ID:    1,
						Email: "test@example.com",
					}, nil
				}
				ms.storeSessionFunc = func(ctx context.Context, session *domain.Session) error {
					return nil
				}
			},
			expectedError: nil,
			expectSession: true,
		},
		{
			name: "empty code",
			code: "",
			setupMocks: func(mo *mockOAuthGateway, mu *mockUserRepository, ms *mockSessionRepository) {
			},
			expectedError: domain.ErrInvalidGoogleCode,
			expectSession: false,
		},
		{
			name: "user not found",
			code: "valid_code",
			setupMocks: func(mo *mockOAuthGateway, mu *mockUserRepository, ms *mockSessionRepository) {
				mo.getOAuthUserInfoFunc = func(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error) {
					return &domain.OAuthUserInfo{
						Email:        "test@example.com",
						ProviderName: "google",
						Sub:          "123456",
					}, nil
				}
				mu.getUserByOAuthInfoFunc = func(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.User, error) {
					return nil, domain.ErrUserNotExists
				}
			},
			expectedError: domain.ErrUserNotExists,
			expectSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepository{}
			mockSessionRepo := &mockSessionRepository{}
			mockOAuthGateway := &mockOAuthGateway{}
			mockCSRF := &mockCSRFTokenGenerator{}

			tt.setupMocks(mockOAuthGateway, mockUserRepo, mockSessionRepo)

			uc := NewUseCase(logger, mockUserRepo, mockSessionRepo, mockOAuthGateway, mockCSRF)
			session, err := uc.LogInWithGoogle(ctx, tt.code)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				if session != nil {
					t.Errorf("expected nil session, got %v", session)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if session == nil {
					t.Error("expected session, got nil")
				} else {
					if session.UserID != 1 {
						t.Errorf("expected userID 1, got %d", session.UserID)
					}
					if session.Token == "" {
						t.Error("expected non-empty token")
					}
				}
			}
		})
	}
}

func TestUseCase_SignUpWithGoogle(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		code          string
		setupMocks    func(*mockOAuthGateway, *mockUserRepository)
		expectedError error
	}{
		{
			name: "successful signup",
			code: "valid_code",
			setupMocks: func(mo *mockOAuthGateway, mu *mockUserRepository) {
				mo.getOAuthUserInfoFunc = func(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error) {
					return &domain.OAuthUserInfo{
						Email:        "test@example.com",
						ProviderName: "google",
						Sub:          "123456",
					}, nil
				}
				mu.createUserWithOAuthInfoFunc = func(ctx context.Context, oauthInfo *domain.OAuthUserInfo) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name: "empty code",
			code: "",
			setupMocks: func(mo *mockOAuthGateway, mu *mockUserRepository) {
			},
			expectedError: domain.ErrInvalidGoogleCode,
		},
		{
			name: "user already exists",
			code: "valid_code",
			setupMocks: func(mo *mockOAuthGateway, mu *mockUserRepository) {
				mo.getOAuthUserInfoFunc = func(ctx context.Context, code, purpose string) (*domain.OAuthUserInfo, error) {
					return &domain.OAuthUserInfo{
						Email:        "test@example.com",
						ProviderName: "google",
						Sub:          "123456",
					}, nil
				}
				mu.createUserWithOAuthInfoFunc = func(ctx context.Context, oauthInfo *domain.OAuthUserInfo) error {
					return domain.ErrUserAlreadyExists
				}
			},
			expectedError: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepository{}
			mockSessionRepo := &mockSessionRepository{}
			mockOAuthGateway := &mockOAuthGateway{}
			mockCSRF := &mockCSRFTokenGenerator{}

			tt.setupMocks(mockOAuthGateway, mockUserRepo)

			uc := NewUseCase(logger, mockUserRepo, mockSessionRepo, mockOAuthGateway, mockCSRF)
			err := uc.SignUpWithGoogle(ctx, tt.code)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestUseCase_GetGoogleAuthURL(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		purpose       string
		setupMocks    func(*mockCSRFTokenGenerator, *mockOAuthGateway)
		expectedError error
		expectedURL   string
		expectedState string
	}{
		{
			name:    "successful get URL",
			purpose: "login",
			setupMocks: func(mc *mockCSRFTokenGenerator, mo *mockOAuthGateway) {
				mc.getCSRFTokenFunc = func(ctx context.Context) (string, error) {
					return "csrf_token_123", nil
				}
				mo.getGoogleAuthURLFunc = func(ctx context.Context, purpose, state string) string {
					return "https://accounts.google.com/auth?state=csrf_token_123"
				}
			},
			expectedError: nil,
			expectedURL:   "https://accounts.google.com/auth?state=csrf_token_123",
			expectedState: "csrf_token_123",
		},
		{
			name:    "csrf token error",
			purpose: "login",
			setupMocks: func(mc *mockCSRFTokenGenerator, mo *mockOAuthGateway) {
				mc.getCSRFTokenFunc = func(ctx context.Context) (string, error) {
					return "", errors.New("csrf error")
				}
			},
			expectedError: errors.New("csrf error"),
			expectedURL:   "",
			expectedState: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepository{}
			mockSessionRepo := &mockSessionRepository{}
			mockOAuthGateway := &mockOAuthGateway{}
			mockCSRF := &mockCSRFTokenGenerator{}

			tt.setupMocks(mockCSRF, mockOAuthGateway)

			uc := NewUseCase(logger, mockUserRepo, mockSessionRepo, mockOAuthGateway, mockCSRF)
			url, state, err := uc.GetGoogleAuthURL(ctx, tt.purpose)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if url != "" {
					t.Errorf("expected empty URL, got %s", url)
				}
				if state != "" {
					t.Errorf("expected empty state, got %s", state)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if url != tt.expectedURL {
					t.Errorf("expected URL %s, got %s", tt.expectedURL, url)
				}
				if state != tt.expectedState {
					t.Errorf("expected state %s, got %s", tt.expectedState, state)
				}
			}
		})
	}
}

func TestUseCase_LogOut(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		session       *domain.Session
		setupMocks    func(*mockSessionRepository)
		expectedError error
	}{
		{
			name: "successful logout",
			session: &domain.Session{
				Token:     "token123",
				UserID:    1,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(ms *mockSessionRepository) {
				ms.deleteSessionFunc = func(ctx context.Context, token string) error {
					if token != "token123" {
						t.Errorf("expected token token123, got %s", token)
					}
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name: "delete session error",
			session: &domain.Session{
				Token:     "token123",
				UserID:    1,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			setupMocks: func(ms *mockSessionRepository) {
				ms.deleteSessionFunc = func(ctx context.Context, token string) error {
					return errors.New("delete error")
				}
			},
			expectedError: errors.New("delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &mockUserRepository{}
			mockSessionRepo := &mockSessionRepository{}
			mockOAuthGateway := &mockOAuthGateway{}
			mockCSRF := &mockCSRFTokenGenerator{}

			tt.setupMocks(mockSessionRepo)

			uc := NewUseCase(logger, mockUserRepo, mockSessionRepo, mockOAuthGateway, mockCSRF)
			err := uc.LogOut(ctx, tt.session)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
