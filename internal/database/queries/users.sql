-- name: CreateUser :one
INSERT INTO
"users" ("name", "email", "status", "hashed_password")
VALUES
($1, $2, $3, $4)
RETURNING "id", "name", "email", "status";

-- name: UpdateUser :exec
UPDATE "users"
SET
    "name" = $2,
    "email" = $3,
    "hashed_password" = $4,
    "status" = $5
WHERE
    "id" = $1
RETURNING "id", "name", "email", "status";

-- name: GetUserByID :one
SELECT
    "name",
    "email",
    "status",
    "hashed_password"
FROM "users"
WHERE "id" = $1;

-- name: GetUserByEmail :one
SELECT
    "name",
    "email",
    "status",
    "hashed_password"
FROM "users"
WHERE "email" = $1;
