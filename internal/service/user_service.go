package service

import (
	"context"

	"taalkbout/internal/domain/user"
	"taalkbout/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) UpsertTelegramUser(ctx context.Context, telegramID int64, username, displayName *string) (user.User, error) {
	if telegramID <= 0 {
		return user.User{}, validation("telegram_id must be > 0")
	}
	u, err := s.users.UpsertTelegramUser(ctx, telegramID, username, displayName)
	if err != nil {
		return user.User{}, err
	}
	return u, nil
}
