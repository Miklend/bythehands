package handlers

import (
	"context"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler interface {
	Command() string
	Handle(ctx context.Context, u *tgbotapi.Update) error
}

type Dispatcher struct {
	log      *slog.Logger
	handlers map[string]CommandHandler
}

func NewDispatcher(log *slog.Logger, hs ...CommandHandler) *Dispatcher {
	m := make(map[string]CommandHandler, len(hs))
	for _, h := range hs {
		m[h.Command()] = h
	}
	return &Dispatcher{log: log, handlers: m}
}

func (d *Dispatcher) Dispatch(ctx context.Context, u *tgbotapi.Update) bool {
	if u == nil || u.Message == nil || !u.Message.IsCommand() {
		return false
	}
	cmd := strings.TrimPrefix(u.Message.Command(), "/")
	h, ok := d.handlers[cmd]
	if !ok {
		return false
	}
	if err := h.Handle(ctx, u); err != nil {
		d.log.Error("command failed", slog.String("cmd", cmd), slog.Any("err", err))
	}
	return true
}
