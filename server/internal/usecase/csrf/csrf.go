package csrf

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
)

type UseCase struct {
	logger *slog.Logger
}

func NewUseCase(logger *slog.Logger) *UseCase {
	return &UseCase{
		logger: logger,
	}
}

func (uc *UseCase) GetCSRFToken(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func (uc *UseCase) ValidateCSRFToken(ctx context.Context, tokenCookie, tokenHeader string) (bool, error) {
	return hmac.Equal([]byte(tokenCookie), []byte(tokenHeader)), nil
}
