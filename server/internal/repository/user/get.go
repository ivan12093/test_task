package user

import (
	"context"
	"database/sql"
	"fmt"
	"server/internal/domain"
)

const ErrDuplicateEntry = 1062

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	var passwordHash, fullName, phone sql.NullString
	row := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, password_hash, full_name, phone FROM user WHERE email = ?",
		email,
	)
	err := row.Scan(&user.ID, &user.Email, &passwordHash, &fullName, &phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotExists
		}
		r.logger.Error("failed to get user by email", "error", err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Password = passwordHash.String
	user.FullName = fullName.String
	user.Phone = phone.String

	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	var user domain.User
	var passwordHash, fullName, phone sql.NullString
	row := r.db.QueryRowContext(
		ctx,
		"SELECT id, email, password_hash, full_name, phone FROM user WHERE id = ?",
		userID,
	)
	err := row.Scan(&user.ID, &user.Email, &passwordHash, &fullName, &phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotExists
		}
		r.logger.Error("failed to get user by id", "error", err)
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	user.Password = passwordHash.String
	user.FullName = fullName.String
	user.Phone = phone.String

	return &user, nil
}

func (r *Repository) GetProfileByUserID(ctx context.Context, userID int64) (*domain.Profile, error) {
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &domain.Profile{
		UserID:   user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Phone:    user.Phone,
	}, nil
}

func (r *Repository) GetUserByOAuthInfo(ctx context.Context, oauthInfo *domain.OAuthUserInfo) (*domain.User, error) {
	var user domain.User
	var fullName, phone sql.NullString
	row := r.db.QueryRowContext(
		ctx,
		`SELECT u.id, u.email, u.full_name, u.phone 
		FROM oauth_account oa
		JOIN user u ON oa.user_id = u.id
		WHERE oa.provider_name = ? AND oa.sub = ?`,
		oauthInfo.ProviderName, oauthInfo.Sub,
	)
	err := row.Scan(&user.ID, &user.Email, &fullName, &phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotExists
		}
		r.logger.Error("failed to get user by oauth info", "error", err)
		return nil, fmt.Errorf("failed to get user by oauth info: %w", err)
	}

	user.FullName = fullName.String
	user.Phone = phone.String

	return &user, nil
}
