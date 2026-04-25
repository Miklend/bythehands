package telegram

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Dispatcher interface {
	HandleUpdate(ctx context.Context, u *tgbotapi.Update)
}

func Run(ctx context.Context, log *slog.Logger, bot *tgbotapi.BotAPI, d Dispatcher) error {
	ucfg := tgbotapi.NewUpdate(0)
	ucfg.Timeout = 30

	updates := bot.GetUpdatesChan(ucfg)
	log.Info("bot started", slog.String("username", bot.Self.UserName))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case upd := <-updates:
			d.HandleUpdate(ctx, &upd)
		}
	}
}
