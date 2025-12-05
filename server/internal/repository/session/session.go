package session

import (
	"context"
	"server/internal/domain"
	"sync"
	"time"
)

const (
	sessionCleanupInterval  = 1 * time.Hour
	sessionCleanupBatchSize = 100
)

type Repository struct {
	sessions map[string]*domain.Session
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

func NewRepository() *Repository {
	r := &Repository{
		sessions: make(map[string]*domain.Session),
	}
	go r.clearExpiredSessions()
	return r
}

func (r *Repository) clearExpiredSessions() {
	ticker := time.NewTicker(sessionCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanupBatch()
	}
}

func (r *Repository) cleanupBatch() {
	now := time.Now()
	r.mu.RLock()
	expiredSessions := make([]string, 0, sessionCleanupBatchSize)
	count := 0
	for token, session := range r.sessions {
		if now.After(session.ExpiresAt) {
			expiredSessions = append(expiredSessions, token)
			count++
			if count >= sessionCleanupBatchSize {
				break
			}
		}
	}
	r.mu.RUnlock()

	if len(expiredSessions) > 0 {
		r.mu.Lock()
		for _, token := range expiredSessions {
			delete(r.sessions, token)
		}
		r.mu.Unlock()
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
	session, ok := r.sessions[token]
	expired := ok && time.Now().After(session.ExpiresAt)
	r.mu.RUnlock()

	if !ok || expired {
		if ok {
			_ = r.DeleteSession(ctx, token)
		}
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
