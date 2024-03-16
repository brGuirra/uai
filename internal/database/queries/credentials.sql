-- name: CreateCredentials :exec
INSERT INTO "credentials"
  ("user_id", "email", "hashed_password")
VALUES
  ($1, $2, $3);
