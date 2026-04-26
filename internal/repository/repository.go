package repository

import (
	"context"
	"time"

	"taalkbout/internal/domain/conversation"
	"taalkbout/internal/domain/issue"
	"taalkbout/internal/domain/pair"
	"taalkbout/internal/domain/user"
)

type UserRepository interface {
	UpsertTelegramUser(ctx context.Context, telegramID int64, username, displayName *string) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	CreateVirtualUser(ctx context.Context, username, displayName string) (user.User, error)
}

type PairRepository interface {
	CreatePair(ctx context.Context, isTest bool) (pair.Pair, error)
	AddMember(ctx context.Context, pairID, userID string, role pair.Role) (pair.PairMember, error)
	GetPair(ctx context.Context, pairID string) (pair.Pair, error)
	GetMembers(ctx context.Context, pairID string) ([]pair.PairMember, error)
	ListPairsByUser(ctx context.Context, userID string) ([]pair.Pair, error)
	ArchivePair(ctx context.Context, pairID string) (pair.Pair, error)
	SetWelcomeMessage(ctx context.Context, pairID string, text *string) (pair.Pair, error)
}

type InviteRepository interface {
	CreateInvite(ctx context.Context, pairID, token string, expiresAt time.Time) (pair.Invite, error)
	GetByToken(ctx context.Context, token string) (pair.Invite, error)
	MarkUsed(ctx context.Context, inviteID string, usedAt time.Time) (pair.Invite, error)
	MarkExpired(ctx context.Context, inviteID string) (pair.Invite, error)
}

type IssueRepository interface {
	CreateIssue(ctx context.Context, in issue.Issue) (issue.Issue, error)
	ListIssuesByPair(ctx context.Context, pairID string, status *issue.Status) ([]issue.Issue, error)
	GetIssue(ctx context.Context, issueID string) (issue.Issue, error)
	RepeatIssue(ctx context.Context, issueID, userID string, note *string) (issue.Issue, error)
	IncrementRepeat(ctx context.Context, issueID string) (issue.Issue, error)
	CreateRepeat(ctx context.Context, issueID, userID string, note *string) (issue.IssueRepeat, error)
	ListRepeatsByIssue(ctx context.Context, issueID string, limit, offset int) ([]issue.IssueRepeat, error)
	GetRepeat(ctx context.Context, repeatID string) (issue.IssueRepeat, error)
	CreateRepeatDisagreement(ctx context.Context, repeatID, userID string, note string) (issue.IssueRepeatDisagreement, error)
	GetRepeatDisagreement(ctx context.Context, repeatID, userID string) (issue.IssueRepeatDisagreement, error)
	UpdateIssue(ctx context.Context, issueID string, title *string, repeatThreshold *int) (issue.Issue, error)
	UpdateStatus(ctx context.Context, issueID string, status issue.Status) (issue.Issue, error)
	DeleteIssue(ctx context.Context, issueID string) error
	DeleteRepeat(ctx context.Context, repeatID string) error
}

type ConversationRepository interface {
	StartSession(ctx context.Context, in conversation.Session) (conversation.Session, error)
	FinishSession(ctx context.Context, id string, resultStatus conversation.ResultStatus, resultText *string, endState *string, finishedAt time.Time, endedEarly bool, endedByUserID *string, endReason *string) (conversation.Session, error)
	GetSession(ctx context.Context, id string) (conversation.Session, error)
	PauseSession(ctx context.Context, id string) (conversation.Session, error)
	ResumeSession(ctx context.Context, id string) (conversation.Session, error)
	ListByPair(ctx context.Context, pairID string, status *conversation.Status, limit, offset int) ([]conversation.Session, error)
	AddNote(ctx context.Context, conversationID, userID, text string) error
	LinkSideIssue(ctx context.Context, conversationID, issueID string) error
	ListNotes(ctx context.Context, conversationID string, limit, offset int) ([]conversation.Note, error)
	ListSideIssues(ctx context.Context, conversationID string) ([]issue.Issue, error)
	ListNotesByPair(ctx context.Context, pairID string, limit, offset int) ([]conversation.PairNote, error)
	DeleteNote(ctx context.Context, noteID string) error
}

type PreferencesRepository interface {
	GetPreferences(ctx context.Context, userID string) (currentPairID *string, err error)
	SetCurrentPair(ctx context.Context, userID string, pairID *string) error
}
