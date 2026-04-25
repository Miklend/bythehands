package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taalkbout/internal/domain/pair"
	"taalkbout/internal/repository"
)

type InviteRepo struct {
	pool *pgxpool.Pool
}

func NewInviteRepo(pool *pgxpool.Pool) *InviteRepo {
	return &InviteRepo{pool: pool}
}

func (r *InviteRepo) CreateInvite(ctx context.Context, pairID, token string, expiresAt time.Time) (pair.Invite, error) {
	const q = `
INSERT INTO invites (pair_id, token, status, expires_at)
VALUES ($1, $2, 'active', $3)
RETURNING id, pair_id, token, status, expires_at, created_at, used_at;
`
	var inv pair.Invite
	err := r.pool.QueryRow(ctx, q, pairID, token, expiresAt).
		Scan(&inv.ID, &inv.PairID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UsedAt)
	if isUniqueViolation(err) {
		return pair.Invite{}, repository.ErrConflict
	}
	if err != nil {
		return pair.Invite{}, err
	}
	return inv, nil
}

func (r *InviteRepo) GetByToken(ctx context.Context, token string) (pair.Invite, error) {
	const q = `
SELECT id, pair_id, token, status, expires_at, created_at, used_at
FROM invites
WHERE token = $1;
`
	var inv pair.Invite
	err := r.pool.QueryRow(ctx, q, token).
		Scan(&inv.ID, &inv.PairID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UsedAt)
	if isNoRows(err) {
		return pair.Invite{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.Invite{}, err
	}
	return inv, nil
}

func (r *InviteRepo) MarkUsed(ctx context.Context, inviteID string, usedAt time.Time) (pair.Invite, error) {
	const q = `
UPDATE invites
SET status = 'used', used_at = $2
WHERE id = $1
RETURNING id, pair_id, token, status, expires_at, created_at, used_at;
`
	var inv pair.Invite
	err := r.pool.QueryRow(ctx, q, inviteID, usedAt).
		Scan(&inv.ID, &inv.PairID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UsedAt)
	if isNoRows(err) {
		return pair.Invite{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.Invite{}, err
	}
	return inv, nil
}

func (r *InviteRepo) MarkExpired(ctx context.Context, inviteID string) (pair.Invite, error) {
	const q = `
UPDATE invites
SET status = 'expired'
WHERE id = $1
RETURNING id, pair_id, token, status, expires_at, created_at, used_at;
`
	var inv pair.Invite
	err := r.pool.QueryRow(ctx, q, inviteID).
		Scan(&inv.ID, &inv.PairID, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt, &inv.UsedAt)
	if isNoRows(err) {
		return pair.Invite{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.Invite{}, err
	}
	return inv, nil
}
