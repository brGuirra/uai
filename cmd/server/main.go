package main

import (
	"flag"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/brGuirra/uai/internal/gen/user/v1/userv1connect"
	"github.com/brGuirra/uai/internal/smtp"

	database "github.com/brGuirra/uai/internal/database/sqlc"
)

type config struct {
	port int
	env  string
	cors struct{ trustedOrigins []string }
	db   struct {
		dsn string
	}
	jwt struct {
		secretKey string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
}

type uaiServer struct {
	config  config
	querier database.Querier
	logger  *slog.Logger
	mailer  *smtp.Mailer
	wg      sync.WaitGroup
	userv1connect.UnimplementedUserServiceHandler
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.env, "env", "development", "environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "user:pass@localhost:5432/db", "postgreSQL DSN")

	flag.StringVar(&cfg.jwt.secretKey, "jwt-secret-key", "ccw3wg3ombip3656l672bgwm3svsz7sh", "secret key for JWT authentication")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "example.smtp.host", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "smtp port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "example_username", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "pa55word", "smtp password")
	flag.StringVar(&cfg.smtp.from, "smtp-from", "Example Name <no-reply@example.org>", "smtp sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	store, err := database.NewStore(cfg.db.dsn)
	if err != nil {
		return err
	}

	mailer, err := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)
	if err != nil {
		return err
	}

	us := &uaiServer{
		config:  cfg,
		querier: store,
		logger:  logger,
		mailer:  mailer,
	}

	us.logger.Debug(
		"config",
		"port", cfg.port,
		"env", cfg.env,
	)

	return us.serve()
}
