package user

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"server/internal/domain"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	return db, mock
}

func TestRepository_CreateUserWithCredentials(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		credentials   domain.Credentials
		setupMock     func(sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name: "successful create",
			credentials: domain.Credentials{
				Email:    "test@example.com",
				Password: "hashed_password",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO user").
					WithArgs("test@example.com", "hashed_password").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: nil,
		},
		{
			name: "duplicate entry",
			credentials: domain.Credentials{
				Email:    "test@example.com",
				Password: "hashed_password",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				mysqlErr := &mysql.MySQLError{
					Number:  ErrDuplicateEntry,
					Message: "Duplicate entry",
				}
				m.ExpectExec("INSERT INTO user").
					WithArgs("test@example.com", "hashed_password").
					WillReturnError(mysqlErr)
			},
			expectedError: domain.ErrUserAlreadyExists,
		},
		{
			name: "database error",
			credentials: domain.Credentials{
				Email:    "test@example.com",
				Password: "hashed_password",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec("INSERT INTO user").
					WithArgs("test@example.com", "hashed_password").
					WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewRepository(logger, db)
			err := repo.CreateUserWithCredentials(ctx, tt.credentials)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if err != tt.expectedError && !isMySQLError(err, ErrDuplicateEntry) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations were not met: %v", err)
			}
		})
	}
}

func TestRepository_GetUserByEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		email         string
		setupMock     func(sqlmock.Sqlmock)
		expectedError error
		expectedUser  *domain.User
	}{
		{
			name:  "successful get",
			email: "test@example.com",
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "full_name", "phone"}).
					AddRow(1, "test@example.com", "hashed_password", "Test User", "1234567890")
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE email").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedUser: &domain.User{
				ID:       1,
				Email:    "test@example.com",
				Password: "hashed_password",
				FullName: "Test User",
				Phone:    "1234567890",
			},
		},
		{
			name:  "user not found",
			email: "test@example.com",
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE email").
					WithArgs("test@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: domain.ErrUserNotExists,
			expectedUser:  nil,
		},
		{
			name:  "null fields",
			email: "test@example.com",
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "full_name", "phone"}).
					AddRow(1, "test@example.com", nil, nil, nil)
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE email").
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedUser: &domain.User{
				ID:       1,
				Email:    "test@example.com",
				Password: "",
				FullName: "",
				Phone:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewRepository(logger, db)
			user, err := repo.GetUserByEmail(ctx, tt.email)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				if user != nil {
					t.Errorf("expected nil user, got %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil {
					t.Error("expected user, got nil")
				} else {
					if user.ID != tt.expectedUser.ID {
						t.Errorf("expected ID %d, got %d", tt.expectedUser.ID, user.ID)
					}
					if user.Email != tt.expectedUser.Email {
						t.Errorf("expected Email %s, got %s", tt.expectedUser.Email, user.Email)
					}
					if user.Password != tt.expectedUser.Password {
						t.Errorf("expected Password %s, got %s", tt.expectedUser.Password, user.Password)
					}
					if user.FullName != tt.expectedUser.FullName {
						t.Errorf("expected FullName %s, got %s", tt.expectedUser.FullName, user.FullName)
					}
					if user.Phone != tt.expectedUser.Phone {
						t.Errorf("expected Phone %s, got %s", tt.expectedUser.Phone, user.Phone)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations were not met: %v", err)
			}
		})
	}
}

func TestRepository_GetUserByID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		userID        int64
		setupMock     func(sqlmock.Sqlmock)
		expectedError error
		expectedUser  *domain.User
	}{
		{
			name:   "successful get",
			userID: 1,
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "full_name", "phone"}).
					AddRow(1, "test@example.com", "hashed_password", "Test User", "1234567890")
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedUser: &domain.User{
				ID:       1,
				Email:    "test@example.com",
				Password: "hashed_password",
				FullName: "Test User",
				Phone:    "1234567890",
			},
		},
		{
			name:   "user not found",
			userID: 1,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE id").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: domain.ErrUserNotExists,
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewRepository(logger, db)
			user, err := repo.GetUserByID(ctx, tt.userID)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				if user != nil {
					t.Errorf("expected nil user, got %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil {
					t.Error("expected user, got nil")
				} else {
					if user.ID != tt.expectedUser.ID {
						t.Errorf("expected ID %d, got %d", tt.expectedUser.ID, user.ID)
					}
					if user.Email != tt.expectedUser.Email {
						t.Errorf("expected Email %s, got %s", tt.expectedUser.Email, user.Email)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations were not met: %v", err)
			}
		})
	}
}

func TestRepository_GetProfileByUserID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name            string
		userID          int64
		setupMock       func(sqlmock.Sqlmock)
		expectedError   error
		expectedProfile *domain.Profile
	}{
		{
			name:   "successful get profile",
			userID: 1,
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "full_name", "phone"}).
					AddRow(1, "test@example.com", "hashed_password", "Test User", "1234567890")
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE id").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedProfile: &domain.Profile{
				UserID:   1,
				Email:    "test@example.com",
				FullName: "Test User",
				Phone:    "1234567890",
			},
		},
		{
			name:   "user not found",
			userID: 1,
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT id, email, password_hash, full_name, phone FROM user WHERE id").
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError:   domain.ErrUserNotExists,
			expectedProfile: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewRepository(logger, db)
			profile, err := repo.GetProfileByUserID(ctx, tt.userID)

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
					if profile.UserID != tt.expectedProfile.UserID {
						t.Errorf("expected UserID %d, got %d", tt.expectedProfile.UserID, profile.UserID)
					}
					if profile.Email != tt.expectedProfile.Email {
						t.Errorf("expected Email %s, got %s", tt.expectedProfile.Email, profile.Email)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations were not met: %v", err)
			}
		})
	}
}

func TestRepository_GetUserByOAuthInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		oauthInfo     *domain.OAuthUserInfo
		setupMock     func(sqlmock.Sqlmock)
		expectedError error
		expectedUser  *domain.User
	}{
		{
			name: "successful get",
			oauthInfo: &domain.OAuthUserInfo{
				ProviderName: "google",
				Sub:          "123456",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"u.id", "u.email", "u.full_name", "u.phone"}).
					AddRow(1, "test@example.com", "Test User", "1234567890")
				m.ExpectQuery("SELECT u.id, u.email, u.full_name, u.phone").
					WithArgs("google", "123456").
					WillReturnRows(rows)
			},
			expectedError: nil,
			expectedUser: &domain.User{
				ID:       1,
				Email:    "test@example.com",
				FullName: "Test User",
				Phone:    "1234567890",
			},
		},
		{
			name: "user not found",
			oauthInfo: &domain.OAuthUserInfo{
				ProviderName: "google",
				Sub:          "123456",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT u.id, u.email, u.full_name, u.phone").
					WithArgs("google", "123456").
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: domain.ErrUserNotExists,
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewRepository(logger, db)
			user, err := repo.GetUserByOAuthInfo(ctx, tt.oauthInfo)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				if user != nil {
					t.Errorf("expected nil user, got %v", user)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if user == nil {
					t.Error("expected user, got nil")
				} else {
					if user.ID != tt.expectedUser.ID {
						t.Errorf("expected ID %d, got %d", tt.expectedUser.ID, user.ID)
					}
					if user.Email != tt.expectedUser.Email {
						t.Errorf("expected Email %s, got %s", tt.expectedUser.Email, user.Email)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations were not met: %v", err)
			}
		})
	}
}

func TestRepository_UpdateProfile(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	tests := []struct {
		name          string
		profile       *domain.Profile
		setupMock     func(sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name: "successful update",
			profile: &domain.Profile{
				UserID:   1,
				Email:    "updated@example.com",
				FullName: "Updated User",
				Phone:    "9876543210",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec("UPDATE user SET email = \\?, full_name = \\?, phone = \\? WHERE id = \\?").
					WithArgs("updated@example.com", "Updated User", "9876543210", int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: nil,
		},
		{
			name: "update with empty fields",
			profile: &domain.Profile{
				UserID:   1,
				Email:    "updated@example.com",
				FullName: "",
				Phone:    "",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectExec("UPDATE user SET email = \\?, full_name = \\?, phone = \\? WHERE id = \\?").
					WithArgs("updated@example.com", nil, nil, int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedError: nil,
		},
		{
			name: "duplicate email",
			profile: &domain.Profile{
				UserID:   1,
				Email:    "existing@example.com",
				FullName: "Updated User",
				Phone:    "9876543210",
			},
			setupMock: func(m sqlmock.Sqlmock) {
				mysqlErr := &mysql.MySQLError{
					Number:  ErrDuplicateEntry,
					Message: "Duplicate entry",
				}
				m.ExpectExec("UPDATE user SET email = \\?, full_name = \\?, phone = \\? WHERE id = \\?").
					WithArgs("existing@example.com", "Updated User", "9876543210", int64(1)).
					WillReturnError(mysqlErr)
			},
			expectedError: domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			defer db.Close()

			tt.setupMock(mock)

			repo := NewRepository(logger, db)
			err := repo.UpdateProfile(ctx, tt.profile)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
				} else if err != tt.expectedError && !isMySQLError(err, ErrDuplicateEntry) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("mock expectations were not met: %v", err)
			}
		})
	}
}

func isMySQLError(err error, number uint16) bool {
	if err != nil && err.Error() == domain.ErrUserAlreadyExists.Error() {
		return true
	}
	return false
}
