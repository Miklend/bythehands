package service

import (
	"context"
	"testing"

	"talkabout/internal/domain/pair"
	"talkabout/internal/domain/user"
	"talkabout/internal/repository"
)

type fakePrefsRepo struct {
	current *string
}

func (f *fakePrefsRepo) GetPreferences(ctx context.Context, userID string) (*string, error) {
	return f.current, nil
}
func (f *fakePrefsRepo) SetCurrentPair(ctx context.Context, userID string, pairID *string) error {
	f.current = pairID
	return nil
}

type fakeUsersRepoTest struct {
	fakeUsersRepo
	virtual func(ctx context.Context, username, displayName string) (user.User, error)
}

func (f fakeUsersRepoTest) CreateVirtualUser(ctx context.Context, username, displayName string) (user.User, error) {
	return f.virtual(ctx, username, displayName)
}

func TestTestModeService_Start_SetsPreference(t *testing.T) {
	prefs := &fakePrefsRepo{}

	users := fakeUsersRepoTest{
		fakeUsersRepo: fakeUsersRepo{getByID: func(ctx context.Context, id string) (user.User, error) { return user.User{ID: id}, nil }},
		virtual: func(ctx context.Context, username, displayName string) (user.User, error) {
			return user.User{ID: "33333333-3333-3333-3333-333333333333", IsVirtual: true}, nil
		},
	}

	pairsRepo := fakePairsRepo{
		createPair: func(ctx context.Context, isTest bool) (pair.Pair, error) {
			if !isTest {
				t.Fatalf("expected isTest=true")
			}
			return pair.Pair{ID: "11111111-1111-1111-1111-111111111111", IsTest: true, Status: pair.StatusActive}, nil
		},
		addMember: func(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
			return pair.PairMember{}, nil
		},
	}

	svc := NewTestModeService(users, pairsRepo, prefs)
	res, err := svc.Start(context.Background(), "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prefs.current == nil || *prefs.current != res.Pair.ID {
		t.Fatalf("expected current pair to be set")
	}
}

func TestTestModeService_Stop_WhenNotActive(t *testing.T) {
	prefs := &fakePrefsRepo{}
	users := fakeUsersRepoTest{
		fakeUsersRepo: fakeUsersRepo{getByID: func(ctx context.Context, id string) (user.User, error) { return user.User{ID: id}, nil }},
		virtual:       func(ctx context.Context, username, displayName string) (user.User, error) { return user.User{}, nil },
	}
	pairsRepo := fakePairsRepo{
		createPair: func(ctx context.Context, isTest bool) (pair.Pair, error) { return pair.Pair{}, nil },
		addMember: func(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error) {
			return pair.PairMember{}, nil
		},
	}
	svc := NewTestModeService(users, pairsRepo, prefs)
	_, err := svc.Stop(context.Background(), "22222222-2222-2222-2222-222222222222")
	if err == nil || !IsKind(err, KindValidation) {
		t.Fatalf("expected validation error, got: %v", err)
	}
}

// Keep fake repo compile-time checks.
var _ repository.PreferencesRepository = (*fakePrefsRepo)(nil)
