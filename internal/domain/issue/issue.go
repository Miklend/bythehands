package issue

import "time"

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type Visibility string

const (
	VisibilityVisible           Visibility = "visible"
	VisibilityHiddenUntilRepeat Visibility = "hidden_until_repeats"
	VisibilityPrivate           Visibility = "private"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusResolved  Status = "resolved"
	StatusPostponed Status = "postponed"
	StatusArchived  Status = "archived"
)

type Issue struct {
	ID              string     `json:"id"`
	PairID          string     `json:"pair_id"`
	CreatedByUserID string     `json:"created_by_user_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Priority        Priority   `json:"priority"`
	Visibility      Visibility `json:"visibility"`
	RepeatThreshold int        `json:"repeat_threshold"`
	RepeatCount     int        `json:"repeat_count"`
	Status          Status     `json:"status"`
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
