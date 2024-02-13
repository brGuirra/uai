-- name: CreateRoles :copyfrom
INSERT INTO roles (code)
VALUES ($1);

-- name: GetRoles :many
SELECT *
FROM roles
ORDER BY id;
