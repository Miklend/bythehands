package service

import (
	"context"
	"testing"

	"talkabout/internal/domain/issue"
	"talkabout/internal/domain/pair"
	"talkabout/internal/domain/user"
	"talkabout/internal/repository"
)

type fakeIssueRepo struct {
	create func(ctx context.Context, in issue.Issue) (issue.Issue, error)
}

func (f fakeIssueRepo) CreateIssue(ctx context.Context, in issue.Issue) (issue.Issue, error) {
	return f.create(ctx, in)
}
func (f fakeIssueRepo) ListIssuesByPair(ctx context.Context, pairID string, status *issue.Status) ([]issue.Issue, error) {
	return nil, nil
}
func (f fakeIssueRepo) GetIssue(ctx context.Context, issueID string) (issue.Issue, error) {
	return issue.Issue{}, repository.ErrNotFound
}
func (f fakeIssueRepo) RepeatIssue(ctx context.Context, issueID, userID string, note *string) (issue.Issue, error) {
	return issue.Issue{}, nil
}
func (f fakeIssueRepo) IncrementRepeat(ctx context.Context, issueID string) (issue.Issue, error) {
	return issue.Issue{}, nil
}
func (f fakeIssueRepo) CreateRepeat(ctx context.Context, issueID, userID string, note *string) (issue.IssueRepeat, error) {
	return issue.IssueRepeat{}, nil
}
func (f fakeIssueRepo) ListRepeatsByIssue(ctx context.Context, issueID string, limit, offset int) ([]issue.IssueRepeat, error) {
	return nil, nil
}
func (f fakeIssueRepo) GetRepeat(ctx context.Context, repeatID string) (issue.IssueRepeat, error) {
	return issue.IssueRepeat{}, repository.ErrNotFound
}
func (f fakeIssueRepo) CreateRepeatDisagreement(ctx context.Context, repeatID, userID string, note string) (issue.IssueRepeatDisagreement, error) {
	return issue.IssueRepeatDisagreement{}, nil
}
func (f fakeIssueRepo) GetRepeatDisagreement(ctx context.Context, repeatID, userID string) (issue.IssueRepeatDisagreement, error) {
	return issue.IssueRepeatDisagreement{}, repository.ErrNotFound
}
func (f fakeIssueRepo) UpdateIssue(ctx context.Context, issueID string, title *string, repeatThreshold *int, repeatLimit *int) (issue.Issue, error) {
	return issue.Issue{}, nil
}
func (f fakeIssueRepo) UpdateStatus(ctx context.Context, issueID string, status issue.Status) (issue.Issue, error) {
	return issue.Issue{}, nil
}
func (f fakeIssueRepo) DeleteIssue(ctx context.Context, issueID string) error   { return nil }
func (f fakeIssueRepo) DeleteRepeat(ctx context.Context, repeatID string) error { return nil }

type fakePairsRepoRead struct {
	get func(ctx context.Context, pairID string) (pair.Pair, error)
}

func (f fakePairsRepoRead) CreatePair(ctx context.Context, isTest bool) (pair.Pair, error) {
	return pair.Pair{}, nil
}
func (f fakePairsRepoRead) AddMember(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
	return pair.PairMember{}, nil
}
func (f fakePairsRepoRead) GetPair(ctx context.Context, pairID string) (pair.Pair, error) {
	return f.get(ctx, pairID)
}
func (f fakePairsRepoRead) GetMembers(ctx context.Context, pairID string) ([]pair.PairMember, error) {
	return nil, nil
}
func (f fakePairsRepoRead) ListPairsByUser(ctx context.Context, userID string) ([]pair.Pair, error) {
	return nil, nil
}
func (f fakePairsRepoRead) ArchivePair(ctx context.Context, pairID string) (pair.Pair, error) {
	return pair.Pair{}, nil
}
func (f fakePairsRepoRead) SetWelcomeMessage(ctx context.Context, pairID string, text *string) (pair.Pair, error) {
	return pair.Pair{}, nil
}
func (f fakePairsRepoRead) SetMemberName(ctx context.Context, pairID, userID string, name *string) (pair.PairMember, error) {
	return pair.PairMember{}, nil
}

func TestIssueService_CreateIssue_Defaults(t *testing.T) {
	issRepo := fakeIssueRepo{create: func(ctx context.Context, in issue.Issue) (issue.Issue, error) {
		if in.Priority != issue.PriorityMedium {
			t.Fatalf("expected default priority=medium, got %q", in.Priority)
		}
		if in.Visibility != issue.VisibilityVisible {
			t.Fatalf("expected default visibility=visible, got %q", in.Visibility)
		}
		return in, nil
	}}

	svc := NewIssueService(
		fakeUsersRepo{getByID: func(ctx context.Context, id string) (user.User, error) { return user.User{ID: id}, nil }},
		fakePairsRepoRead{get: func(ctx context.Context, pairID string) (pair.Pair, error) { return pair.Pair{ID: pairID}, nil }},
		issRepo,
	)

	_, err := svc.CreateIssue(context.Background(), CreateIssueInput{
		PairID:          "11111111-1111-1111-1111-111111111111",
		CreatedByUserID: "22222222-2222-2222-2222-222222222222",
		Title:           "A",
		Description:     "B",
		RepeatThreshold: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIssueService_CreateIssue_EmptyTitle(t *testing.T) {
	svc := NewIssueService(
		fakeUsersRepo{getByID: func(ctx context.Context, id string) (user.User, error) { return user.User{ID: id}, nil }},
		fakePairsRepoRead{get: func(ctx context.Context, pairID string) (pair.Pair, error) { return pair.Pair{ID: pairID}, nil }},
		fakeIssueRepo{create: func(ctx context.Context, in issue.Issue) (issue.Issue, error) { return in, nil }},
	)

	_, err := svc.CreateIssue(context.Background(), CreateIssueInput{
		PairID:          "11111111-1111-1111-1111-111111111111",
		CreatedByUserID: "22222222-2222-2222-2222-222222222222",
		Title:           "   ",
		Description:     "B",
	})
	if err == nil || !IsKind(err, KindValidation) {
		t.Fatalf("expected validation error, got: %v", err)
	}
}
