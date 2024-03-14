//go:build database
// +build database

package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
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
		name  string
		codes []string
	}{
		{
			name:  "Single code",
			codes: []string{"employee"},
		},
		{
			name:  "Multiple codes",
			codes: []string{"employee", "leader"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userID := createRandomUser(t)

			roleIDS, err := testStore.GetRolesByCodes(context.Background(), tc.codes)
			require.NoError(t, err)
			require.Len(t, roleIDS, len(tc.codes))

			args := make([]AddRolesForUserParams, len(roleIDS))
			for i := range roleIDS {
				args[i] = AddRolesForUserParams{
					UserID:  userID,
					Grantor: userID,
					RoleID:  roleIDS[i],
				}
			}

			rows, err := testStore.AddRolesForUser(context.Background(), args)
			require.NoError(t, err)
			require.Equal(t, rows, int64(len(roleIDS)))
		})
	}
}
