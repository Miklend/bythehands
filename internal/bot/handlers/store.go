package handlers

import (
	"sync"
	"time"
)

type BotState string

const (
	StateIdle                    BotState = "idle"
	StateAddIssueTitle           BotState = "add_issue_title"
	StateAddIssueDescription     BotState = "add_issue_description"
	StateAddIssuePriority        BotState = "add_issue_priority"
	StateAddIssueVisibility      BotState = "add_issue_visibility"
	StateAddIssueThreshold       BotState = "add_issue_repeat_threshold"
	StateAddRepeatNote           BotState = "add_repeat_note"
	StateAddRepeatForwardComment BotState = "add_repeat_forward_comment"
	StateAddRepeatDisagreeNote   BotState = "add_repeat_disagree_note"

	StateFocusSelectIssue       BotState = "focus_select_issue"
	StateFocusGoal              BotState = "focus_goal"
	StateFocusQuestions         BotState = "focus_questions"
	StateFocusStartStateSelf    BotState = "focus_start_state_self"
	StateFocusStartStatePartner BotState = "focus_start_state_partner"
	StateFocusPlan              BotState = "focus_plan"

	StateConvNote           BotState = "conv_note"
	StateConvForwardComment BotState = "conv_forward_comment"
	StateConvSideTitle      BotState = "conv_side_title"
	StateConvSideDesc       BotState = "conv_side_desc"

	StateFinishStatus          BotState = "finish_status"
	StateFinishEndStateSelf    BotState = "finish_end_state_self"
	StateFinishEndStatePartner BotState = "finish_end_state_partner"
	StateFinishText            BotState = "finish_text"

	StatePairWelcome BotState = "pair_welcome"
)

type Session struct {
	TelegramUserID int64
	ChatID         int64

	APIUserID string

	CurrentPairID     *string
	CurrentPairIsTest bool

	State BotState

	PendingInviteToken string

	CurrentIssueID     string
	CurrentRepeatID    string
	PendingForwardText string

	AddIssueTitle       string
	AddIssueDescription string
	AddIssuePriority    string
	AddIssueVisibility  string
	AddIssueThreshold   int

	FocusIssueID           string
	FocusGoal              string
	FocusQuestions         string
	FocusStartStateSelf    string
	FocusStartStatePartner string

	ConversationID        string
	ConversationIssueID   string
	ConversationStartedAt time.Time

	SideIssueTitle string

	FinishResultStatus    string
	FinishEndStateSelf    string
	FinishEndStatePartner string

	QuickIssueTitle string
}

type MemoryStore struct {
	mu   sync.RWMutex
	sess map[int64]*Session
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sess: make(map[int64]*Session)}
}

func (s *MemoryStore) GetOrCreate(telegramUserID int64, chatID int64) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.sess[telegramUserID]; ok {
		v.ChatID = chatID
		return v
	}
	v := &Session{
		TelegramUserID: telegramUserID,
		ChatID:         chatID,
		State:          StateIdle,
	}
	s.sess[telegramUserID] = v
	return v
}

func (s *MemoryStore) Clear(telegramUserID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sess, telegramUserID)
}
