package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taalkbout/internal/domain/conversation"
	"taalkbout/internal/domain/issue"
	"taalkbout/internal/repository"
)

type ConversationRepo struct {
	pool *pgxpool.Pool
}

func NewConversationRepo(pool *pgxpool.Pool) *ConversationRepo {
	return &ConversationRepo{pool: pool}
}

func (r *ConversationRepo) StartSession(ctx context.Context, in conversation.Session) (conversation.Session, error) {
	const q = `
INSERT INTO conversation_sessions (issue_id, pair_id, status, goal, questions, start_state)
VALUES ($1, $2, 'started', $3, $4, $5)
RETURNING id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at;
`
	var out conversation.Session
	err := r.pool.QueryRow(ctx, q, in.IssueID, in.PairID, in.Goal, in.Questions, in.StartState).
		Scan(&out.ID, &out.IssueID, &out.PairID, &out.Status, &out.Goal, &out.Questions, &out.StartState, &out.EndState, &out.ResultStatus, &out.ResultText, &out.CreatedAt, &out.FinishedAt)
	if err != nil {
		return conversation.Session{}, err
	}
	return out, nil
}

func (r *ConversationRepo) FinishSession(ctx context.Context, id string, resultStatus conversation.ResultStatus, resultText *string, endState *string, finishedAt time.Time) (conversation.Session, error) {
	const q = `
UPDATE conversation_sessions
SET status = 'finished', result_status = $2, result_text = $3, end_state = $4, finished_at = $5
WHERE id = $1
RETURNING id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at;
`
	var out conversation.Session
	err := r.pool.QueryRow(ctx, q, id, resultStatus, resultText, endState, finishedAt).
		Scan(&out.ID, &out.IssueID, &out.PairID, &out.Status, &out.Goal, &out.Questions, &out.StartState, &out.EndState, &out.ResultStatus, &out.ResultText, &out.CreatedAt, &out.FinishedAt)
	if isNoRows(err) {
		return conversation.Session{}, repository.ErrNotFound
	}
	if err != nil {
		return conversation.Session{}, err
	}
	return out, nil
}

func (r *ConversationRepo) GetSession(ctx context.Context, id string) (conversation.Session, error) {
	const q = `
SELECT id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at
FROM conversation_sessions
WHERE id = $1;
`
	var out conversation.Session
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&out.ID, &out.IssueID, &out.PairID, &out.Status, &out.Goal, &out.Questions, &out.StartState, &out.EndState, &out.ResultStatus, &out.ResultText, &out.CreatedAt, &out.FinishedAt)
	if isNoRows(err) {
		return conversation.Session{}, repository.ErrNotFound
	}
	if err != nil {
		return conversation.Session{}, err
	}
	return out, nil
}

func (r *ConversationRepo) PauseSession(ctx context.Context, id string) (conversation.Session, error) {
	const q = `
UPDATE conversation_sessions
SET status = 'paused'
WHERE id = $1
RETURNING id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at;
`
	var out conversation.Session
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&out.ID, &out.IssueID, &out.PairID, &out.Status, &out.Goal, &out.Questions, &out.StartState, &out.EndState, &out.ResultStatus, &out.ResultText, &out.CreatedAt, &out.FinishedAt)
	if isNoRows(err) {
		return conversation.Session{}, repository.ErrNotFound
	}
	if err != nil {
		return conversation.Session{}, err
	}
	return out, nil
}

func (r *ConversationRepo) ResumeSession(ctx context.Context, id string) (conversation.Session, error) {
	const q = `
UPDATE conversation_sessions
SET status = 'started'
WHERE id = $1
RETURNING id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at;
`
	var out conversation.Session
	err := r.pool.QueryRow(ctx, q, id).
		Scan(&out.ID, &out.IssueID, &out.PairID, &out.Status, &out.Goal, &out.Questions, &out.StartState, &out.EndState, &out.ResultStatus, &out.ResultText, &out.CreatedAt, &out.FinishedAt)
	if isNoRows(err) {
		return conversation.Session{}, repository.ErrNotFound
	}
	if err != nil {
		return conversation.Session{}, err
	}
	return out, nil
}

func (r *ConversationRepo) ListByPair(ctx context.Context, pairID string, status *conversation.Status, limit, offset int) ([]conversation.Session, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	q := `
SELECT id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at
FROM conversation_sessions
WHERE pair_id = $1
`
	args := []any{pairID}
	if status != nil {
		q += " AND status = $2"
		args = append(args, *status)
	}
	q += " ORDER BY created_at DESC LIMIT $3 OFFSET $4;"
	if status == nil {
		q = `
SELECT id, issue_id, pair_id, status, goal, questions, start_state, end_state, result_status, result_text, created_at, finished_at
FROM conversation_sessions
WHERE pair_id = $1
ORDER BY created_at DESC LIMIT $2 OFFSET $3;
`
		args = []any{pairID, limit, offset}
	} else {
		args = []any{pairID, *status, limit, offset}
	}

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []conversation.Session
	for rows.Next() {
		var s conversation.Session
		if err := rows.Scan(&s.ID, &s.IssueID, &s.PairID, &s.Status, &s.Goal, &s.Questions, &s.StartState, &s.EndState, &s.ResultStatus, &s.ResultText, &s.CreatedAt, &s.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *ConversationRepo) AddNote(ctx context.Context, conversationID, userID, text string) error {
	const q = `
INSERT INTO conversation_notes (conversation_id, user_id, text)
VALUES ($1, $2, $3);
`
	_, err := r.pool.Exec(ctx, q, conversationID, userID, text)
	return err
}

func (r *ConversationRepo) LinkSideIssue(ctx context.Context, conversationID, issueID string) error {
	const q = `
INSERT INTO conversation_side_issues (conversation_id, issue_id)
VALUES ($1, $2)
ON CONFLICT (conversation_id, issue_id) DO NOTHING;
`
	_, err := r.pool.Exec(ctx, q, conversationID, issueID)
	return err
}

func (r *ConversationRepo) ListNotes(ctx context.Context, conversationID string, limit, offset int) ([]conversation.Note, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	const q = `
SELECT id, conversation_id, user_id, text, created_at
FROM conversation_notes
WHERE conversation_id = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;
`
	rows, err := r.pool.Query(ctx, q, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []conversation.Note
	for rows.Next() {
		var n conversation.Note
		if err := rows.Scan(&n.ID, &n.ConversationID, &n.UserID, &n.Text, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *ConversationRepo) ListSideIssues(ctx context.Context, conversationID string) ([]issue.Issue, error) {
	const q = `
SELECT i.id, i.pair_id, i.created_by_user_id, i.title, i.description, i.priority, i.visibility,
       i.repeat_threshold, i.repeat_count, i.status, i.created_at, i.updated_at, i.resolved_at
FROM conversation_side_issues csi
JOIN issues i ON i.id = csi.issue_id
WHERE csi.conversation_id = $1
ORDER BY csi.created_at ASC;
`
	rows, err := r.pool.Query(ctx, q, conversationID)
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

func (r *ConversationRepo) ListNotesByPair(ctx context.Context, pairID string, limit, offset int) ([]conversation.PairNote, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	const q = `
SELECT cn.id, cn.conversation_id, cs.issue_id, i.title, cn.user_id, cn.text, cn.created_at
FROM conversation_notes cn
JOIN conversation_sessions cs ON cs.id = cn.conversation_id
JOIN issues i ON i.id = cs.issue_id
WHERE cs.pair_id = $1
ORDER BY cn.created_at DESC
LIMIT $2 OFFSET $3;
`
	rows, err := r.pool.Query(ctx, q, pairID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []conversation.PairNote
	for rows.Next() {
		var n conversation.PairNote
		if err := rows.Scan(&n.ID, &n.ConversationID, &n.IssueID, &n.IssueTitle, &n.UserID, &n.Text, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *ConversationRepo) DeleteNote(ctx context.Context, noteID string) error {
	const q = `DELETE FROM conversation_notes WHERE id = $1;`
	ct, err := r.pool.Exec(ctx, q, noteID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}
