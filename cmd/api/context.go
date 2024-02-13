package main

import (
	"context"
	"net/http"

	database "github.com/brGuirra/uai/internal/database/sqlc"
)

type contextKey string

const (
	authenticatedUserContextKey = contextKey("authenticatedUser")
)

func contextSetAuthenticatedUser(r *http.Request, employee *database.Employee) *http.Request {
	ctx := context.WithValue(r.Context(), authenticatedUserContextKey, employee)
	return r.WithContext(ctx)
}

func contextGetAuthenticatedUser(r *http.Request) *database.Employee {
	employee, ok := r.Context().Value(authenticatedUserContextKey).(*database.Employee)
	if !ok {
		return nil
	}

	return employee
}
