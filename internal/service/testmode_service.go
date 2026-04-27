package service

import (
	"context"

	"talkabout/internal/domain/pair"
	"talkabout/internal/domain/user"
	"talkabout/internal/repository"
)

type TestModeService struct {
	users repository.UserRepository
	pairs repository.PairRepository
	pref  repository.PreferencesRepository
}

type TestModeStartResult struct {
	Pair           pair.Pair `json:"pair"`
	VirtualPartner user.User `json:"virtual_partner"`
}

type TestModeStopResult struct {
	ArchivedPair  pair.Pair  `json:"archived_pair"`
	CurrentPairID *string    `json:"current_pair_id"`
	CurrentIsTest bool       `json:"current_is_test"`
	CurrentPair   *pair.Pair `json:"current_pair,omitempty"`
}

func NewTestModeService(users repository.UserRepository, pairs repository.PairRepository, pref repository.PreferencesRepository) *TestModeService {
	return &TestModeService{users: users, pairs: pairs, pref: pref}
}

func (s *TestModeService) Start(ctx context.Context, userID string) (TestModeStartResult, error) {
	if !validateUUID(userID) {
		return TestModeStartResult{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return TestModeStartResult{}, notFound("user not found", err)
		}
		return TestModeStartResult{}, err
	}

	// Reuse existing active test pair to avoid creating many test pairs.
	pairs, err := s.pairs.ListPairsByUser(ctx, userID)
	if err != nil {
		return TestModeStartResult{}, err
	}
	var activeTest *pair.Pair
	for _, it := range pairs {
		if !it.IsTest || it.Status != pair.StatusActive {
			continue
		}
		if activeTest == nil {
			c := it
			activeTest = &c
			continue
		}
		// Best-effort cleanup of older active test pairs.
		_, _ = s.pairs.ArchivePair(ctx, it.ID)
	}
	if activeTest != nil {
		members, err := s.pairs.GetMembers(ctx, activeTest.ID)
		if err != nil {
			return TestModeStartResult{}, err
		}
		var virtual *user.User
		for _, m := range members {
			u, err := s.users.GetByID(ctx, m.UserID)
			if err != nil {
				continue
			}
			if u.IsVirtual {
				c := u
				virtual = &c
				break
			}
		}
		if virtual == nil {
			v, err := s.users.CreateVirtualUser(ctx, "test_partner", "Тестовый партнер")
			if err != nil {
				return TestModeStartResult{}, err
			}
			if _, err := s.pairs.AddMember(ctx, activeTest.ID, v.ID, pair.RolePartner); err != nil && err != repository.ErrConflict {
				return TestModeStartResult{}, err
			}
			virtual = &v
		}

		pid := activeTest.ID
		if err := s.pref.SetCurrentPair(ctx, userID, &pid); err != nil {
			return TestModeStartResult{}, err
		}
		return TestModeStartResult{Pair: *activeTest, VirtualPartner: *virtual}, nil
	}

	p, err := s.pairs.CreatePair(ctx, true)
	if err != nil {
		return TestModeStartResult{}, err
	}

	if _, err := s.pairs.AddMember(ctx, p.ID, userID, pair.RoleCreator); err != nil {
		if err == repository.ErrConflict {
			return TestModeStartResult{}, conflict("pair member conflict", err)
		}
		return TestModeStartResult{}, err
	}

	virtual, err := s.users.CreateVirtualUser(ctx, "test_partner", "Тестовый партнер")
	if err != nil {
		return TestModeStartResult{}, err
	}
	if _, err := s.pairs.AddMember(ctx, p.ID, virtual.ID, pair.RolePartner); err != nil {
		if err == repository.ErrConflict {
			return TestModeStartResult{}, conflict("pair member conflict", err)
		}
		return TestModeStartResult{}, err
	}

	pid := p.ID
	if err := s.pref.SetCurrentPair(ctx, userID, &pid); err != nil {
		return TestModeStartResult{}, err
	}

	return TestModeStartResult{Pair: p, VirtualPartner: virtual}, nil
}

func (s *TestModeService) Stop(ctx context.Context, userID string) (TestModeStopResult, error) {
	if !validateUUID(userID) {
		return TestModeStopResult{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return TestModeStopResult{}, notFound("user not found", err)
		}
		return TestModeStopResult{}, err
	}

	cur, err := s.pref.GetPreferences(ctx, userID)
	if err != nil {
		return TestModeStopResult{}, err
	}
	if cur == nil {
		return TestModeStopResult{}, validation("test mode is not active")
	}

	p, err := s.pairs.GetPair(ctx, *cur)
	if err != nil {
		if err == repository.ErrNotFound {
			return TestModeStopResult{}, notFound("pair not found", err)
		}
		return TestModeStopResult{}, err
	}
	if !p.IsTest {
		return TestModeStopResult{}, validation("test mode is not active")
	}

	archived, err := s.pairs.ArchivePair(ctx, p.ID)
	if err != nil {
		return TestModeStopResult{}, err
	}

	pairs, err := s.pairs.ListPairsByUser(ctx, userID)
	if err != nil {
		return TestModeStopResult{}, err
	}
	var next *pair.Pair
	for _, it := range pairs {
		if it.IsTest {
			continue
		}
		if it.Status == pair.StatusActive {
			c := it
			next = &c
			break
		}
	}
	var nextID *string
	var nextIsTest bool
	if next != nil {
		nextID = &next.ID
		nextIsTest = next.IsTest
	}

	if err := s.pref.SetCurrentPair(ctx, userID, nextID); err != nil {
		return TestModeStopResult{}, err
	}

	return TestModeStopResult{
		ArchivedPair:  archived,
		CurrentPairID: nextID,
		CurrentIsTest: nextIsTest,
		CurrentPair:   next,
	}, nil
}
