package service

import (
	"context"
	"strings"
	"time"

	"talkabout/internal/domain/pair"
	"talkabout/internal/repository"
)

type PairService struct {
	users   repository.UserRepository
	pairs   repository.PairRepository
	invites repository.InviteRepository
	now     func() time.Time
}

type CreatePairResult struct {
	Pair        pair.Pair `json:"pair"`
	InviteToken string    `json:"invite_token"`
}

func NewPairService(users repository.UserRepository, pairs repository.PairRepository, invites repository.InviteRepository) *PairService {
	return &PairService{
		users:   users,
		pairs:   pairs,
		invites: invites,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

func (s *PairService) CreatePair(ctx context.Context, userID string) (CreatePairResult, error) {
	if !validateUUID(userID) {
		return CreatePairResult{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return CreatePairResult{}, notFound("user not found", err)
		}
		return CreatePairResult{}, err
	}

	// One user -> one active non-test pair. Reuse if exists.
	pairs, err := s.pairs.ListPairsByUser(ctx, userID)
	if err != nil {
		return CreatePairResult{}, err
	}
	var existing *pair.Pair
	for _, it := range pairs {
		if it.IsTest || it.Status != pair.StatusActive {
			continue
		}
		if existing == nil {
			c := it
			existing = &c
			continue
		}
		// Best-effort cleanup of legacy duplicates.
		_, _ = s.pairs.ArchivePair(ctx, it.ID)
	}

	var p pair.Pair
	if existing != nil {
		p = *existing
	} else {
		created, err := s.pairs.CreatePair(ctx, false)
		if err != nil {
			return CreatePairResult{}, err
		}
		if _, err := s.pairs.AddMember(ctx, created.ID, userID, pair.RoleCreator); err != nil {
			if err == repository.ErrConflict {
				return CreatePairResult{}, conflict("pair member conflict", err)
			}
			return CreatePairResult{}, err
		}
		p = created
	}

	token, err := s.CreateInvite(ctx, p.ID, userID)
	if err != nil {
		return CreatePairResult{}, err
	}

	return CreatePairResult{Pair: p, InviteToken: token}, nil
}

func (s *PairService) CreateInvite(ctx context.Context, pairID string, userID string) (string, error) {
	if !validateUUID(pairID) {
		return "", validation("invalid pair_id")
	}
	if !validateUUID(userID) {
		return "", validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return "", notFound("user not found", err)
		}
		return "", err
	}
	p, err := s.pairs.GetPair(ctx, pairID)
	if err != nil {
		if err == repository.ErrNotFound {
			return "", notFound("pair not found", err)
		}
		return "", err
	}
	if p.Status != pair.StatusActive {
		return "", validation("pair is not active")
	}
	members, err := s.pairs.GetMembers(ctx, pairID)
	if err != nil {
		return "", err
	}
	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return "", forbidden("user is not member of pair", nil)
	}

	var token string
	created := false
	for i := 0; i < 3; i++ {
		t, err := newInviteToken()
		if err != nil {
			return "", err
		}
		token = t
		_, err = s.invites.CreateInvite(ctx, pairID, token, s.now().Add(7*24*time.Hour))
		if err == repository.ErrConflict {
			continue
		}
		if err != nil {
			return "", err
		}
		created = true
		break
	}
	if !created {
		return "", &Error{Kind: KindInternal, Message: "failed to generate unique invite token"}
	}
	return token, nil
}

func (s *PairService) SetWelcomeMessage(ctx context.Context, pairID string, userID string, text *string) (pair.Pair, error) {
	if !validateUUID(pairID) {
		return pair.Pair{}, validation("invalid pair_id")
	}
	if !validateUUID(userID) {
		return pair.Pair{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return pair.Pair{}, notFound("user not found", err)
		}
		return pair.Pair{}, err
	}
	p, err := s.pairs.GetPair(ctx, pairID)
	if err != nil {
		if err == repository.ErrNotFound {
			return pair.Pair{}, notFound("pair not found", err)
		}
		return pair.Pair{}, err
	}
	if p.Status != pair.StatusActive {
		return pair.Pair{}, validation("pair is not active")
	}
	members, err := s.pairs.GetMembers(ctx, pairID)
	if err != nil {
		return pair.Pair{}, err
	}
	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return pair.Pair{}, forbidden("user is not member of pair", nil)
	}

	var cleaned *string
	if text != nil {
		v := strings.TrimSpace(*text)
		if v != "" {
			cleaned = &v
		}
	}
	updated, err := s.pairs.SetWelcomeMessage(ctx, pairID, cleaned)
	if err != nil {
		if err == repository.ErrNotFound {
			return pair.Pair{}, notFound("pair not found", err)
		}
		return pair.Pair{}, err
	}
	return updated, nil
}

func (s *PairService) SetMemberName(ctx context.Context, pairID string, userID string, name string) (pair.PairMember, error) {
	if !validateUUID(pairID) {
		return pair.PairMember{}, validation("invalid pair_id")
	}
	if !validateUUID(userID) {
		return pair.PairMember{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return pair.PairMember{}, notFound("user not found", err)
		}
		return pair.PairMember{}, err
	}
	p, err := s.pairs.GetPair(ctx, pairID)
	if err != nil {
		if err == repository.ErrNotFound {
			return pair.PairMember{}, notFound("pair not found", err)
		}
		return pair.PairMember{}, err
	}
	if p.Status != pair.StatusActive {
		return pair.PairMember{}, validation("pair is not active")
	}
	clean := strings.TrimSpace(name)
	if clean == "" {
		return pair.PairMember{}, validation("name is required")
	}
	members, err := s.pairs.GetMembers(ctx, pairID)
	if err != nil {
		return pair.PairMember{}, err
	}
	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return pair.PairMember{}, forbidden("user is not member of pair", nil)
	}
	updated, err := s.pairs.SetMemberName(ctx, pairID, userID, &clean)
	if err != nil {
		if err == repository.ErrNotFound {
			return pair.PairMember{}, notFound("pair member not found", err)
		}
		return pair.PairMember{}, err
	}
	return updated, nil
}

func (s *PairService) GetPair(ctx context.Context, pairID string) (pair.Pair, []pair.PairMember, error) {
	if !validateUUID(pairID) {
		return pair.Pair{}, nil, validation("invalid pair_id")
	}
	p, err := s.pairs.GetPair(ctx, pairID)
	if err != nil {
		if err == repository.ErrNotFound {
			return pair.Pair{}, nil, notFound("pair not found", err)
		}
		return pair.Pair{}, nil, err
	}
	members, err := s.pairs.GetMembers(ctx, pairID)
	if err != nil {
		return pair.Pair{}, nil, err
	}
	return p, members, nil
}

func (s *PairService) JoinByInvite(ctx context.Context, token string, userID string) (pair.Pair, []pair.PairMember, error) {
	if token == "" {
		return pair.Pair{}, nil, validation("token is required")
	}
	if !validateUUID(userID) {
		return pair.Pair{}, nil, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return pair.Pair{}, nil, notFound("user not found", err)
		}
		return pair.Pair{}, nil, err
	}

	inv, err := s.invites.GetByToken(ctx, token)
	if err != nil {
		if err == repository.ErrNotFound {
			return pair.Pair{}, nil, notFound("invite not found", err)
		}
		return pair.Pair{}, nil, err
	}

	now := s.now()
	if inv.Status != pair.InviteActive {
		return pair.Pair{}, nil, forbidden("invite is not active", err)
	}
	if now.After(inv.ExpiresAt) {
		_, _ = s.invites.MarkExpired(ctx, inv.ID)
		return pair.Pair{}, nil, forbidden("invite expired", err)
	}

	if _, err := s.pairs.AddMember(ctx, inv.PairID, userID, pair.RolePartner); err != nil {
		if err == repository.ErrConflict {
			return pair.Pair{}, nil, conflict("pair member conflict", err)
		}
		return pair.Pair{}, nil, err
	}

	if _, err := s.invites.MarkUsed(ctx, inv.ID, now); err != nil {
		return pair.Pair{}, nil, err
	}

	p, members, err := s.GetPair(ctx, inv.PairID)
	if err != nil {
		return pair.Pair{}, nil, err
	}
	return p, members, nil
}
