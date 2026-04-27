package service

import (
	"context"

	"talkabout/internal/domain/pair"
	"talkabout/internal/repository"
)

type PreferencesService struct {
	users repository.UserRepository
	pairs repository.PairRepository
	pref  repository.PreferencesRepository
}

type Preferences struct {
	CurrentPairID *string `json:"current_pair_id"`
}

func NewPreferencesService(users repository.UserRepository, pairs repository.PairRepository, pref repository.PreferencesRepository) *PreferencesService {
	return &PreferencesService{users: users, pairs: pairs, pref: pref}
}

func (s *PreferencesService) Get(ctx context.Context, userID string) (Preferences, error) {
	if !validateUUID(userID) {
		return Preferences{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return Preferences{}, notFound("user not found", err)
		}
		return Preferences{}, err
	}
	cur, err := s.pref.GetPreferences(ctx, userID)
	if err != nil {
		return Preferences{}, err
	}
	return Preferences{CurrentPairID: cur}, nil
}

func (s *PreferencesService) SetCurrentPair(ctx context.Context, userID string, pairID *string) (Preferences, error) {
	if !validateUUID(userID) {
		return Preferences{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return Preferences{}, notFound("user not found", err)
		}
		return Preferences{}, err
	}
	if pairID != nil {
		if !validateUUID(*pairID) {
			return Preferences{}, validation("invalid current_pair_id")
		}
		p, err := s.pairs.GetPair(ctx, *pairID)
		if err != nil {
			if err == repository.ErrNotFound {
				return Preferences{}, notFound("pair not found", err)
			}
			return Preferences{}, err
		}
		members, err := s.pairs.GetMembers(ctx, p.ID)
		if err != nil {
			return Preferences{}, err
		}
		ok := false
		for _, m := range members {
			if m.UserID == userID {
				ok = true
				break
			}
		}
		if !ok {
			return Preferences{}, forbidden("user is not a member of pair", nil)
		}
	}
	if err := s.pref.SetCurrentPair(ctx, userID, pairID); err != nil {
		return Preferences{}, err
	}
	return Preferences{CurrentPairID: pairID}, nil
}

func (s *PreferencesService) ListPairs(ctx context.Context, userID string) ([]pair.Pair, error) {
	if !validateUUID(userID) {
		return nil, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return nil, notFound("user not found", err)
		}
		return nil, err
	}
	return s.pairs.ListPairsByUser(ctx, userID)
}
