//go:build database
// +build database

package database

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
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
	t.Run("Success", func(t *testing.T) {
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

func TestGetUserByEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		newUser := CreateUserParams{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		}

		userID, err := testStore.CreateUser(context.Background(), newUser)
		require.NoError(t, err)

		user, err := testStore.GetUserByEmail(context.Background(), newUser.Email)
		require.NoError(t, err)

		require.Equal(t, user.ID, userID)
		require.Equal(t, user.Email, newUser.Email)
		require.Equal(t, user.Name, newUser.Name)
		require.Equal(t, user.Status, UserStatusCreated)
		require.Equal(t, user.Version, int32(1))
		require.NotZero(t, user.CreatedAt)
		require.NotZero(t, user.UpdatedAt)
	})

	t.Run("User Not Found", func(t *testing.T) {
		user, err := testStore.GetUserByEmail(context.Background(), gofakeit.Email())
		require.ErrorIs(t, err, pgx.ErrNoRows)

		require.Zero(t, user)
	})
}
