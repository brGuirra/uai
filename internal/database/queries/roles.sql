-- name: CreateRoles :copyfrom
INSERT INTO roles (code)
VALUES ($1);

-- name: GetRoles :many
SELECT
    "id",
    "code",
    "description"
FROM roles;
