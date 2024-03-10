package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/brGuirra/uai/internal/config"
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

	s := NewServer(cfg, tokenMaker, store, logger, mailer)

	s.logger.Info(
		"config",
		"port", cfg.Port,
		"env", cfg.Env,
	)

	return s.serve()
}
