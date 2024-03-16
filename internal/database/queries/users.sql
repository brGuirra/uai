-- name: CreateUser :one
INSERT INTO
"users" ("name", "email")
VALUES ($1, $2)
RETURNING "id";

-- name: CreateAdminUser :one
INSERT INTO
"users" ("name", "email", "status")
VALUES ($1, $2, $3)
RETURNING "id";

-- name: GetUserByEmail :one
SELECT * FROM "users" WHERE email = $1;
