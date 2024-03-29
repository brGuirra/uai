package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/brGuirra/uai/internal/config"
	"github.com/brGuirra/uai/internal/password"
	"github.com/brGuirra/uai/internal/smtp"
	"github.com/brGuirra/uai/internal/token"

	database "github.com/brGuirra/uai/internal/database/sqlc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	var configFile string
	flag.StringVar(&configFile, "config-file", "pkl/DevelopmentConfig.pkl", "The path to the pkl config file")

	flag.Parse()

	err := run(logger, configFile)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger, configFile string) error {
	cfg, err := config.LoadFromPath(context.Background(), configFile)
	if err != nil {
		return err
	}

	store, err := database.NewStore(cfg.Db.Dsn)
	if err != nil {
		return err
	}

	mailer, err := smtp.NewMailtrapMailer(cfg.Smtp.Host, int(cfg.Smtp.Port), cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.From)
	if err != nil {
		return err
	}

	tokenMaker, err := token.NewPasetoMaker(cfg.Token.SimmetricKey)
	if err != nil {
		return err
	}

	passwordHasher := password.NewPasswordHasher()

	hashedPassword, err := passwordHasher.Hash(cfg.DefaultUser.Password)
	if err != nil {
		return fmt.Errorf("failed to hash intial user password: %w", err)
	}

	err = database.CreateInitialAdminUser(store, cfg.DefaultUser.Name, cfg.DefaultUser.Email, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to create initial user: %w", err)
	}

	s := NewServer(cfg, tokenMaker, store, logger, mailer, passwordHasher)

	return s.serve()
}
