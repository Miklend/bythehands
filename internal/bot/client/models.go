package client

import "time"

type Envelope[T any] struct {
	Data  T    `json:"data"`
	Error *Err `json:"error,omitempty"`
}

type Err struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type User struct {
	ID          string    `json:"id"`
	TelegramID  *int64    `json:"telegram_id,omitempty"`
	Username    *string   `json:"username,omitempty"`
	DisplayName *string   `json:"display_name,omitempty"`
	IsVirtual   bool      `json:"is_virtual"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Pair struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	IsTest         bool      `json:"is_test"`
	WelcomeMessage *string   `json:"welcome_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PairMember struct {
	ID        string    `json:"id"`
	PairID    string    `json:"pair_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePairResult struct {
	Pair        Pair   `json:"pair"`
	InviteToken string `json:"invite_token"`
}

type Preferences struct {
	CurrentPairID *string `json:"current_pair_id"`
}

type Issue struct {
	ID              string     `json:"id"`
	PairID          string     `json:"pair_id"`
	CreatedByUserID string     `json:"created_by_user_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Priority        string     `json:"priority"`
	Visibility      string     `json:"visibility"`
	RepeatThreshold int        `json:"repeat_threshold"`
	RepeatCount     int        `json:"repeat_count"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

type IssueRepeat struct {
	ID        string    `json:"id"`
	IssueID   string    `json:"issue_id"`
	UserID    string    `json:"user_id"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type IssueRepeatDisagreement struct {
	ID        string    `json:"id"`
	RepeatID  string    `json:"repeat_id"`
	UserID    string    `json:"user_id"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type ConversationSession struct {
	ID           string     `json:"id"`
	IssueID      string     `json:"issue_id"`
	PairID       string     `json:"pair_id"`
	Status       string     `json:"status"`
	Goal         *string    `json:"goal,omitempty"`
	Questions    *string    `json:"questions,omitempty"`
	StartState   *string    `json:"start_state,omitempty"`
	EndState     *string    `json:"end_state,omitempty"`
	ResultStatus *string    `json:"result_status,omitempty"`
	ResultText   *string    `json:"result_text,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

type ConversationNote struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	UserID         string    `json:"user_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConversationPairNote struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	IssueID        string    `json:"issue_id"`
	IssueTitle     string    `json:"issue_title"`
	UserID         string    `json:"user_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

type TestModeStartResult struct {
	Pair           Pair `json:"pair"`
	VirtualPartner User `json:"virtual_partner"`
}

type TestModeStopResult struct {
	ArchivedPair  Pair    `json:"archived_pair"`
	CurrentPairID *string `json:"current_pair_id"`
	CurrentIsTest bool    `json:"current_is_test"`
	CurrentPair   *Pair   `json:"current_pair,omitempty"`
}
