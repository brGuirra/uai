-- name: GetRolesByCodes :many
SELECT "id" FROM "roles" WHERE "code" = ANY(sqlc.arg(codes)::text[]);

-- name: AddRolesForUser :copyfrom
INSERT INTO "users_roles" ("user_id", "role_id", "grantor") VALUES ($1, $2, $3);
