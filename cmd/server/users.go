package main

import (
	"context"
	"errors"

	connect "connectrpc.com/connect"
	database "github.com/brGuirra/uai/internal/database/sqlc"
	userv1 "github.com/brGuirra/uai/internal/gen/user/v1"
	"github.com/brGuirra/uai/internal/token"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (s *server) AddUser(ctx context.Context, req *connect.Request[userv1.AddUserRequest]) (*connect.Response[emptypb.Empty], error) {
	args := database.CreateUserParams{
		Name:  req.Msg.Name,
		Email: req.Msg.Email,
	}

	userID, err := s.store.CreateUser(ctx, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("email already in use"))
			}
		}

		s.logger.Error("error creating user", "cause", err)
		return nil, connect.NewError(connect.CodeInternal, errors.New("internal server error"))

	}

	token := s.tokenMaker.CreateToken(userID.String(), token.ScopeActivation)

	s.background(func() {
		data := map[string]any{
			"activationToken": token,
			"userID":          userID,
		}

		err = s.mailer.Send(args.Email, data, "welcome.tmpl")
		if err != nil {
			s.logger.Error("error sending welcome email", "cause", err)
		}
	})

	return connect.NewResponse(&emptypb.Empty{}), nil
}
