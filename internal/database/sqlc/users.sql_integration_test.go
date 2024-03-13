//go:build database
// +build database

package database

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		arg := CreateUserParams{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		}

		userID, err := testStore.CreateUser(context.Background(), arg)
		require.NoError(t, err)
		require.NotEmpty(t, userID.String())
	})

	t.Run("Emaill already exist", func(t *testing.T) {
		arg := CreateUserParams{
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
		}

		_, err := testStore.CreateUser(context.Background(), arg)
		require.NoError(t, err)

		_, err = testStore.CreateUser(context.Background(), arg)

		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr)
		require.Equal(t, pgerrcode.UniqueViolation, pgErr.Code)
	})
}
