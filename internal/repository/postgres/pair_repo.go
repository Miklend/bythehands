package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"taalkbout/internal/domain/pair"
	"taalkbout/internal/repository"
)

type PairRepo struct {
	pool *pgxpool.Pool
}

func NewPairRepo(pool *pgxpool.Pool) *PairRepo {
	return &PairRepo{pool: pool}
}

func (r *PairRepo) CreatePair(ctx context.Context, isTest bool) (pair.Pair, error) {
	const q = `
INSERT INTO pairs (status, is_test)
VALUES ('active', $1)
RETURNING id, status, is_test, welcome_message, created_at, updated_at;
`
	var p pair.Pair
	err := r.pool.QueryRow(ctx, q, isTest).Scan(&p.ID, &p.Status, &p.IsTest, &p.WelcomeMessage, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return pair.Pair{}, err
	}
	return p, nil
}

func (r *PairRepo) AddMember(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
	const q = `
INSERT INTO pair_members (pair_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING id, pair_id, user_id, role, custom_name, created_at;
`
	var m pair.PairMember
	err := r.pool.QueryRow(ctx, q, pairID, userID, role).
		Scan(&m.ID, &m.PairID, &m.UserID, &m.Role, &m.CustomName, &m.CreatedAt)
	if isUniqueViolation(err) {
		return pair.PairMember{}, repository.ErrConflict
	}
	if err != nil {
		return pair.PairMember{}, err
	}
	return m, nil
}

func (r *PairRepo) GetPair(ctx context.Context, pairID string) (pair.Pair, error) {
	const q = `
SELECT id, status, is_test, welcome_message, created_at, updated_at
FROM pairs
WHERE id = $1;
`
	var p pair.Pair
	err := r.pool.QueryRow(ctx, q, pairID).Scan(&p.ID, &p.Status, &p.IsTest, &p.WelcomeMessage, &p.CreatedAt, &p.UpdatedAt)
	if isNoRows(err) {
		return pair.Pair{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.Pair{}, err
	}
	return p, nil
}

func (r *PairRepo) GetMembers(ctx context.Context, pairID string) ([]pair.PairMember, error) {
	const q = `
SELECT id, pair_id, user_id, role, custom_name, created_at
FROM pair_members
WHERE pair_id = $1
ORDER BY created_at ASC;
`
	rows, err := r.pool.Query(ctx, q, pairID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []pair.PairMember
	for rows.Next() {
		var m pair.PairMember
		if err := rows.Scan(&m.ID, &m.PairID, &m.UserID, &m.Role, &m.CustomName, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *PairRepo) ListPairsByUser(ctx context.Context, userID string) ([]pair.Pair, error) {
	const q = `
SELECT p.id, p.status, p.is_test, p.welcome_message, p.created_at, p.updated_at
FROM pairs p
JOIN pair_members pm ON pm.pair_id = p.id
WHERE pm.user_id = $1
ORDER BY p.created_at DESC;
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []pair.Pair
	for rows.Next() {
		var p pair.Pair
		if err := rows.Scan(&p.ID, &p.Status, &p.IsTest, &p.WelcomeMessage, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *PairRepo) ArchivePair(ctx context.Context, pairID string) (pair.Pair, error) {
	const q = `
UPDATE pairs
SET status = 'archived', updated_at = NOW()
WHERE id = $1
RETURNING id, status, is_test, welcome_message, created_at, updated_at;
`
	var p pair.Pair
	err := r.pool.QueryRow(ctx, q, pairID).Scan(&p.ID, &p.Status, &p.IsTest, &p.WelcomeMessage, &p.CreatedAt, &p.UpdatedAt)
	if isNoRows(err) {
		return pair.Pair{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.Pair{}, err
	}
	return p, nil
}

func (r *PairRepo) SetWelcomeMessage(ctx context.Context, pairID string, text *string) (pair.Pair, error) {
	const q = `
UPDATE pairs
SET welcome_message = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, status, is_test, welcome_message, created_at, updated_at;
`
	var p pair.Pair
	err := r.pool.QueryRow(ctx, q, pairID, text).Scan(&p.ID, &p.Status, &p.IsTest, &p.WelcomeMessage, &p.CreatedAt, &p.UpdatedAt)
	if isNoRows(err) {
		return pair.Pair{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.Pair{}, err
	}
	return p, nil
}

func (r *PairRepo) SetMemberName(ctx context.Context, pairID, userID string, name *string) (pair.PairMember, error) {
	const q = `
UPDATE pair_members
SET custom_name = $3
WHERE pair_id = $1 AND user_id = $2
RETURNING id, pair_id, user_id, role, custom_name, created_at;
`
	var m pair.PairMember
	err := r.pool.QueryRow(ctx, q, pairID, userID, name).
		Scan(&m.ID, &m.PairID, &m.UserID, &m.Role, &m.CustomName, &m.CreatedAt)
	if isNoRows(err) {
		return pair.PairMember{}, repository.ErrNotFound
	}
	if err != nil {
		return pair.PairMember{}, err
	}
	return m, nil
}
