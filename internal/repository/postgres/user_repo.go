package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"taalkbout/internal/domain/user"
	"taalkbout/internal/repository"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) UpsertTelegramUser(ctx context.Context, telegramID int64, username, displayName *string) (user.User, error) {
	const q = `
INSERT INTO users (telegram_id, username, display_name, is_virtual)
VALUES ($1, $2, $3, false)
ON CONFLICT (telegram_id)
DO UPDATE SET
  username = EXCLUDED.username,
  display_name = EXCLUDED.display_name,
  updated_at = NOW()
RETURNING id, telegram_id, username, display_name, is_virtual, created_at, updated_at;
`
	var u user.User
	err := r.pool.QueryRow(ctx, q, telegramID, username, displayName).
		Scan(&u.ID, &u.TelegramID, &u.Username, &u.DisplayName, &u.IsVirtual, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	const q = `
SELECT id, telegram_id, username, display_name, is_virtual, created_at, updated_at
FROM users
WHERE id = $1;
`
	var u user.User
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.TelegramID, &u.Username, &u.DisplayName, &u.IsVirtual, &u.CreatedAt, &u.UpdatedAt)
	if isNoRows(err) {
		return user.User{}, repository.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}

func (r *UserRepo) CreateVirtualUser(ctx context.Context, username, displayName string) (user.User, error) {
	const q = `
INSERT INTO users (telegram_id, username, display_name, is_virtual)
VALUES (NULL, $1, $2, true)
RETURNING id, telegram_id, username, display_name, is_virtual, created_at, updated_at;
`
	var u user.User
	err := r.pool.QueryRow(ctx, q, username, displayName).
		Scan(&u.ID, &u.TelegramID, &u.Username, &u.DisplayName, &u.IsVirtual, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}
