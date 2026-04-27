package service

import (
	"context"
	"strings"

	"taalkbout/internal/domain/issue"
	"taalkbout/internal/repository"
)

type IssueService struct {
	users repository.UserRepository
	pairs repository.PairRepository
	iss   repository.IssueRepository
}

func NewIssueService(users repository.UserRepository, pairs repository.PairRepository, iss repository.IssueRepository) *IssueService {
	return &IssueService{users: users, pairs: pairs, iss: iss}
}

type CreateIssueInput struct {
	PairID          string
	CreatedByUserID string
	Title           string
	Description     string
	Priority        issue.Priority
	Visibility      issue.Visibility
	RepeatThreshold int // лимит показа (для скрытых тем)
	RepeatLimit     int // лимит повторений (0 = без лимита)
}

func (s *IssueService) CreateIssue(ctx context.Context, in CreateIssueInput) (issue.Issue, error) {
	if !validateUUID(in.PairID) {
		return issue.Issue{}, validation("invalid pair_id")
	}
	if !validateUUID(in.CreatedByUserID) {
		return issue.Issue{}, validation("invalid created_by_user_id")
	}
	if strings.TrimSpace(in.Title) == "" {
		return issue.Issue{}, validation("title is required")
	}
	if strings.TrimSpace(in.Description) == "" {
		return issue.Issue{}, validation("description is required")
	}
	if in.RepeatThreshold < 0 {
		return issue.Issue{}, validation("repeat_threshold must be >= 0")
	}
	if in.RepeatLimit < 0 {
		return issue.Issue{}, validation("repeat_limit must be >= 0")
	}

	if in.Priority == "" {
		in.Priority = issue.PriorityMedium
	}
	switch in.Priority {
	case issue.PriorityLow, issue.PriorityMedium, issue.PriorityHigh:
	default:
		return issue.Issue{}, validation("invalid priority")
	}

	if in.Visibility == "" {
		in.Visibility = issue.VisibilityVisible
	}
	switch in.Visibility {
	case issue.VisibilityVisible, issue.VisibilityHiddenUntilRepeat, issue.VisibilityPrivate:
	default:
		return issue.Issue{}, validation("invalid visibility")
	}

	if _, err := s.pairs.GetPair(ctx, in.PairID); err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("pair not found", err)
		}
		return issue.Issue{}, err
	}
	if _, err := s.users.GetByID(ctx, in.CreatedByUserID); err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("user not found", err)
		}
		return issue.Issue{}, err
	}

	// no duplicates in the same pair (trim + case-insensitive)
	existing, err := s.iss.ListIssuesByPair(ctx, in.PairID, nil)
	if err != nil {
		return issue.Issue{}, err
	}
	titleNorm := strings.TrimSpace(in.Title)
	for _, it := range existing {
		if strings.EqualFold(strings.TrimSpace(it.Title), titleNorm) {
			return issue.Issue{}, conflict("issue title already exists", nil)
		}
	}

	return s.iss.CreateIssue(ctx, issue.Issue{
		PairID:          in.PairID,
		CreatedByUserID: in.CreatedByUserID,
		Title:           strings.TrimSpace(in.Title),
		Description:     strings.TrimSpace(in.Description),
		Priority:        in.Priority,
		Visibility:      in.Visibility,
		RepeatThreshold: in.RepeatThreshold,
		RepeatLimit:     in.RepeatLimit,
	})
}

func (s *IssueService) ListIssues(ctx context.Context, pairID string, status *issue.Status) ([]issue.Issue, error) {
	if !validateUUID(pairID) {
		return nil, validation("invalid pair_id")
	}
	if status != nil {
		switch *status {
		case issue.StatusActive, issue.StatusResolved, issue.StatusPostponed, issue.StatusArchived:
		default:
			return nil, validation("invalid status")
		}
	}
	if _, err := s.pairs.GetPair(ctx, pairID); err != nil {
		if err == repository.ErrNotFound {
			return nil, notFound("pair not found", err)
		}
		return nil, err
	}
	return s.iss.ListIssuesByPair(ctx, pairID, status)
}

func (s *IssueService) GetIssue(ctx context.Context, issueID string) (issue.Issue, error) {
	if !validateUUID(issueID) {
		return issue.Issue{}, validation("invalid issue_id")
	}
	it, err := s.iss.GetIssue(ctx, issueID)
	if err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("issue not found", err)
		}
		return issue.Issue{}, err
	}
	return it, nil
}

type UpdateIssueInput struct {
	Title           *string
	RepeatThreshold *int
	RepeatLimit     *int
}

func (s *IssueService) UpdateIssue(ctx context.Context, issueID string, in UpdateIssueInput) (issue.Issue, error) {
	if !validateUUID(issueID) {
		return issue.Issue{}, validation("invalid issue_id")
	}
	var title *string
	if in.Title != nil {
		t := strings.TrimSpace(*in.Title)
		if t == "" {
			return issue.Issue{}, validation("title is required")
		}
		title = &t
	}
	var thr *int
	if in.RepeatThreshold != nil {
		if *in.RepeatThreshold < 0 {
			return issue.Issue{}, validation("repeat_threshold must be >= 0")
		}
		v := *in.RepeatThreshold
		thr = &v
	}
	var repLimit *int
	if in.RepeatLimit != nil {
		if *in.RepeatLimit < 0 {
			return issue.Issue{}, validation("repeat_limit must be >= 0")
		}
		v := *in.RepeatLimit
		repLimit = &v
	}
	if title == nil && thr == nil && repLimit == nil {
		return issue.Issue{}, validation("no fields to update")
	}

	// no duplicates on rename (trim + case-insensitive)
	if title != nil {
		current, err := s.iss.GetIssue(ctx, issueID)
		if err != nil {
			if err == repository.ErrNotFound {
				return issue.Issue{}, notFound("issue not found", err)
			}
			return issue.Issue{}, err
		}
		existing, err := s.iss.ListIssuesByPair(ctx, current.PairID, nil)
		if err != nil {
			return issue.Issue{}, err
		}
		for _, it := range existing {
			if it.ID == issueID {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(it.Title), strings.TrimSpace(*title)) {
				return issue.Issue{}, conflict("issue title already exists", nil)
			}
		}
	}

	it, err := s.iss.UpdateIssue(ctx, issueID, title, thr, repLimit)
	if err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("issue not found", err)
		}
		return issue.Issue{}, err
	}
	return it, nil
}

func (s *IssueService) Repeat(ctx context.Context, issueID, userID string, note *string) (issue.Issue, error) {
	if !validateUUID(issueID) {
		return issue.Issue{}, validation("invalid issue_id")
	}
	if !validateUUID(userID) {
		return issue.Issue{}, validation("invalid user_id")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("user not found", err)
		}
		return issue.Issue{}, err
	}
	it, err := s.iss.RepeatIssue(ctx, issueID, userID, note)
	if err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("issue not found", err)
		}
		if err == repository.ErrConflict {
			return issue.Issue{}, conflict("repeat limit reached", err)
		}
		return issue.Issue{}, err
	}
	return it, nil
}

func (s *IssueService) UpdateStatus(ctx context.Context, issueID string, status issue.Status) (issue.Issue, error) {
	if !validateUUID(issueID) {
		return issue.Issue{}, validation("invalid issue_id")
	}
	switch status {
	case issue.StatusActive, issue.StatusResolved, issue.StatusPostponed, issue.StatusArchived:
	default:
		return issue.Issue{}, validation("invalid status")
	}
	it, err := s.iss.UpdateStatus(ctx, issueID, status)
	if err != nil {
		if err == repository.ErrNotFound {
			return issue.Issue{}, notFound("issue not found", err)
		}
		return issue.Issue{}, err
	}
	return it, nil
}

func (s *IssueService) ListRepeats(ctx context.Context, issueID string, limit, offset int) ([]issue.IssueRepeat, error) {
	if !validateUUID(issueID) {
		return nil, validation("invalid issue_id")
	}
	if limit < 0 || offset < 0 {
		return nil, validation("invalid pagination")
	}
	if _, err := s.iss.GetIssue(ctx, issueID); err != nil {
		if err == repository.ErrNotFound {
			return nil, notFound("issue not found", err)
		}
		return nil, err
	}
	return s.iss.ListRepeatsByIssue(ctx, issueID, limit, offset)
}

func (s *IssueService) GetRepeat(ctx context.Context, repeatID string) (issue.IssueRepeat, error) {
	if !validateUUID(repeatID) {
		return issue.IssueRepeat{}, validation("invalid repeat_id")
	}
	rep, err := s.iss.GetRepeat(ctx, repeatID)
	if err != nil {
		if err == repository.ErrNotFound {
			return issue.IssueRepeat{}, notFound("repeat not found", err)
		}
		return issue.IssueRepeat{}, err
	}
	return rep, nil
}

func (s *IssueService) AddRepeatDisagreement(ctx context.Context, repeatID, userID, note string) (issue.IssueRepeatDisagreement, error) {
	if !validateUUID(repeatID) {
		return issue.IssueRepeatDisagreement{}, validation("invalid repeat_id")
	}
	if !validateUUID(userID) {
		return issue.IssueRepeatDisagreement{}, validation("invalid user_id")
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return issue.IssueRepeatDisagreement{}, validation("note is required")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return issue.IssueRepeatDisagreement{}, notFound("user not found", err)
		}
		return issue.IssueRepeatDisagreement{}, err
	}
	if _, err := s.iss.GetRepeat(ctx, repeatID); err != nil {
		if err == repository.ErrNotFound {
			return issue.IssueRepeatDisagreement{}, notFound("repeat not found", err)
		}
		return issue.IssueRepeatDisagreement{}, err
	}
	d, err := s.iss.CreateRepeatDisagreement(ctx, repeatID, userID, note)
	if err == repository.ErrConflict {
		return issue.IssueRepeatDisagreement{}, conflict("disagreement already exists", err)
	}
	if err != nil {
		return issue.IssueRepeatDisagreement{}, err
	}
	return d, nil
}

func (s *IssueService) DeleteIssue(ctx context.Context, issueID string) error {
	if !validateUUID(issueID) {
		return validation("invalid issue_id")
	}
	if err := s.iss.DeleteIssue(ctx, issueID); err != nil {
		if err == repository.ErrNotFound {
			return notFound("issue not found", err)
		}
		return err
	}
	return nil
}

func (s *IssueService) DeleteRepeat(ctx context.Context, repeatID string) error {
	if !validateUUID(repeatID) {
		return validation("invalid repeat_id")
	}
	if err := s.iss.DeleteRepeat(ctx, repeatID); err != nil {
		if err == repository.ErrNotFound {
			return notFound("repeat not found", err)
		}
		return err
	}
	return nil
}
