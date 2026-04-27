package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taalkbout/internal/domain/issue"
	"taalkbout/internal/repository"
)

type IssueRepo struct {
	pool *pgxpool.Pool
}

func NewIssueRepo(pool *pgxpool.Pool) *IssueRepo {
	return &IssueRepo{pool: pool}
}

func (r *IssueRepo) CreateIssue(ctx context.Context, in issue.Issue) (issue.Issue, error) {
	const q = `
INSERT INTO issues (
  pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status
)
VALUES ($1,$2,$3,$4,$5,$6,$7,0,$8,NULL,'active')
RETURNING
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at;
`
	var out issue.Issue
	err := r.pool.QueryRow(
		ctx,
		q,
		in.PairID,
		in.CreatedByUserID,
		in.Title,
		in.Description,
		in.Priority,
		in.Visibility,
		in.RepeatThreshold,
		in.RepeatLimit,
	).Scan(
		&out.ID,
		&out.PairID,
		&out.CreatedByUserID,
		&out.Title,
		&out.Description,
		&out.Priority,
		&out.Visibility,
		&out.RepeatThreshold,
		&out.RepeatCount,
		&out.RepeatLimit,
		&out.LastRepeatedAt,
		&out.Status,
		&out.CreatedAt,
		&out.UpdatedAt,
		&out.ResolvedAt,
	)
	if err != nil {
		return issue.Issue{}, err
	}
	return out, nil
}

func (r *IssueRepo) ListIssuesByPair(ctx context.Context, pairID string, status *issue.Status) ([]issue.Issue, error) {
	q := `
SELECT
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at
FROM issues
WHERE pair_id = $1
`
	args := []any{pairID}
	if status != nil {
		q += " AND status = $2"
		args = append(args, *status)
	}
	q += " ORDER BY created_at DESC;"

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []issue.Issue
	for rows.Next() {
		var it issue.Issue
		if err := rows.Scan(
			&it.ID,
			&it.PairID,
			&it.CreatedByUserID,
			&it.Title,
			&it.Description,
			&it.Priority,
			&it.Visibility,
			&it.RepeatThreshold,
			&it.RepeatCount,
			&it.RepeatLimit,
			&it.LastRepeatedAt,
			&it.Status,
			&it.CreatedAt,
			&it.UpdatedAt,
			&it.ResolvedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *IssueRepo) GetIssue(ctx context.Context, issueID string) (issue.Issue, error) {
	const q = `
SELECT
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at
FROM issues
WHERE id = $1;
`
	var it issue.Issue
	err := r.pool.QueryRow(ctx, q, issueID).Scan(
		&it.ID,
		&it.PairID,
		&it.CreatedByUserID,
		&it.Title,
		&it.Description,
		&it.Priority,
		&it.Visibility,
		&it.RepeatThreshold,
		&it.RepeatCount,
		&it.RepeatLimit,
		&it.LastRepeatedAt,
		&it.Status,
		&it.CreatedAt,
		&it.UpdatedAt,
		&it.ResolvedAt,
	)
	if isNoRows(err) {
		return issue.Issue{}, repository.ErrNotFound
	}
	if err != nil {
		return issue.Issue{}, err
	}
	return it, nil
}

func (r *IssueRepo) RepeatIssue(ctx context.Context, issueID, userID string, note *string) (issue.Issue, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return issue.Issue{}, err
	}
	defer tx.Rollback(ctx)

	const insertRepeat = `
INSERT INTO issue_repeats (issue_id, user_id, note)
VALUES ($1, $2, $3);
`
	if _, err := tx.Exec(ctx, insertRepeat, issueID, userID, note); err != nil {
		return issue.Issue{}, err
	}

	const bumpIssue = `
UPDATE issues
SET repeat_count = repeat_count + 1, updated_at = NOW(), last_repeated_at = NOW()
WHERE id = $1 AND (repeat_limit = 0 OR repeat_count < repeat_limit)
RETURNING
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at;
`
	var it issue.Issue
	err = tx.QueryRow(ctx, bumpIssue, issueID).Scan(
		&it.ID,
		&it.PairID,
		&it.CreatedByUserID,
		&it.Title,
		&it.Description,
		&it.Priority,
		&it.Visibility,
		&it.RepeatThreshold,
		&it.RepeatCount,
		&it.RepeatLimit,
		&it.LastRepeatedAt,
		&it.Status,
		&it.CreatedAt,
		&it.UpdatedAt,
		&it.ResolvedAt,
	)
	if isNoRows(err) {
		const existsQ = `SELECT 1 FROM issues WHERE id = $1;`
		var one int
		exErr := tx.QueryRow(ctx, existsQ, issueID).Scan(&one)
		if isNoRows(exErr) {
			return issue.Issue{}, repository.ErrNotFound
		}
		if exErr != nil {
			return issue.Issue{}, exErr
		}
		return issue.Issue{}, repository.ErrConflict
	}
	if err != nil {
		return issue.Issue{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return issue.Issue{}, err
	}
	return it, nil
}

func (r *IssueRepo) IncrementRepeat(ctx context.Context, issueID string) (issue.Issue, error) {
	const q = `
UPDATE issues
SET repeat_count = repeat_count + 1, updated_at = NOW(), last_repeated_at = NOW()
WHERE id = $1
RETURNING
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at;
`
	var it issue.Issue
	err := r.pool.QueryRow(ctx, q, issueID).Scan(
		&it.ID,
		&it.PairID,
		&it.CreatedByUserID,
		&it.Title,
		&it.Description,
		&it.Priority,
		&it.Visibility,
		&it.RepeatThreshold,
		&it.RepeatCount,
		&it.RepeatLimit,
		&it.LastRepeatedAt,
		&it.Status,
		&it.CreatedAt,
		&it.UpdatedAt,
		&it.ResolvedAt,
	)
	if isNoRows(err) {
		return issue.Issue{}, repository.ErrNotFound
	}
	if err != nil {
		return issue.Issue{}, err
	}
	return it, nil
}

func (r *IssueRepo) CreateRepeat(ctx context.Context, issueID, userID string, note *string) (issue.IssueRepeat, error) {
	const q = `
INSERT INTO issue_repeats (issue_id, user_id, note)
VALUES ($1, $2, $3)
RETURNING id, issue_id, user_id, note, created_at;
`
	var rep issue.IssueRepeat
	err := r.pool.QueryRow(ctx, q, issueID, userID, note).
		Scan(&rep.ID, &rep.IssueID, &rep.UserID, &rep.Note, &rep.CreatedAt)
	if err != nil {
		return issue.IssueRepeat{}, err
	}
	return rep, nil
}

func (r *IssueRepo) ListRepeatsByIssue(ctx context.Context, issueID string, limit, offset int) ([]issue.IssueRepeat, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	const q = `
SELECT id, issue_id, user_id, note, created_at
FROM issue_repeats
WHERE issue_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
`
	rows, err := r.pool.Query(ctx, q, issueID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []issue.IssueRepeat
	for rows.Next() {
		var rep issue.IssueRepeat
		if err := rows.Scan(&rep.ID, &rep.IssueID, &rep.UserID, &rep.Note, &rep.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rep)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *IssueRepo) GetRepeat(ctx context.Context, repeatID string) (issue.IssueRepeat, error) {
	const q = `
SELECT id, issue_id, user_id, note, created_at
FROM issue_repeats
WHERE id = $1;
`
	var rep issue.IssueRepeat
	err := r.pool.QueryRow(ctx, q, repeatID).
		Scan(&rep.ID, &rep.IssueID, &rep.UserID, &rep.Note, &rep.CreatedAt)
	if isNoRows(err) {
		return issue.IssueRepeat{}, repository.ErrNotFound
	}
	if err != nil {
		return issue.IssueRepeat{}, err
	}
	return rep, nil
}

func (r *IssueRepo) CreateRepeatDisagreement(ctx context.Context, repeatID, userID string, note string) (issue.IssueRepeatDisagreement, error) {
	const q = `
INSERT INTO issue_repeat_disagreements (repeat_id, user_id, note)
VALUES ($1, $2, $3)
RETURNING id, repeat_id, user_id, note, created_at;
`
	var d issue.IssueRepeatDisagreement
	err := r.pool.QueryRow(ctx, q, repeatID, userID, note).
		Scan(&d.ID, &d.RepeatID, &d.UserID, &d.Note, &d.CreatedAt)
	if isUniqueViolation(err) {
		return issue.IssueRepeatDisagreement{}, repository.ErrConflict
	}
	if err != nil {
		return issue.IssueRepeatDisagreement{}, err
	}
	return d, nil
}

func (r *IssueRepo) GetRepeatDisagreement(ctx context.Context, repeatID, userID string) (issue.IssueRepeatDisagreement, error) {
	const q = `
SELECT id, repeat_id, user_id, note, created_at
FROM issue_repeat_disagreements
WHERE repeat_id = $1 AND user_id = $2;
`
	var d issue.IssueRepeatDisagreement
	err := r.pool.QueryRow(ctx, q, repeatID, userID).
		Scan(&d.ID, &d.RepeatID, &d.UserID, &d.Note, &d.CreatedAt)
	if isNoRows(err) {
		return issue.IssueRepeatDisagreement{}, repository.ErrNotFound
	}
	if err != nil {
		return issue.IssueRepeatDisagreement{}, err
	}
	return d, nil
}

func (r *IssueRepo) UpdateStatus(ctx context.Context, issueID string, status issue.Status) (issue.Issue, error) {
	const q = `
UPDATE issues
SET
  status = $2,
  updated_at = NOW(),
  resolved_at = CASE WHEN $2 = 'resolved' THEN COALESCE(resolved_at, NOW()) ELSE NULL END
WHERE id = $1
RETURNING
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at;
`
	var it issue.Issue
	err := r.pool.QueryRow(ctx, q, issueID, status).Scan(
		&it.ID,
		&it.PairID,
		&it.CreatedByUserID,
		&it.Title,
		&it.Description,
		&it.Priority,
		&it.Visibility,
		&it.RepeatThreshold,
		&it.RepeatCount,
		&it.RepeatLimit,
		&it.LastRepeatedAt,
		&it.Status,
		&it.CreatedAt,
		&it.UpdatedAt,
		&it.ResolvedAt,
	)
	if isNoRows(err) {
		return issue.Issue{}, repository.ErrNotFound
	}
	if err != nil {
		return issue.Issue{}, err
	}
	return it, nil
}

func (r *IssueRepo) UpdateIssue(ctx context.Context, issueID string, title *string, repeatThreshold *int, repeatLimit *int) (issue.Issue, error) {
	const q = `
UPDATE issues
SET
  title = COALESCE($2::text, title),
  repeat_threshold = COALESCE($3::int, repeat_threshold),
  repeat_limit = COALESCE($4::int, repeat_limit),
  updated_at = NOW()
WHERE id = $1
RETURNING
  id, pair_id, created_by_user_id, title, description, priority, visibility,
  repeat_threshold, repeat_count, repeat_limit, last_repeated_at, status, created_at, updated_at, resolved_at;
`
	var it issue.Issue
	err := r.pool.QueryRow(ctx, q, issueID, title, repeatThreshold, repeatLimit).Scan(
		&it.ID,
		&it.PairID,
		&it.CreatedByUserID,
		&it.Title,
		&it.Description,
		&it.Priority,
		&it.Visibility,
		&it.RepeatThreshold,
		&it.RepeatCount,
		&it.RepeatLimit,
		&it.LastRepeatedAt,
		&it.Status,
		&it.CreatedAt,
		&it.UpdatedAt,
		&it.ResolvedAt,
	)
	if isNoRows(err) {
		return issue.Issue{}, repository.ErrNotFound
	}
	if err != nil {
		return issue.Issue{}, err
	}
	return it, nil
}

func (r *IssueRepo) DeleteIssue(ctx context.Context, issueID string) error {
	const q = `DELETE FROM issues WHERE id = $1;`
	ct, err := r.pool.Exec(ctx, q, issueID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *IssueRepo) DeleteRepeat(ctx context.Context, repeatID string) error {
	const q = `DELETE FROM issue_repeats WHERE id = $1;`
	ct, err := r.pool.Exec(ctx, q, repeatID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Convenience helper for service tests; not part of interface.
func nowUTC() time.Time { return time.Now().UTC() }
