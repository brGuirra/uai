CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- noqa: L057

CREATE TABLE IF NOT EXISTS "users" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "name" varchar NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "hashed_password" varchar DEFAULT NULL,
    "status" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "roles" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "display_name" varchar NOT NULL,
    "description" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "users_roles" (
    "user_id" uuid NOT NULL,
    "role_id" uuid NOT NULL,
    "grantor" uuid NOT NULL,
    "granted_at" timestamp NOT NULL DEFAULT (now()),
    PRIMARY KEY ("user_id", "role_id")
);

CREATE TABLE IF NOT EXISTS "permissions" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "display_name" varchar NOT NULL,
    "description" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "roles_permissions" (
    "role_id" uuid NOT NULL,
    "permission_id" uuid NOT NULL,
    PRIMARY KEY ("role_id", "permission_id")
);

ALTER TABLE "users_roles" ADD CONSTRAINT "user_role" FOREIGN KEY (
    "user_id"
) REFERENCES "users" (
    "id"
);

ALTER TABLE "users_roles" ADD CONSTRAINT "role_user" FOREIGN KEY (
    "role_id"
) REFERENCES "roles" ("id");

ALTER TABLE "users_roles" ADD CONSTRAINT "role_grantor" FOREIGN KEY (
    "grantor"
) REFERENCES "users" ("id");

ALTER TABLE "roles_permissions" ADD CONSTRAINT "role_permission" FOREIGN KEY (
    "role_id"
) REFERENCES "roles" ("id");

ALTER TABLE "roles_permissions" ADD CONSTRAINT "permission_role" FOREIGN KEY (
    "permission_id"
) REFERENCES "permissions" ("id");
