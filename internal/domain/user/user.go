package user

import "time"

type User struct {
	ID          string    `json:"id"`
	TelegramID  *int64    `json:"telegram_id,omitempty"`
	Username    *string   `json:"username,omitempty"`
	DisplayName *string   `json:"display_name,omitempty"`
	IsVirtual   bool      `json:"is_virtual"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
