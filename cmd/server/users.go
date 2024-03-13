package main

import (
	"context"
	"errors"

	connect "connectrpc.com/connect"
	database "github.com/brGuirra/uai/internal/database/sqlc"
	userv1 "github.com/brGuirra/uai/internal/gen/user/v1"
	"github.com/brGuirra/uai/internal/token"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

var roleCodesMapper = map[userv1.Role]string{
	userv1.Role_ROLE_STAFF:    "staff",
	userv1.Role_ROLE_EMPLOYEE: "employee",
	userv1.Role_ROLE_LEADER:   "leader",
}

func (s *server) AddUser(ctx context.Context, req *connect.Request[userv1.AddUserRequest]) (*connect.Response[emptypb.Empty], error) {
	var userID uuid.UUID
	err := s.store.ExecTx(context.Background(), func(q database.Querier) error {
		var err error
		userID, err = q.CreateUser(ctx, database.CreateUserParams{
			Name:  req.Msg.Name,
			Email: req.Msg.Email,
		})
		if err != nil {
			return err
		}

		codes := make([]string, len(req.Msg.Roles))
		for i, role := range req.Msg.Roles {
			codes[i] = roleCodesMapper[role]
		}

		roleIDS, err := q.GetRolesByCodes(ctx, codes)
		if err != nil {
			return err
		}

		args := make([]database.AddRolesForUserParams, len(roleIDS))
		for i, roleID := range roleIDS {
			args[i] = database.AddRolesForUserParams{
				UserID:  userID,
				Grantor: userID,
				RoleID:  roleID,
			}
		}

		_, err = q.AddRolesForUser(ctx, args)
		if err != nil {
			return err
		}

		return nil
	})
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

		err = s.mailer.Send(req.Msg.Email, data, "welcome.tmpl")
		if err != nil {
			s.logger.Error("error sending welcome email", "cause", err)
		}
	})

	return connect.NewResponse(&emptypb.Empty{}), nil
}
