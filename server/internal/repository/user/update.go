package user

import (
	"context"
	"errors"
	"server/internal/domain"

	"github.com/go-sql-driver/mysql"
)

func (r *Repository) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	// Преобразуем пустые строки в NULL для nullable полей
	var fullName, phone interface{}
	if profile.FullName == "" {
		fullName = nil
	} else {
		fullName = profile.FullName
	}
	if profile.Phone == "" {
		phone = nil
	} else {
		phone = profile.Phone
	}
	
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE user SET email = ?, full_name = ?, phone = ? WHERE id = ?",
		profile.Email, fullName, phone, profile.UserID,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == ErrDuplicateEntry {
				return domain.ErrUserAlreadyExists
			}
		}
		r.logger.Error("failed to update profile", "error", err)
		return err
	}
	return nil
}
