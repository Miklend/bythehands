package pair

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

type Role string

const (
	RoleCreator Role = "creator"
	RolePartner Role = "partner"
)

type Pair struct {
	ID             string    `json:"id"`
	Status         Status    `json:"status"`
	IsTest         bool      `json:"is_test"`
	WelcomeMessage *string   `json:"welcome_message,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PairMember struct {
	ID         string    `json:"id"`
	PairID     string    `json:"pair_id"`
	UserID     string    `json:"user_id"`
	Role       Role      `json:"role"`
	CustomName *string   `json:"custom_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type InviteStatus string

const (
	InviteActive  InviteStatus = "active"
	InviteUsed    InviteStatus = "used"
	InviteExpired InviteStatus = "expired"
)

type Invite struct {
	ID        string       `json:"id"`
	PairID    string       `json:"pair_id"`
	Token     string       `json:"token"`
	Status    InviteStatus `json:"status"`
	ExpiresAt time.Time    `json:"expires_at"`
	CreatedAt time.Time    `json:"created_at"`
	UsedAt    *time.Time   `json:"used_at,omitempty"`
}
