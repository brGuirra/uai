//go:build database
// +build database

package database

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestGetRolesByCodes(t *testing.T) {
	testCases := []struct {
		checkResult func(t *testing.T, roleIDS []uuid.UUID, err error)
		name        string
		codes       []string
	}{
		{
			name:  "Multiple codes",
			codes: []string{"employee", "staff", "leader", "admin"},
			checkResult: func(t *testing.T, roleIDS []uuid.UUID, err error) {
				require.NoError(t, err)
				require.Len(t, roleIDS, 4)
			},
		},
		{
			name:  "Single codes",
			codes: []string{"employee"},
			checkResult: func(t *testing.T, roleIDS []uuid.UUID, err error) {
				require.NoError(t, err)
				require.Len(t, roleIDS, 1)
			},
		},
		{
			name:  "Non existing code",
			codes: []string{"any_code"},
			checkResult: func(t *testing.T, roleIDS []uuid.UUID, err error) {
				require.NoError(t, err)
				require.Len(t, roleIDS, 0)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleIDS, err := testStore.GetRolesByCodes(context.Background(), tc.codes)
			tc.checkResult(t, roleIDS, err)
		})
	}
}

func TestAddRolesForUser(t *testing.T) {
	testCases := []struct {
		checkResult func(t *testing.T, rows int64, err error)
		name        string
		codes       []string
		userID      uuid.UUID
	}{
		{
			name:   "Single code",
			userID: createRandomUser(t),
			codes:  []string{"employee"},
			checkResult: func(t *testing.T, rows int64, err error) {
				require.NoError(t, err)
				require.Equal(t, rows, int64(1))
			},
		},
		{
			name:   "Multiple codes",
			userID: createRandomUser(t),
			codes:  []string{"employee", "leader"},
			checkResult: func(t *testing.T, rows int64, err error) {
				require.NoError(t, err)
				require.Equal(t, rows, int64(2))
			},
		},
		{
			name:   "User Does Not Exist",
			userID: uuid.MustParse(gofakeit.UUID()),
			codes:  []string{"employee"},
			checkResult: func(t *testing.T, rows int64, err error) {
				var pgErr *pgconn.PgError
				require.ErrorAs(t, err, &pgErr)
				require.Equal(t, pgerrcode.ForeignKeyViolation, pgErr.Code)

				require.Equal(t, rows, int64(0))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleIDS, err := testStore.GetRolesByCodes(context.Background(), tc.codes)
			require.NoError(t, err)
			require.Len(t, roleIDS, len(tc.codes))

			args := make([]AddRolesForUserParams, len(roleIDS))
			for i := range roleIDS {
				args[i] = AddRolesForUserParams{
					UserID:  tc.userID,
					Grantor: tc.userID,
					RoleID:  roleIDS[i],
				}
			}

			rows, err := testStore.AddRolesForUser(context.Background(), args)
			tc.checkResult(t, rows, err)
		})
	}
}
