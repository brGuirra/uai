// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: copyfrom.go

package database

import (
	"context"
)

// iteratorForCreateRoles implements pgx.CopyFromSource.
type iteratorForCreateRoles struct {
	rows                 []string
	skippedFirstNextCall bool
}

func (r *iteratorForCreateRoles) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForCreateRoles) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0],
	}, nil
}

func (r iteratorForCreateRoles) Err() error {
	return nil
}

func (q *Queries) CreateRoles(ctx context.Context, code []string) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"roles"}, []string{"code"}, &iteratorForCreateRoles{rows: code})
}