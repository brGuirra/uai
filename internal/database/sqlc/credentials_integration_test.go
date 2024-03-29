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

func TestCreateCredentials(t *testing.T) {
	testCases := []struct {
		name        string
		userID      uuid.UUID
		checkResult func(t *testing.T, err error)
	}{
		{
			name:   "Success",
			userID: createRandomUser(t),
			checkResult: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name:   "User Does Not Exist",
			userID: uuid.MustParse(gofakeit.UUID()),
			checkResult: func(t *testing.T, err error) {
				var pgErr *pgconn.PgError
				require.ErrorAs(t, err, &pgErr)
				require.Equal(t, pgerrcode.ForeignKeyViolation, pgErr.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := testStore.CreateCredentials(context.Background(), CreateCredentialsParams{
				UserID:         tc.userID,
				Email:          gofakeit.Email(),
				HashedPassword: gofakeit.Password(true, true, true, true, false, 8),
			})
			tc.checkResult(t, err)
		})
	}
}
