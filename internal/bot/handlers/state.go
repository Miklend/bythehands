package handlers

import (
	"sync"

	"talkabout/internal/bot/client"
)

type State struct {
	mu    sync.RWMutex
	users map[int64]client.User // telegram user id -> api user
}

func NewState() *State {
	return &State{users: make(map[int64]client.User)}
}

func (s *State) SetUser(telegramUserID int64, u client.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[telegramUserID] = u
}

func (s *State) GetUser(telegramUserID int64) (client.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[telegramUserID]
	return u, ok
}
