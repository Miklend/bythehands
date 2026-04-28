package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"talkabout/internal/api/routes"
	"talkabout/internal/config"
	"talkabout/internal/database"
	"talkabout/internal/logger"
	"talkabout/internal/repository/postgres"
	"talkabout/internal/service"
	"time"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	log := logger.New(logger.Config{Env: cfg.Env})
	slog.SetDefault(log)

	pool, err := database.OpenPostgres(ctx, database.PostgresConfig{
		DatabaseURL: cfg.DB.DatabaseURL,
		MaxConns:    cfg.DB.MaxConns,
	})
	if err != nil {
		log.Error("db open failed", slog.Any("err", err))
		os.Exit(1)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepo(pool)
	pairRepo := postgres.NewPairRepo(pool)
	invRepo := postgres.NewInviteRepo(pool)
	issueRepo := postgres.NewIssueRepo(pool)
	convRepo := postgres.NewConversationRepo(pool)
	prefRepo := postgres.NewPreferencesRepo(pool)

	usersSvc := service.NewUserService(userRepo)
	pairsSvc := service.NewPairService(userRepo, pairRepo, invRepo)
	issuesSvc := service.NewIssueService(userRepo, pairRepo, issueRepo)
	convSvc := service.NewConversationService(userRepo, pairRepo, issueRepo, convRepo)
	prefSvc := service.NewPreferencesService(userRepo, pairRepo, prefRepo)
	testSvc := service.NewTestModeService(userRepo, pairRepo, prefRepo)

	router := routes.NewRouter(log, routes.Services{
		Users:         usersSvc,
		Pairs:         pairsSvc,
		Issues:        issuesSvc,
		Conversations: convSvc,
		Preferences:   prefSvc,
		TestMode:      testSvc,
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.API.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("api listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error ", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("shutdown complete")
}
