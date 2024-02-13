-- name: CreateEmployee :one
INSERT INTO
employees (name, email, status, hashed_password)
VALUES
($1, $2, $3, $4)
RETURNING id, name, email, status;

-- name: UpdateEmployee :exec
UPDATE employees
SET
    name = $2,
    email = $3,
    hashed_password = $4,
    status = $5
WHERE
    id = $1
RETURNING id, name, email, status;

-- name: GetEmployeeByID :one
SELECT *
FROM employees
WHERE id = $1;

-- name: GetEmployeeByEmail :one
SELECT * FROM employees
WHERE email = $1;

-- name: CheckEmployeeEmailExists :one
SELECT EXISTS(SELECT 1 FROM employees WHERE email = $1) AS user_exists;

-- name: AddRolesForEmployee :copyfrom
INSERT INTO employees_roles (employee_id, role_id, grantor)
VALUES
($1, $2, $3);
