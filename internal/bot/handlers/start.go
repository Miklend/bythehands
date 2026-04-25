package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"taalkbout/internal/bot/client"
)

type StartHandler struct {
	log   *slog.Logger
	bot   *tgbotapi.BotAPI
	api   *client.Client
	state *State
}

func NewStartHandler(log *slog.Logger, bot *tgbotapi.BotAPI, api *client.Client, state *State) *StartHandler {
	return &StartHandler{log: log, bot: bot, api: api, state: state}
}

func (h *StartHandler) Command() string { return "start" }

func (h *StartHandler) Handle(ctx context.Context, u *tgbotapi.Update) error {
	if u.Message == nil || u.Message.From == nil {
		return nil
	}

	tgUser := u.Message.From
	display := strings.TrimSpace(strings.Join([]string{tgUser.FirstName, tgUser.LastName}, " "))
	if display == "" {
		display = tgUser.UserName
	}

	var username *string
	if strings.TrimSpace(tgUser.UserName) != "" {
		v := strings.TrimSpace(tgUser.UserName)
		username = &v
	}
	var displayName *string
	if strings.TrimSpace(display) != "" {
		v := display
		displayName = &v
	}

	apiUser, err := h.api.UpsertTelegramUser(ctx, client.UpsertTelegramUserRequest{
		TelegramID:  tgUser.ID,
		Username:    username,
		DisplayName: displayName,
	})
	if err != nil {
		h.log.Error("api upsert telegram user failed", slog.Any("err", err))
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Не получилось связаться с API. Попробуй позже.")
		_, _ = h.bot.Send(msg)
		return nil
	}
	h.state.SetUser(tgUser.ID, apiUser)

	text := "Привет! TalkaBot помогает фиксировать важные темы и обсуждать их спокойнее. Сценарии бота будут добавлены позже."
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, fmt.Sprintf("%s\n\nТвой user_id: %s", text, apiUser.ID))
	_, err = h.bot.Send(msg)
	return err
}
