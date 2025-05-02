package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

var (
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				if strings.Contains(pgErr.Message, "email") {
					return ErrEmailAlreadyExists
				}
				if strings.Contains(pgErr.Message, "username") {
					return ErrUsernameAlreadyExists
				}
			}
		}
	}

	return err
}
