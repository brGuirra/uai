package main

import (
	"context"
	"net/http"
	"time"

	database "github.com/brGuirra/uai/internal/database/sqlc"
	"github.com/brGuirra/uai/internal/password"
	"github.com/brGuirra/uai/internal/request"
	"github.com/brGuirra/uai/internal/response"
	"github.com/brGuirra/uai/internal/validator"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pascaldekloe/jwt"
)

func (app *application) createEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string              `json:"name"`
		Email     string              `json:"email"`
		Password  string              `json:"password"`
		Roles     []string            `json:"roles"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := app.store.CheckEmployeeEmailExists(ctx, input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(validator.Matches(input.Email, validator.RgxEmail), "Email", "Must be a valid email address")
	input.Validator.CheckField(!exists, "Email", "Email is already in use")

	input.Validator.CheckField(input.Password != "", "Password", "Password is required")
	input.Validator.CheckField(len(input.Password) >= 8, "Password", "Password is too short")
	input.Validator.CheckField(len(input.Password) <= 72, "Password", "Password is too long")
	input.Validator.CheckField(validator.NotIn(input.Password, password.CommonPasswords...), "Password", "Password is too common")
	input.Validator.CheckField(validator.AllIn(input.Roles, "staff", "leader", "employee"), "Roles", "Invalid role, must be 'staff', 'leader' or 'employee'")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	_, err = app.store.CreateEmployee(ctx, database.CreateEmployeeParams{
		Name:           input.Name,
		Email:          input.Email,
		Status:         "unverified",
		HashedPassword: pgtype.Text{},
	})
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email     string              `json:"Email"`
		Password  string              `json:"Password"`
		Validator validator.Validator `json:"-"`
	}

	err := request.DecodeJSON(w, r, &input)
	if err != nil {
		app.badRequest(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	employee, err := app.store.GetEmployeeByEmail(ctx, input.Email)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Email != "", "Email", "Email is required")
	input.Validator.CheckField(employee.Email != "", "Email", "Email address could not be found")

	passwordMatches, err := password.Matches(input.Password, employee.HashedPassword.String)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	input.Validator.CheckField(input.Password != "", "Password", "Password is required")
	input.Validator.CheckField(passwordMatches, "Password", "Password is incorrect")

	if input.Validator.HasErrors() {
		app.failedValidation(w, r, input.Validator)
		return
	}

	var claims jwt.Claims

	claims.Subject = employee.ID.String()

	expiry := time.Now().Add(24 * time.Hour)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = app.config.baseURL
	claims.Audiences = []string{app.config.baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secretKey))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := map[string]string{
		"authenticationToken":       string(jwtBytes),
		"authenticationTokenExpiry": expiry.Format(time.RFC3339),
	}

	err = response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

func (app *application) protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a protected handler"))
}
