package session

import (
	"context"
	"server/internal/domain"
	"testing"
	"time"
)

func TestRepository_StoreSession(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	session := &domain.Session{
		Token:     "test_token_123",
		UserID:    1,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.StoreSession(ctx, session)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	retrieved, err := repo.GetSessionByToken(ctx, "test_token_123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if retrieved == nil {
		t.Error("expected session, got nil")
		return
	}
	if retrieved.UserID != session.UserID {
		t.Errorf("expected UserID %d, got %d", session.UserID, retrieved.UserID)
	}
	if retrieved.Token != session.Token {
		t.Errorf("expected Token %s, got %s", session.Token, retrieved.Token)
	}
}

func TestRepository_GetSessionByToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		token         string
		setupSession  func(*Repository)
		expectedError error
		expectSession bool
	}{
		{
			name:  "successful get",
			token: "valid_token",
			setupSession: func(r *Repository) {
				session := &domain.Session{
					Token:     "valid_token",
					UserID:    1,
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				r.StoreSession(ctx, session)
			},
			expectedError: nil,
			expectSession: true,
		},
		{
			name:  "session not found",
			token: "invalid_token",
			setupSession: func(r *Repository) {
			},
			expectedError: domain.ErrSessionNotFound,
			expectSession: false,
		},
		{
			name:  "expired session",
			token: "expired_token",
			setupSession: func(r *Repository) {
				session := &domain.Session{
					Token:     "expired_token",
					UserID:    1,
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
				r.StoreSession(ctx, session)
			},
			expectedError: domain.ErrSessionNotFound,
			expectSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewRepository()
			tt.setupSession(repo)

			session, err := repo.GetSessionByToken(ctx, tt.token)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if err != tt.expectedError {
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
					if session.Token != tt.token {
						t.Errorf("expected token %s, got %s", tt.token, session.Token)
					}
				}
			}
		})
	}
}

func TestRepository_DeleteSession(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	session := &domain.Session{
		Token:     "test_token_delete",
		UserID:    1,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.StoreSession(ctx, session)
	if err != nil {
		t.Errorf("unexpected error storing session: %v", err)
	}

	err = repo.DeleteSession(ctx, "test_token_delete")
	if err != nil {
		t.Errorf("unexpected error deleting session: %v", err)
	}

	retrieved, err := repo.GetSessionByToken(ctx, "test_token_delete")
	if err == nil || retrieved != nil {
		t.Error("expected session to be deleted")
	}
	if err != domain.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestRepository_ConcurrentAccess(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	session1 := &domain.Session{
		Token:     "token1",
		UserID:    1,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	session2 := &domain.Session{
		Token:     "token2",
		UserID:    2,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := repo.StoreSession(ctx, session1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = repo.StoreSession(ctx, session2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	retrieved1, err := repo.GetSessionByToken(ctx, "token1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved1.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", retrieved1.UserID)
	}

	retrieved2, err := repo.GetSessionByToken(ctx, "token2")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved2.UserID != 2 {
		t.Errorf("expected UserID 2, got %d", retrieved2.UserID)
	}
}
