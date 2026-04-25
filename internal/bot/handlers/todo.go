package handlers

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TODOHandler struct {
	bot     *tgbotapi.BotAPI
	command string
	text    string
}

func NewTODOHandler(bot *tgbotapi.BotAPI, command string) *TODOHandler {
	return &TODOHandler{
		bot:     bot,
		command: command,
		text:    "Команда в разработке. Сценарии бота будут добавлены позже.",
	}
}

func (h *TODOHandler) Command() string { return h.command }

func (h *TODOHandler) Handle(ctx context.Context, u *tgbotapi.Update) error {
	if u.Message == nil {
		return nil
	}
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, h.text)
	_, err := h.bot.Send(msg)
	return err
}
