package conversation

import "time"

type Status string

const (
	StatusStarted   Status = "started"
	StatusPaused    Status = "paused"
	StatusFinished  Status = "finished"
	StatusCancelled Status = "cancelled"
)

type ResultStatus string

const (
	ResultResolved          ResultStatus = "resolved"
	ResultPartiallyResolved ResultStatus = "partially_resolved"
	ResultPostponed         ResultStatus = "postponed"
	ResultUnresolved        ResultStatus = "unresolved"
)

type Session struct {
	ID                 string        `json:"id"`
	IssueID            string        `json:"issue_id"`
	PairID             string        `json:"pair_id"`
	Status             Status        `json:"status"`
	Goal               *string       `json:"goal,omitempty"`
	Questions          *string       `json:"questions,omitempty"`
	StartState         *string       `json:"start_state,omitempty"`
	RuleViolationLimit int           `json:"rule_violation_limit"`
	EndState           *string       `json:"end_state,omitempty"`
	ResultStatus       *ResultStatus `json:"result_status,omitempty"`
	ResultText         *string       `json:"result_text,omitempty"`
	EndedEarly         bool          `json:"ended_early"`
	EndedByUserID      *string       `json:"ended_by_user_id,omitempty"`
	EndedInitiative    *string       `json:"ended_initiative,omitempty"`
	EndReason          *string       `json:"end_reason,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	FinishedAt         *time.Time    `json:"finished_at,omitempty"`
}

type Note struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	UserID         string    `json:"user_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

type PairNote struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	IssueID        string    `json:"issue_id"`
	IssueTitle     string    `json:"issue_title"`
	UserID         string    `json:"user_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

type RuleViolation struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	UserID         string    `json:"user_id"`
	RuleCode       string    `json:"rule_code"`
	Note           string    `json:"note"`
	CreatedAt      time.Time `json:"created_at"`
}
