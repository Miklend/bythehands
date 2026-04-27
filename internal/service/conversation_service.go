package service

import (
	"context"
	"strings"
	"time"

	"taalkbout/internal/domain/conversation"
	"taalkbout/internal/domain/issue"
	"taalkbout/internal/repository"
)

type ConversationService struct {
	users repository.UserRepository
	pairs repository.PairRepository
	iss   repository.IssueRepository
	conv  repository.ConversationRepository
	now   func() time.Time
}

type StartConversationInput struct {
	IssueID            string
	PairID             string
	Goal               *string
	Questions          *string
	StartState         *string
	RuleViolationLimit int
}

type FinishConversationInput struct {
	ResultStatus    conversation.ResultStatus
	ResultText      *string
	EndState        *string
	EndedEarly      bool
	EndedByUserID   *string
	EndedInitiative *string
	EndReason       *string
}

func NewConversationService(users repository.UserRepository, pairs repository.PairRepository, iss repository.IssueRepository, conv repository.ConversationRepository) *ConversationService {
	return &ConversationService{
		users: users,
		pairs: pairs,
		iss:   iss,
		conv:  conv,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

func (s *ConversationService) Start(ctx context.Context, in StartConversationInput) (conversation.Session, error) {
	if !validateUUID(in.IssueID) {
		return conversation.Session{}, validation("invalid issue_id")
	}
	if !validateUUID(in.PairID) {
		return conversation.Session{}, validation("invalid pair_id")
	}
	if _, err := s.pairs.GetPair(ctx, in.PairID); err != nil {
		if err == repository.ErrNotFound {
			return conversation.Session{}, notFound("pair not found", err)
		}
		return conversation.Session{}, err
	}
	it, err := s.iss.GetIssue(ctx, in.IssueID)
	if err != nil {
		if err == repository.ErrNotFound {
			return conversation.Session{}, notFound("issue not found", err)
		}
		return conversation.Session{}, err
	}
	if it.PairID != in.PairID {
		return conversation.Session{}, forbidden("issue does not belong to pair", nil)
	}

	var goal *string
	if in.Goal != nil {
		g := strings.TrimSpace(*in.Goal)
		if g != "" {
			goal = &g
		}
	}

	var questions *string
	if in.Questions != nil {
		q := strings.TrimSpace(*in.Questions)
		if q != "" {
			questions = &q
		}
	}
	var startState *string
	if in.StartState != nil {
		ss := strings.TrimSpace(*in.StartState)
		if ss != "" {
			startState = &ss
		}
	}
	if in.RuleViolationLimit < 0 {
		return conversation.Session{}, validation("rule_violation_limit must be >= 0")
	}

	return s.conv.StartSession(ctx, conversation.Session{
		IssueID:            in.IssueID,
		PairID:             in.PairID,
		Goal:               goal,
		Questions:          questions,
		StartState:         startState,
		RuleViolationLimit: in.RuleViolationLimit,
	})
}

func (s *ConversationService) Finish(ctx context.Context, conversationID string, in FinishConversationInput) (conversation.Session, error) {
	if !validateUUID(conversationID) {
		return conversation.Session{}, validation("invalid conversation_id")
	}
	switch in.ResultStatus {
	case conversation.ResultResolved, conversation.ResultPartiallyResolved, conversation.ResultPostponed, conversation.ResultUnresolved:
	default:
		return conversation.Session{}, validation("invalid result_status")
	}
	var text *string
	if in.ResultText != nil {
		t := strings.TrimSpace(*in.ResultText)
		if t != "" {
			text = &t
		}
	}
	var endState *string
	if in.EndState != nil {
		v := strings.TrimSpace(*in.EndState)
		if v != "" {
			endState = &v
		}
	}

	var endedBy *string
	var endedInitiative *string
	var endReason *string
	if in.EndedEarly {
		if in.EndedByUserID == nil || !validateUUID(*in.EndedByUserID) {
			return conversation.Session{}, validation("invalid ended_by_user_id")
		}
		if _, err := s.users.GetByID(ctx, *in.EndedByUserID); err != nil {
			if err == repository.ErrNotFound {
				return conversation.Session{}, notFound("user not found", err)
			}
			return conversation.Session{}, err
		}
		endedBy = in.EndedByUserID
		if in.EndedInitiative == nil {
			return conversation.Session{}, validation("ended_initiative is required")
		}
		switch strings.TrimSpace(*in.EndedInitiative) {
		case "self", "partner", "both":
		default:
			return conversation.Session{}, validation("invalid ended_initiative")
		}
		endedInitiative = in.EndedInitiative
		if in.EndReason != nil {
			r := strings.TrimSpace(*in.EndReason)
			if r == "" {
				return conversation.Session{}, validation("end_reason is required")
			}
			endReason = &r
		} else {
			return conversation.Session{}, validation("end_reason is required")
		}
	}

	sess, err := s.conv.FinishSession(ctx, conversationID, in.ResultStatus, text, endState, s.now(), in.EndedEarly, endedBy, endedInitiative, endReason)
	if err != nil {
		if err == repository.ErrNotFound {
			return conversation.Session{}, notFound("conversation not found", err)
		}
		return conversation.Session{}, err
	}
	return sess, nil
}

func (s *ConversationService) Get(ctx context.Context, conversationID string) (conversation.Session, error) {
	if !validateUUID(conversationID) {
		return conversation.Session{}, validation("invalid conversation_id")
	}
	sess, err := s.conv.GetSession(ctx, conversationID)
	if err != nil {
		if err == repository.ErrNotFound {
			return conversation.Session{}, notFound("conversation not found", err)
		}
		return conversation.Session{}, err
	}
	return sess, nil
}

func (s *ConversationService) Pause(ctx context.Context, conversationID string) (conversation.Session, error) {
	sess, err := s.Get(ctx, conversationID)
	if err != nil {
		return conversation.Session{}, err
	}
	if sess.Status != conversation.StatusStarted {
		return conversation.Session{}, conflict("conversation is not started", nil)
	}
	updated, err := s.conv.PauseSession(ctx, conversationID)
	if err != nil {
		return conversation.Session{}, err
	}
	return updated, nil
}

func (s *ConversationService) Resume(ctx context.Context, conversationID string) (conversation.Session, error) {
	sess, err := s.Get(ctx, conversationID)
	if err != nil {
		return conversation.Session{}, err
	}
	if sess.Status != conversation.StatusPaused {
		return conversation.Session{}, conflict("conversation is not paused", nil)
	}
	updated, err := s.conv.ResumeSession(ctx, conversationID)
	if err != nil {
		return conversation.Session{}, err
	}
	return updated, nil
}

func (s *ConversationService) ListByPair(ctx context.Context, pairID string, status *conversation.Status, limit, offset int) ([]conversation.Session, error) {
	if !validateUUID(pairID) {
		return nil, validation("invalid pair_id")
	}
	if status != nil {
		switch *status {
		case conversation.StatusStarted, conversation.StatusPaused, conversation.StatusFinished, conversation.StatusCancelled:
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
	return s.conv.ListByPair(ctx, pairID, status, limit, offset)
}

func (s *ConversationService) AddNote(ctx context.Context, conversationID, userID, text string) error {
	if !validateUUID(conversationID) {
		return validation("invalid conversation_id")
	}
	if !validateUUID(userID) {
		return validation("invalid user_id")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return validation("text is required")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return notFound("user not found", err)
		}
		return err
	}
	sess, err := s.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	if sess.Status == conversation.StatusFinished || sess.Status == conversation.StatusCancelled {
		return conflict("conversation is finished", nil)
	}
	return s.conv.AddNote(ctx, conversationID, userID, text)
}

func (s *ConversationService) ListNotes(ctx context.Context, conversationID string, limit, offset int) ([]conversation.Note, error) {
	if !validateUUID(conversationID) {
		return nil, validation("invalid conversation_id")
	}
	if limit < 0 || offset < 0 {
		return nil, validation("invalid pagination")
	}
	if _, err := s.conv.GetSession(ctx, conversationID); err != nil {
		if err == repository.ErrNotFound {
			return nil, notFound("conversation not found", err)
		}
		return nil, err
	}
	return s.conv.ListNotes(ctx, conversationID, limit, offset)
}

func (s *ConversationService) ListSideIssues(ctx context.Context, conversationID string) ([]issue.Issue, error) {
	if !validateUUID(conversationID) {
		return nil, validation("invalid conversation_id")
	}
	if _, err := s.conv.GetSession(ctx, conversationID); err != nil {
		if err == repository.ErrNotFound {
			return nil, notFound("conversation not found", err)
		}
		return nil, err
	}
	return s.conv.ListSideIssues(ctx, conversationID)
}

func (s *ConversationService) ListNotesByPair(ctx context.Context, pairID string, limit, offset int) ([]conversation.PairNote, error) {
	if !validateUUID(pairID) {
		return nil, validation("invalid pair_id")
	}
	if limit < 0 || offset < 0 {
		return nil, validation("invalid pagination")
	}
	if _, err := s.pairs.GetPair(ctx, pairID); err != nil {
		if err == repository.ErrNotFound {
			return nil, notFound("pair not found", err)
		}
		return nil, err
	}
	return s.conv.ListNotesByPair(ctx, pairID, limit, offset)
}

func (s *ConversationService) DeleteNote(ctx context.Context, noteID string) error {
	if !validateUUID(noteID) {
		return validation("invalid note_id")
	}
	if err := s.conv.DeleteNote(ctx, noteID); err != nil {
		if err == repository.ErrNotFound {
			return notFound("note not found", err)
		}
		return err
	}
	return nil
}

type AddSideIssueInput struct {
	ConversationID  string
	CreatedByUserID string
	Title           string
	Description     string
}

func (s *ConversationService) AddRuleViolation(ctx context.Context, conversationID, userID, ruleCode, note string) error {
	if !validateUUID(conversationID) {
		return validation("invalid conversation_id")
	}
	if !validateUUID(userID) {
		return validation("invalid user_id")
	}
	ruleCode = strings.TrimSpace(ruleCode)
	if ruleCode == "" {
		return validation("rule_code is required")
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return validation("note is required")
	}
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		if err == repository.ErrNotFound {
			return notFound("user not found", err)
		}
		return err
	}
	sess, err := s.Get(ctx, conversationID)
	if err != nil {
		return err
	}
	if sess.Status == conversation.StatusFinished || sess.Status == conversation.StatusCancelled {
		return conflict("conversation is finished", nil)
	}
	if sess.RuleViolationLimit > 0 {
		cnt, err := s.conv.CountRuleViolations(ctx, conversationID)
		if err != nil {
			return err
		}
		if cnt >= sess.RuleViolationLimit {
			return conflict("rule violations limit reached", nil)
		}
	}
	if err := s.conv.AddRuleViolation(ctx, conversationID, userID, ruleCode, note); err != nil {
		return err
	}
	return nil
}

func (s *ConversationService) ListRuleViolations(ctx context.Context, conversationID string, limit, offset int) ([]conversation.RuleViolation, error) {
	if !validateUUID(conversationID) {
		return nil, validation("invalid conversation_id")
	}
	if limit < 0 || offset < 0 {
		return nil, validation("invalid pagination")
	}
	if _, err := s.Get(ctx, conversationID); err != nil {
		return nil, err
	}
	return s.conv.ListRuleViolations(ctx, conversationID, limit, offset)
}

func (s *ConversationService) AddSideIssue(ctx context.Context, in AddSideIssueInput) (string, error) {
	if !validateUUID(in.ConversationID) {
		return "", validation("invalid conversation_id")
	}
	if !validateUUID(in.CreatedByUserID) {
		return "", validation("invalid created_by_user_id")
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if in.Title == "" {
		return "", validation("title is required")
	}
	if in.Description == "" {
		return "", validation("description is required")
	}
	if _, err := s.users.GetByID(ctx, in.CreatedByUserID); err != nil {
		if err == repository.ErrNotFound {
			return "", notFound("user not found", err)
		}
		return "", err
	}

	sess, err := s.Get(ctx, in.ConversationID)
	if err != nil {
		return "", err
	}
	if sess.Status == conversation.StatusFinished || sess.Status == conversation.StatusCancelled {
		return "", conflict("conversation is finished", nil)
	}

	newIssue, err := s.iss.CreateIssue(ctx, issue.Issue{
		PairID:          sess.PairID,
		CreatedByUserID: in.CreatedByUserID,
		Title:           in.Title,
		Description:     in.Description,
		Priority:        issue.PriorityMedium,
		Visibility:      issue.VisibilityVisible,
		RepeatThreshold: 0,
	})
	if err != nil {
		return "", err
	}
	if err := s.conv.LinkSideIssue(ctx, in.ConversationID, newIssue.ID); err != nil {
		return "", err
	}
	return newIssue.ID, nil
}
