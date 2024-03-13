// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: roles.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

type AddRolesForUserParams struct {
	UserID  uuid.UUID `json:"user_id"`
	RoleID  uuid.UUID `json:"role_id"`
	Grantor uuid.UUID `json:"grantor"`
}

const getRolesByCodes = `-- name: GetRolesByCodes :many
SELECT "id" FROM "roles" WHERE "code" = ANY($1::text[])
`

func (q *Queries) GetRolesByCodes(ctx context.Context, codes []string) ([]uuid.UUID, error) {
	rows, err := q.db.Query(ctx, getRolesByCodes, codes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
