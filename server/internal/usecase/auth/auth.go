package auth

import (
	"context"
	"fmt"
	"regexp"
	"server/internal/domain"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func (uc *UseCase) SignUpWithEmail(ctx context.Context, email, password string) error {
	if !validateEmail(email) {
		return domain.ErrNotValidEmail
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password hash: %w", err)
	}

	err = uc.userRepo.CreateUserWithCredentials(ctx, domain.Credentials{
		Email:    email,
		Password: string(hash),
	})
	if err != nil {
		return fmt.Errorf("failed to create user with credentials: %w", err)
	}

	return nil
}

func checkPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (uc *UseCase) createSession(ctx context.Context, userID int64) (*domain.Session, error) {
	token := uuid.New().String()
	session := &domain.Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := uc.sessionRepo.StoreSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}
	return session, nil
}

func (uc *UseCase) LogInWithEmail(ctx context.Context, email, password string) (*domain.Session, error) {
	user, err := uc.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if !checkPassword(password, user.Password) {
		return nil, domain.ErrInvalidPassword
	}

	session, err := uc.createSession(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (uc *UseCase) LogInWithGoogle(ctx context.Context, code string) (*domain.Session, error) {
	if code == "" {
		return nil, domain.ErrInvalidGoogleCode
	}

	userInfo, err := uc.oauthGateway.GetOAuthUserInfo(ctx, code, "login")
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth user info: %w", err)
	}

	user, err := uc.userRepo.GetUserByOAuthInfo(ctx, userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	session, err := uc.createSession(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (uc *UseCase) SignUpWithGoogle(ctx context.Context, code string) error {
	if code == "" {
		return domain.ErrInvalidGoogleCode
	}

	userInfo, err := uc.oauthGateway.GetOAuthUserInfo(ctx, code, "signup")
	if err != nil {
		return fmt.Errorf("failed to get oauth user info: %w", err)
	}

	err = uc.userRepo.CreateUserWithOAuthInfo(ctx, userInfo)
	if err != nil {
		return fmt.Errorf("failed to create user with oauth info: %w", err)
	}

	return nil
}

func (uc *UseCase) GetGoogleAuthURL(ctx context.Context, purpose string) (string, string, error) {
	state, err := uc.csrfUC.GetCSRFToken(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get csrf token: %w", err)
	}
	return uc.oauthGateway.GetGoogleAuthURL(ctx, purpose, state), state, nil
}
