package profile

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"server/internal/domain"
	"testing"
)

type mockProfileRepository struct {
	getProfileByUserIDFunc func(ctx context.Context, userID int64) (*domain.Profile, error)
	updateProfileFunc      func(ctx context.Context, profile *domain.Profile) error
}

func (m *mockProfileRepository) GetProfileByUserID(ctx context.Context, userID int64) (*domain.Profile, error) {
	if m.getProfileByUserIDFunc != nil {
		return m.getProfileByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockProfileRepository) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, profile)
	}
	return nil
}

func TestUseCase_GetProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int64
		setupMocks    func(*mockProfileRepository)
		expectedError error
		expectedEmail string
	}{
		{
			name:   "successful get profile",
			userID: 1,
			setupMocks: func(m *mockProfileRepository) {
				m.getProfileByUserIDFunc = func(ctx context.Context, userID int64) (*domain.Profile, error) {
					if userID != 1 {
						t.Errorf("expected userID 1, got %d", userID)
					}
					return &domain.Profile{
						UserID:   1,
						Email:    "test@example.com",
						FullName: "Test User",
						Phone:    "1234567890",
					}, nil
				}
			},
			expectedError: nil,
			expectedEmail: "test@example.com",
		},
		{
			name:   "profile not found",
			userID: 1,
			setupMocks: func(m *mockProfileRepository) {
				m.getProfileByUserIDFunc = func(ctx context.Context, userID int64) (*domain.Profile, error) {
					return nil, errors.New("profile not found")
				}
			},
			expectedError: errors.New("profile not found"),
			expectedEmail: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProfileRepo := &mockProfileRepository{}
			tt.setupMocks(mockProfileRepo)

			uc := NewUseCase(logger, mockProfileRepo)
			profile, err := uc.GetProfile(ctx, tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
				if profile != nil {
					t.Errorf("expected nil profile, got %v", profile)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if profile == nil {
					t.Error("expected profile, got nil")
				} else {
					if profile.Email != tt.expectedEmail {
						t.Errorf("expected email %s, got %s", tt.expectedEmail, profile.Email)
					}
					if profile.UserID != tt.userID {
						t.Errorf("expected userID %d, got %d", tt.userID, profile.UserID)
					}
				}
			}
		})
	}
}

func TestUseCase_UpdateProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int64
		profile       *domain.Profile
		setupMocks    func(*mockProfileRepository)
		expectedError error
	}{
		{
			name:   "successful update",
			userID: 1,
			profile: &domain.Profile{
				Email:    "updated@example.com",
				FullName: "Updated User",
				Phone:    "9876543210",
			},
			setupMocks: func(m *mockProfileRepository) {
				m.updateProfileFunc = func(ctx context.Context, profile *domain.Profile) error {
					if profile.UserID != 1 {
						t.Errorf("expected userID 1, got %d", profile.UserID)
					}
					if profile.Email != "updated@example.com" {
						t.Errorf("expected email updated@example.com, got %s", profile.Email)
					}
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:   "update error",
			userID: 1,
			profile: &domain.Profile{
				Email:    "updated@example.com",
				FullName: "Updated User",
				Phone:    "9876543210",
			},
			setupMocks: func(m *mockProfileRepository) {
				m.updateProfileFunc = func(ctx context.Context, profile *domain.Profile) error {
					return errors.New("update error")
				}
			},
			expectedError: errors.New("update error"),
		},
		{
			name:   "update with empty fields",
			userID: 1,
			profile: &domain.Profile{
				Email:    "updated@example.com",
				FullName: "",
				Phone:    "",
			},
			setupMocks: func(m *mockProfileRepository) {
				m.updateProfileFunc = func(ctx context.Context, profile *domain.Profile) error {
					if profile.FullName != "" {
						t.Errorf("expected empty FullName, got %s", profile.FullName)
					}
					if profile.Phone != "" {
						t.Errorf("expected empty Phone, got %s", profile.Phone)
					}
					return nil
				}
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProfileRepo := &mockProfileRepository{}
			tt.setupMocks(mockProfileRepo)

			uc := NewUseCase(logger, mockProfileRepo)
			err := uc.UpdateProfile(ctx, tt.userID, tt.profile)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
