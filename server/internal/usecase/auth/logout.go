package auth

import (
	"context"
	"server/internal/domain"
)

func (uc *UseCase) LogOut(ctx context.Context, session *domain.Session) error {
	return uc.sessionRepo.DeleteSession(ctx, session.Token)
}
