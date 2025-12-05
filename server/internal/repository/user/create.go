package user

import (
	"context"
	"errors"
	"fmt"
	"server/internal/domain"

	"github.com/go-sql-driver/mysql"
)

func (r *Repository) CreateUserWithCredentials(ctx context.Context, credentials domain.Credentials) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO user (email, password_hash) VALUES (?, ?)", credentials.Email, credentials.Password)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == ErrDuplicateEntry {
				return domain.ErrUserAlreadyExists
			}
		}
		r.logger.Error("failed to create user with credentials", "error", err)
		return err
	}
	return nil
}

func (r *Repository) CreateUserWithOAuthInfo(ctx context.Context, oauthInfo *domain.OAuthUserInfo) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				r.logger.Error("failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	result, err := tx.ExecContext(
		ctx,
		"INSERT INTO user (email) VALUES (?)",
		oauthInfo.Email,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == ErrDuplicateEntry {
				return domain.ErrUserAlreadyExists
			}
		}
		r.logger.Error("failed to create user with oauth info", "error", err)
		return fmt.Errorf("failed to create user with oauth info: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("failed to get last insert id", "error", err)
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO oauth_account (user_id, provider_name, sub) VALUES (?, ?, ?)",
		userID, oauthInfo.ProviderName, oauthInfo.Sub,
	)
	if err != nil {
		r.logger.Error("failed to create oauth account", "error", err)
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true
	return nil
}
