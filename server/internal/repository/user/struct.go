package user

import (
	"database/sql"
	"log/slog"
)

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(logger *slog.Logger, db *sql.DB) *Repository {
	return &Repository{logger: logger, db: db}
}
