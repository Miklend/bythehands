package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"talkabout/internal/bot/client"
	"talkabout/internal/bot/handlers"
	"talkabout/internal/bot/telegram"
	"talkabout/internal/config"
	"talkabout/internal/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	log := logger.New(logger.Config{Env: cfg.Env})
	slog.SetDefault(log)

	if strings.TrimSpace(cfg.Bot.TelegramToken) == "" || cfg.Bot.TelegramToken == "change_me" {
		log.Error("TELEGRAM_BOT_TOKEN is required")
		os.Exit(1)
	}

	apiClient, err := client.New(cfg.Bot.APIBaseURL, cfg.Bot.HTTPTimeout)
	if err != nil {
		log.Error("api client init failed", slog.Any("err", err))
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Bot.TelegramToken)
	if err != nil {
		log.Error("telegram init failed", slog.Any("err", err))
		os.Exit(1)
	}

	state := handlers.NewState()
	_ = state
	app := handlers.NewApp(log, bot, apiClient)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		cancel()
	}()

	if err := telegram.Run(ctx, log, bot, app); err != nil && err != context.Canceled {
		log.Error("bot stopped", slog.Any("err", err))
		os.Exit(1)
	}
}
