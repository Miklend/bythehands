package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"taalkbout/internal/repository"
)

type PreferencesRepo struct {
	pool *pgxpool.Pool
}

func NewPreferencesRepo(pool *pgxpool.Pool) *PreferencesRepo {
	return &PreferencesRepo{pool: pool}
}

func (r *PreferencesRepo) GetPreferences(ctx context.Context, userID string) (*string, error) {
	const q = `
SELECT current_pair_id
FROM user_preferences
WHERE user_id = $1;
`
	var current *string
	err := r.pool.QueryRow(ctx, q, userID).Scan(&current)
	if isNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return current, nil
}

func (r *PreferencesRepo) SetCurrentPair(ctx context.Context, userID string, pairID *string) error {
	const q = `
INSERT INTO user_preferences (user_id, current_pair_id)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET current_pair_id = EXCLUDED.current_pair_id, updated_at = NOW();
`
	ct, err := r.pool.Exec(ctx, q, userID, pairID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
