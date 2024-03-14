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

func createRandomUser(t *testing.T) uuid.UUID {
	arg := CreateUserParams{
		Name:  gofakeit.Name(),
		Email: gofakeit.Email(),
	}

	userID, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotZero(t, userID.String())

	return userID
}

func TestCreateUser(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		createRandomUser(t)
	})

	t.Run("Emaill already exist", func(t *testing.T) {
		arg := CreateUserParams{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		}

		userID, err := testStore.CreateUser(context.Background(), arg)
		require.NotEmpty(t, userID.String())
		require.NoError(t, err)

		userID, err = testStore.CreateUser(context.Background(), arg)
		require.Zero(t, userID)

		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr)
		require.Equal(t, pgerrcode.UniqueViolation, pgErr.Code)
	})
}
