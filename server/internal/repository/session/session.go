package session

import (
	"context"
	"server/internal/domain"
	"sync"
	"time"
)

type Repository struct {
	sessions map[string]*domain.Session
	mu       sync.RWMutex
}

func NewRepository() *Repository {
	r := &Repository{
		sessions: make(map[string]*domain.Session),
	}
	go r.clearExpiredSessions()
	return r
}

func (r *Repository) clearExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		r.mu.RLock()
		expiredSessions := make([]string, 0)
		for token, session := range r.sessions {
			if time.Now().After(session.ExpiresAt) {
				expiredSessions = append(expiredSessions, token)
			}
		}
		r.mu.RUnlock()
		func() {
			r.mu.Lock()
			defer r.mu.Unlock()
			for _, token := range expiredSessions {
				delete(r.sessions, token)
			}
		}()
	}
}

func (r *Repository) StoreSession(_ context.Context, session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.Token] = session
	return nil
}

func (r *Repository) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[token]
	if !ok || time.Now().After(session.ExpiresAt) {
		_ = r.DeleteSession(ctx, token)
		return nil, domain.ErrSessionNotFound
	}
	return session, nil
}

func (r *Repository) DeleteSession(_ context.Context, token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
	return nil
}
