package main

import (
	"context"

	connect "connectrpc.com/connect"
	database "github.com/brGuirra/uai/internal/database/sqlc"
	userv1 "github.com/brGuirra/uai/internal/gen/user/v1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func (us *uaiServer) AddUser(ctx context.Context, req *connect.Request[userv1.AddUserRequest]) (*connect.Response[emptypb.Empty], error) {
	args := database.CreateUserParams{
		Name:  req.Msg.Name,
		Email: req.Msg.Email,
	}

	err := us.querier.CreateUser(ctx, args)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
