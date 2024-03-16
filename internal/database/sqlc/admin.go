package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateInitialAdminUser checks if the intial admin user is already created
// and if not creates it.
//
// This function should be called on server startup.
func CreateInitialAdminUser(store Store, name, email, hashedPassword string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	return store.ExecTx(ctx, func(q Querier) error {
		_, err := q.GetUserByEmail(ctx, email)
		if err != nil {
			if err == pgx.ErrNoRows {
				userID, err := q.CreateAdminUser(ctx, CreateAdminUserParams{
					Name:   name,
					Email:  email,
					Status: UserStatusActive,
				})
				if err != nil {
					return err
				}

				roleIDS, err := q.GetRolesByCodes(ctx, []string{"admin"})
				if err != nil {
					return err
				}

				args := make([]AddRolesForUserParams, len(roleIDS))
				for i := range roleIDS {
					args[i] = AddRolesForUserParams{
						UserID:  userID,
						Grantor: userID,
						RoleID:  roleIDS[i],
					}
				}

				_, err = q.AddRolesForUser(ctx, args)
				if err != nil {
					return err
				}

				err = q.CreateCredentials(ctx, CreateCredentialsParams{
					UserID:         userID,
					Email:          email,
					HashedPassword: hashedPassword,
				})
				if err != nil {
					return err
				}

				return nil
			}

			return err
		}

		return nil
	})
}
