//go:build integration
// +build integration

package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	validatepb "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	connect "connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/brGuirra/uai/internal/config"
	"github.com/brGuirra/uai/internal/config/environment"
	database "github.com/brGuirra/uai/internal/database/sqlc"
	userv1 "github.com/brGuirra/uai/internal/gen/user/v1"
	"github.com/brGuirra/uai/internal/gen/user/v1/userv1connect"
	"github.com/brGuirra/uai/internal/token"
	dbMock "github.com/brGuirra/uai/mocks/database"
	passwordMock "github.com/brGuirra/uai/mocks/password"
	smptMock "github.com/brGuirra/uai/mocks/smtp"
	tokenMock "github.com/brGuirra/uai/mocks/token"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

func setupTestServer(interceptor *validate.Interceptor, s *server) *httptest.Server {
	mux := http.NewServeMux()

	mux.Handle(userv1connect.NewUserServiceHandler(s, connect.WithInterceptors(interceptor)))

	server := httptest.NewUnstartedServer(mux)
	server.EnableHTTP2 = true
	server.StartTLS()

	return server
}

func setupClient(srv *httptest.Server) userv1connect.UserServiceClient {
	return userv1connect.NewUserServiceClient(srv.Client(), srv.URL)
}

func TestAddUser(t *testing.T) {
	testCases := []struct {
		name          string
		req           *userv1.AddUserRequest
		buildStubs    func(req *userv1.AddUserRequest, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer)
		checkMocks    func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer)
		checkResponse func(t *testing.T, res *connect.Response[emptypb.Empty], err error)
	}{
		{
			name: "Success",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{userv1.Role_ROLE_EMPLOYEE},
			},
			buildStubs: func(req *userv1.AddUserRequest, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				fakeUserID := uuid.MustParse(gofakeit.UUID())
				storeMock.EXPECT().CreateUser(
					mock.Anything,
					database.CreateUserParams{Name: req.Name, Email: req.Email}).
					Return(fakeUserID, nil)

				fakeRoleIDS := []uuid.UUID{uuid.MustParse(gofakeit.UUID())}
				storeMock.EXPECT().GetRolesByCodes(mock.Anything, []string{"employee"}).Return(fakeRoleIDS, nil)

				storeMock.EXPECT().AddRolesForUser(mock.Anything, []database.AddRolesForUserParams{
					{
						UserID:  fakeUserID,
						Grantor: fakeUserID,
						RoleID:  fakeRoleIDS[0],
					},
				}).Return(1, nil)

				storeMock.EXPECT().ExecTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
					return fn(storeMock)
				})

				fakeToken := gofakeit.Word()
				makerMock.EXPECT().CreateToken(fakeUserID.String(), token.ScopeActivation).Return(fakeToken)

				data := map[string]any{
					"activationToken": fakeToken,
					"userID":          fakeUserID,
				}
				mailerMock.EXPECT().Send(req.Email, data, "welcome.tmpl").Return(nil)
			},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.NoError(t, err)
				require.IsType(t, &emptypb.Empty{}, res.Msg)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 1)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 1)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 1)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 1)
				mailerMock.AssertNumberOfCalls(t, "Send", 1)
			},
		},
		{
			name: "Duplicate Email",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{userv1.Role_ROLE_EMPLOYEE},
			},
			buildStubs: func(req *userv1.AddUserRequest, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.EXPECT().CreateUser(
					mock.Anything,
					database.CreateUserParams{Name: req.Name, Email: req.Email}).
					Return(uuid.Nil, &pgconn.PgError{Code: pgerrcode.UniqueViolation})

				storeMock.EXPECT().ExecTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
					return fn(storeMock)
				})
			},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeAlreadyExists)
				require.Equal(t, connectErr.Message(), "email already in use")
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 1)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "Unexpected error from Store",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{userv1.Role_ROLE_EMPLOYEE},
			},
			buildStubs: func(req *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {
				storeMock.EXPECT().CreateUser(
					mock.Anything,
					database.CreateUserParams{Name: req.Name, Email: req.Email}).
					Return(uuid.Nil, gofakeit.Error())

				storeMock.EXPECT().ExecTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
					return fn(storeMock)
				})
			},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInternal)
				require.Equal(t, connectErr.Message(), "internal server error")
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 1)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "Empty Name",
			req: &userv1.AddUserRequest{
				Name:  "",
				Email: gofakeit.Email(),
				Roles: []userv1.Role{userv1.Role_ROLE_EMPLOYEE},
			},
			buildStubs: func(_ *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "name", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 0)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "Invalid Email",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Word(),
				Roles: []userv1.Role{userv1.Role_ROLE_EMPLOYEE},
			},
			buildStubs: func(_ *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "email", violations.Violations[0].FieldPath)
				require.Equal(t, "valid_email", violations.Violations[0].ConstraintId)
				require.Equal(t, "email must be a valid email", violations.Violations[0].Message)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 0)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "Empty Roles",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{},
			},
			buildStubs: func(_ *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "roles", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 0)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "ROLE_UNSPECIFIED",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{userv1.Role_ROLE_UNSPECIFIED, userv1.Role_ROLE_STAFF},
			},
			buildStubs: func(_ *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "roles_specified", violations.Violations[0].ConstraintId)
				require.Equal(t, "elemests of roles list must be non-zero", violations.Violations[0].Message)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 0)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "Roles Unique",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{userv1.Role_ROLE_STAFF, userv1.Role_ROLE_STAFF},
			},
			buildStubs: func(_ *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "roles", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 0)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
		{
			name: "Roles Empty",
			req: &userv1.AddUserRequest{
				Name:  gofakeit.Name(),
				Email: gofakeit.Email(),
				Roles: []userv1.Role{},
			},
			buildStubs: func(_ *userv1.AddUserRequest, storeMock *dbMock.Store, _ *tokenMock.Maker, _ *smptMock.Mailer) {},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "roles", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 0)
				storeMock.AssertNumberOfCalls(t, "GetRolesByCodes", 0)
				storeMock.AssertNumberOfCalls(t, "AddRolesForUser", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
				mailerMock.AssertNumberOfCalls(t, "Send", 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				Port: 4000,
				Env:  environment.Test,
				Token: &config.Token{
					SimmetricKey: gofakeit.DigitN(32),
				},
			}

			makerMock := new(tokenMock.Maker)
			storeMock := new(dbMock.Store)
			mailerMock := new(smptMock.Mailer)

			s := NewServer(&cfg, makerMock, storeMock, slog.Default(), mailerMock, nil)

			validateInterceptor, err := validate.NewInterceptor()
			require.NoError(t, err)

			srv := setupTestServer(validateInterceptor, &s)
			defer srv.Close()

			client := setupClient(srv)

			tc.buildStubs(tc.req, storeMock, makerMock, mailerMock)

			res, err := client.AddUser(context.Background(), connect.NewRequest(tc.req))

			tc.checkResponse(t, res, err)
			tc.checkMocks(t, storeMock, makerMock, mailerMock)
		})
	}
}

func TestActivateUser(t *testing.T) {
	testCases := []struct {
		name          string
		req           *userv1.ActivateUserRequest
		expectedRes   *userv1.ActivateUserResponse
		buildStubs    func(req *userv1.ActivateUserRequest, expectedRes *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher)
		checkResponse func(t *testing.T, expectedRes *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error)
		checkMocks    func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher)
	}{
		{
			name: "Success",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: &userv1.ActivateUserResponse{
				AuthenticationToken: gofakeit.Word(),
			},
			buildStubs: func(req *userv1.ActivateUserRequest, expectedRes *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				fakePayload := token.Payload{
					UserID: gofakeit.UUID(),
					Scope:  token.ScopeActivation,
				}
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(&fakePayload, nil)

				fakeUser := database.User{
					ID:    uuid.MustParse(fakePayload.UserID),
					Email: gofakeit.Email(),
				}
				storeMock.EXPECT().
					GetUserByID(mock.Anything, uuid.MustParse(fakePayload.UserID)).
					Return(fakeUser, nil)

				fakeHashedPassword := gofakeit.Word()
				hasherMock.EXPECT().
					Hash(req.GetPassword()).
					Return(fakeHashedPassword, nil)

				storeMock.EXPECT().
					ActivateUser(mock.Anything, fakeUser.ID).
					Return(nil)

				storeMock.EXPECT().
					CreateCredentials(mock.Anything, database.CreateCredentialsParams{
						UserID:         fakeUser.ID,
						Email:          fakeUser.Email,
						HashedPassword: fakeHashedPassword,
					}).
					Return(nil)

				storeMock.EXPECT().
					ExecTx(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
						return fn(storeMock)
					})

				makerMock.EXPECT().
					CreateToken(fakeUser.ID.String(), token.ScopeAuthentication).
					Return(expectedRes.GetAuthenticationToken())
			},
			checkResponse: func(t *testing.T, expectedRes *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				require.NoError(t, err)
				require.Equal(t, expectedRes.GetAuthenticationToken(), res.Msg.GetAuthenticationToken())
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 1)
				hasherMock.AssertNumberOfCalls(t, "Hash", 1)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 1)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 1)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 1)
			},
		},
		{
			name: "Invalid Token",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(nil, token.ErrInvalidToken)
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeUnauthenticated)
				require.Equal(t, connectErr.Message(), token.ErrInvalidToken.Error())

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 0)
				hasherMock.AssertNumberOfCalls(t, "Hash", 0)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Invalid Token Scope",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(nil, token.ErrInvalidToken)
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeUnauthenticated)
				require.Equal(t, connectErr.Message(), token.ErrInvalidToken.Error())

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 0)
				hasherMock.AssertNumberOfCalls(t, "Hash", 0)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Expired Token",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(nil, token.ErrExpiredToken)
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeUnauthenticated)
				require.Equal(t, connectErr.Message(), token.ErrExpiredToken.Error())

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 0)
				hasherMock.AssertNumberOfCalls(t, "Hash", 0)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Invalid User ID In Token Payload",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				fakePayload := token.Payload{
					UserID: gofakeit.UUID(),
					Scope:  token.ScopeActivation,
				}
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(&fakePayload, nil)

				storeMock.EXPECT().
					GetUserByID(mock.Anything, uuid.MustParse(fakePayload.UserID)).
					Return(database.User{}, gofakeit.Error())
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeUnauthenticated)
				require.Equal(t, connectErr.Message(), token.ErrInvalidToken.Error())

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 1)
				hasherMock.AssertNumberOfCalls(t, "Hash", 0)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Password Hashing Fails",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				fakePayload := token.Payload{
					UserID: gofakeit.UUID(),
					Scope:  token.ScopeActivation,
				}
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(&fakePayload, nil)

				fakeUser := database.User{
					ID:    uuid.MustParse(fakePayload.UserID),
					Email: gofakeit.Email(),
				}
				storeMock.EXPECT().
					GetUserByID(mock.Anything, uuid.MustParse(fakePayload.UserID)).
					Return(fakeUser, nil)

				hasherMock.EXPECT().
					Hash(req.GetPassword()).
					Return("", gofakeit.Error())
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInternal)
				require.Equal(t, connectErr.Message(), "internal server error")

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 1)
				hasherMock.AssertNumberOfCalls(t, "Hash", 1)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "User Status Update Fails",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				fakePayload := token.Payload{
					UserID: gofakeit.UUID(),
					Scope:  token.ScopeActivation,
				}
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(&fakePayload, nil)

				fakeUser := database.User{
					ID:    uuid.MustParse(fakePayload.UserID),
					Email: gofakeit.Email(),
				}
				storeMock.EXPECT().
					GetUserByID(mock.Anything, uuid.MustParse(fakePayload.UserID)).
					Return(fakeUser, nil)

				fakeHashedPassword := gofakeit.Word()
				hasherMock.EXPECT().
					Hash(req.GetPassword()).
					Return(fakeHashedPassword, nil)

				storeMock.EXPECT().
					ActivateUser(mock.Anything, fakeUser.ID).
					Return(gofakeit.Error())

				storeMock.EXPECT().
					ExecTx(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
						return fn(storeMock)
					})
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInternal)
				require.Equal(t, connectErr.Message(), "internal server error")

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 1)
				hasherMock.AssertNumberOfCalls(t, "Hash", 1)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 1)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Credentials Creation Fails",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(req *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				fakePayload := token.Payload{
					UserID: gofakeit.UUID(),
					Scope:  token.ScopeActivation,
				}
				makerMock.EXPECT().
					VerifyToken(req.GetActivationToken(), token.ScopeActivation).
					Return(&fakePayload, nil)

				fakeUser := database.User{
					ID:    uuid.MustParse(fakePayload.UserID),
					Email: gofakeit.Email(),
				}
				storeMock.EXPECT().
					GetUserByID(mock.Anything, uuid.MustParse(fakePayload.UserID)).
					Return(fakeUser, nil)

				fakeHashedPassword := gofakeit.Word()
				hasherMock.EXPECT().
					Hash(req.GetPassword()).
					Return(fakeHashedPassword, nil)

				storeMock.EXPECT().
					ActivateUser(mock.Anything, fakeUser.ID).
					Return(nil)

				storeMock.EXPECT().
					CreateCredentials(mock.Anything, database.CreateCredentialsParams{
						UserID:         fakeUser.ID,
						Email:          fakeUser.Email,
						HashedPassword: fakeHashedPassword,
					}).
					Return(gofakeit.Error())

				storeMock.EXPECT().
					ExecTx(mock.Anything, mock.Anything).
					RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
						return fn(storeMock)
					})
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInternal)
				require.Equal(t, connectErr.Message(), "internal server error")

				require.Nil(t, res)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 1)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 1)
				hasherMock.AssertNumberOfCalls(t, "Hash", 1)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 1)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 1)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Empty Authentication Token",
			req: &userv1.ActivateUserRequest{
				ActivationToken: "",
				Password:        gofakeit.Password(true, true, true, true, false, 8),
			},
			expectedRes: nil,
			buildStubs: func(_ *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, _ *dbMock.Store, _ *tokenMock.Maker, _ *passwordMock.Hasher) {
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "activation_token", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 0)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 0)
				hasherMock.AssertNumberOfCalls(t, "Hash", 0)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
		{
			name: "Password Too Short",
			req: &userv1.ActivateUserRequest{
				ActivationToken: gofakeit.Word(),
				Password:        gofakeit.Password(true, true, true, true, false, 6),
			},
			expectedRes: nil,
			buildStubs: func(_ *userv1.ActivateUserRequest, _ *userv1.ActivateUserResponse, _ *dbMock.Store, _ *tokenMock.Maker, _ *passwordMock.Hasher) {
			},
			checkResponse: func(t *testing.T, _ *userv1.ActivateUserResponse, res *connect.Response[userv1.ActivateUserResponse], err error) {
				require.Nil(t, res)

				var connectErr *connect.Error
				require.ErrorAs(t, err, &connectErr)
				require.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				require.Len(t, details, 1)

				detail, err := details[0].Value()
				require.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				require.True(t, ok)
				require.Len(t, violations.Violations, 1)
				require.Equal(t, "password", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, hasherMock *passwordMock.Hasher) {
				makerMock.AssertNumberOfCalls(t, "VerifyToken", 0)
				storeMock.AssertNumberOfCalls(t, "GetUserByID", 0)
				hasherMock.AssertNumberOfCalls(t, "Hash", 0)
				storeMock.AssertNumberOfCalls(t, "ExecTx", 0)
				storeMock.AssertNumberOfCalls(t, "ActivateUser", 0)
				storeMock.AssertNumberOfCalls(t, "CreateCredentials", 0)
				makerMock.AssertNumberOfCalls(t, "CreateToken", 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				Port: 4000,
				Env:  environment.Test,
				Token: &config.Token{
					SimmetricKey: gofakeit.DigitN(32),
				},
			}

			makerMock := new(tokenMock.Maker)
			storeMock := new(dbMock.Store)
			hasherMock := new(passwordMock.Hasher)

			s := NewServer(&cfg, makerMock, storeMock, slog.Default(), nil, hasherMock)

			validateInterceptor, err := validate.NewInterceptor()
			require.NoError(t, err)

			srv := setupTestServer(validateInterceptor, &s)
			defer srv.Close()

			client := setupClient(srv)
			tc.buildStubs(tc.req, tc.expectedRes, storeMock, makerMock, hasherMock)

			res, err := client.ActivateUser(context.Background(), connect.NewRequest(tc.req))

			tc.checkResponse(t, tc.expectedRes, res, err)
			tc.checkMocks(t, storeMock, makerMock, hasherMock)
		})
	}
}
