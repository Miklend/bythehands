package service

import (
	"context"
	"testing"
	"time"

	"taalkbout/internal/domain/pair"
	"taalkbout/internal/domain/user"
	"taalkbout/internal/repository"
)

type fakeUsersRepo struct {
	getByID func(ctx context.Context, id string) (user.User, error)
}

func (f fakeUsersRepo) UpsertTelegramUser(ctx context.Context, telegramID int64, username, displayName *string) (user.User, error) {
	return user.User{}, nil
}
func (f fakeUsersRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	return f.getByID(ctx, id)
}
func (f fakeUsersRepo) CreateVirtualUser(ctx context.Context, username, displayName string) (user.User, error) {
	return user.User{}, nil
}

type fakePairsRepo struct {
	createPair func(ctx context.Context, isTest bool) (pair.Pair, error)
	addMember  func(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error)
}

func (f fakePairsRepo) CreatePair(ctx context.Context, isTest bool) (pair.Pair, error) {
	return f.createPair(ctx, isTest)
}
func (f fakePairsRepo) AddMember(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
	return f.addMember(ctx, pairID, userID, role)
}
func (f fakePairsRepo) GetPair(ctx context.Context, pairID string) (pair.Pair, error) {
	return pair.Pair{}, nil
}
func (f fakePairsRepo) GetMembers(ctx context.Context, pairID string) ([]pair.PairMember, error) {
	return nil, nil
}
func (f fakePairsRepo) ListPairsByUser(ctx context.Context, userID string) ([]pair.Pair, error) {
	return nil, nil
}
func (f fakePairsRepo) ArchivePair(ctx context.Context, pairID string) (pair.Pair, error) {
	return pair.Pair{}, nil
}
func (f fakePairsRepo) SetWelcomeMessage(ctx context.Context, pairID string, text *string) (pair.Pair, error) {
	return pair.Pair{}, nil
}
func (f fakePairsRepo) SetMemberName(ctx context.Context, pairID, userID string, name *string) (pair.PairMember, error) {
	return pair.PairMember{}, nil
}

type fakeInvitesRepo struct {
	create func(ctx context.Context, pairID, token string, expiresAt time.Time) (pair.Invite, error)
}

func (f fakeInvitesRepo) CreateInvite(ctx context.Context, pairID, token string, expiresAt time.Time) (pair.Invite, error) {
	return f.create(ctx, pairID, token, expiresAt)
}
func (f fakeInvitesRepo) GetByToken(ctx context.Context, token string) (pair.Invite, error) {
	return pair.Invite{}, nil
}
func (f fakeInvitesRepo) MarkUsed(ctx context.Context, inviteID string, usedAt time.Time) (pair.Invite, error) {
	return pair.Invite{}, nil
}
func (f fakeInvitesRepo) MarkExpired(ctx context.Context, inviteID string) (pair.Invite, error) {
	return pair.Invite{}, nil
}

func TestPairService_CreatePair_InvalidUserID(t *testing.T) {
	svc := NewPairService(
		fakeUsersRepo{getByID: func(ctx context.Context, id string) (user.User, error) { return user.User{}, nil }},
		fakePairsRepo{
			createPair: func(ctx context.Context, isTest bool) (pair.Pair, error) { return pair.Pair{}, nil },
			addMember: func(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
				return pair.PairMember{}, nil
			},
		},
		fakeInvitesRepo{create: func(ctx context.Context, pairID, token string, expiresAt time.Time) (pair.Invite, error) {
			return pair.Invite{}, nil
		}},
	)

	_, err := svc.CreatePair(context.Background(), "not-a-uuid")
	if err == nil || !IsKind(err, KindValidation) {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

func TestPairService_CreatePair_UserNotFound(t *testing.T) {
	svc := NewPairService(
		fakeUsersRepo{getByID: func(ctx context.Context, id string) (user.User, error) {
			return user.User{}, repository.ErrNotFound
		}},
		fakePairsRepo{
			createPair: func(ctx context.Context, isTest bool) (pair.Pair, error) {
				t.Fatal("should not be called")
				return pair.Pair{}, nil
			},
			addMember: func(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
				return pair.PairMember{}, nil
			},
		},
		fakeInvitesRepo{create: func(ctx context.Context, pairID, token string, expiresAt time.Time) (pair.Invite, error) {
			return pair.Invite{}, nil
		}},
	)

	_, err := svc.CreatePair(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err == nil || !IsKind(err, KindNotFound) {
		t.Fatalf("expected not_found error, got: %v", err)
	}
}
