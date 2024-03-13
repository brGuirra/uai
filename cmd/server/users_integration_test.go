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
	smptMock "github.com/brGuirra/uai/mocks/smtp"
	tokenMock "github.com/brGuirra/uai/mocks/token"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func setupClients(srv *httptest.Server) []userv1connect.UserServiceClient {
	connectClient := userv1connect.NewUserServiceClient(srv.Client(), srv.URL)
	gRPCClient := userv1connect.NewUserServiceClient(srv.Client(), srv.URL, connect.WithGRPC())
	webRPCClient := userv1connect.NewUserServiceClient(srv.Client(), srv.URL, connect.WithGRPCWeb())

	return []userv1connect.UserServiceClient{connectClient, gRPCClient, webRPCClient}
}

func TestAddUser(t *testing.T) {
	testCases := []struct {
		req           *userv1.AddUserRequest
		buildStubs    func(req *userv1.AddUserRequest, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer)
		checkResponse func(t *testing.T, res *connect.Response[emptypb.Empty], err error)
		checkMocks    func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer)
		name          string
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
					Return(fakeUserID, nil).Once()

				fakeRoleIDS := []uuid.UUID{uuid.MustParse(gofakeit.UUID())}
				storeMock.EXPECT().GetRolesByCodes(mock.Anything, []string{"employee"}).Return(fakeRoleIDS, nil).Once()

				storeMock.EXPECT().AddRolesForUser(mock.Anything, []database.AddRolesForUserParams{
					{
						UserID:  fakeUserID,
						Grantor: fakeUserID,
						RoleID:  fakeRoleIDS[0],
					},
				}).Return(1, nil).Once()

				storeMock.EXPECT().ExecTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
					return fn(storeMock)
				}).Once()

				fakeToken := gofakeit.Word()
				makerMock.EXPECT().CreateToken(fakeUserID.String(), token.ScopeActivation).Return(fakeToken).Once()

				data := map[string]any{
					"activationToken": fakeToken,
					"userID":          fakeUserID,
				}
				mailerMock.EXPECT().Send(req.Email, data, "welcome.tmpl").Return(nil).Once()
			},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				assert.NoError(t, err)
				assert.IsType(t, &emptypb.Empty{}, res.Msg)
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
					Return(uuid.Nil, &pgconn.PgError{Code: pgerrcode.UniqueViolation}).Once()

				storeMock.EXPECT().ExecTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
					return fn(storeMock)
				}).Once()
			},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeAlreadyExists)
				assert.Equal(t, connectErr.Message(), "email already in use")
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 1)
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
					Return(uuid.Nil, gofakeit.Error()).Once()

				storeMock.EXPECT().ExecTx(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, fn func(database.Querier) error) error {
					return fn(storeMock)
				}).Once()
			},
			checkResponse: func(t *testing.T, res *connect.Response[emptypb.Empty], err error) {
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInternal)
				assert.Equal(t, connectErr.Message(), "internal server error")
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNumberOfCalls(t, "ExecTx", 1)
				storeMock.AssertNumberOfCalls(t, "CreateUser", 1)
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				assert.Len(t, details, 1)

				detail, err := details[0].Value()
				assert.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				assert.True(t, ok)
				assert.Len(t, violations.Violations, 1)
				assert.Equal(t, "name", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNotCalled(t, "ExecTx")
				storeMock.AssertNotCalled(t, "CreateUser")
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				assert.Len(t, details, 1)

				detail, err := details[0].Value()
				assert.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				assert.True(t, ok)
				assert.Len(t, violations.Violations, 1)
				assert.Equal(t, "email", violations.Violations[0].FieldPath)
				assert.Equal(t, "valid_email", violations.Violations[0].ConstraintId)
				assert.Equal(t, "email must be a valid email", violations.Violations[0].Message)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNotCalled(t, "ExecTx")
				storeMock.AssertNotCalled(t, "CreateUser")
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				assert.Len(t, details, 1)

				detail, err := details[0].Value()
				assert.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				assert.True(t, ok)
				assert.Len(t, violations.Violations, 1)
				assert.Equal(t, "roles", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNotCalled(t, "ExecTx")
				storeMock.AssertNotCalled(t, "CreateUser")
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				assert.Len(t, details, 1)

				detail, err := details[0].Value()
				assert.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				assert.True(t, ok)
				assert.Len(t, violations.Violations, 1)
				assert.Equal(t, "roles_specified", violations.Violations[0].ConstraintId)
				assert.Equal(t, "elemests of roles list must be non-zero", violations.Violations[0].Message)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNotCalled(t, "ExecTx")
				storeMock.AssertNotCalled(t, "CreateUser")
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				assert.Len(t, details, 1)

				detail, err := details[0].Value()
				assert.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				assert.True(t, ok)
				assert.Len(t, violations.Violations, 1)
				assert.Equal(t, "roles", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNotCalled(t, "ExecTx")
				storeMock.AssertNotCalled(t, "CreateUser")
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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
				assert.Nil(t, res)

				var connectErr *connect.Error
				assert.ErrorAs(t, err, &connectErr)
				assert.Equal(t, connectErr.Code(), connect.CodeInvalidArgument)

				details := connectErr.Details()
				assert.Len(t, details, 1)

				detail, err := details[0].Value()
				assert.NoError(t, err)

				violations, ok := detail.(*validatepb.Violations)
				assert.True(t, ok)
				assert.Len(t, violations.Violations, 1)
				assert.Equal(t, "roles", violations.Violations[0].FieldPath)
			},
			checkMocks: func(t *testing.T, storeMock *dbMock.Store, makerMock *tokenMock.Maker, mailerMock *smptMock.Mailer) {
				storeMock.AssertNotCalled(t, "ExecTx")
				storeMock.AssertNotCalled(t, "CreateUser")
				storeMock.AssertNotCalled(t, "GetRolesByCodes")
				storeMock.AssertNotCalled(t, "AddRolesForUser")
				makerMock.AssertNotCalled(t, "CreateToken")
				mailerMock.AssertNotCalled(t, "Send")
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

			s := NewServer(&cfg, makerMock, storeMock, slog.Default(), mailerMock)

			validateInterceptor, err := validate.NewInterceptor()
			assert.NoError(t, err)

			srv := setupTestServer(validateInterceptor, &s)
			defer srv.Close()

			clients := setupClients(srv)

			for _, client := range clients {
				tc.buildStubs(tc.req, storeMock, makerMock, mailerMock)

				res, err := client.AddUser(context.Background(), connect.NewRequest(tc.req))

				tc.checkResponse(t, res, err)
			}
		})
	}
}
