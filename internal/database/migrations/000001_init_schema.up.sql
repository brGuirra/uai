CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- noqa: L057

CREATE TYPE "user_status" AS ENUM (
    'created',
    'active',
    'former_employee'
);

CREATE TYPE "ticket_status" AS ENUM (
    'pending',
    'approved',
    'rejected',
    'closed'
);

CREATE TYPE "ticket_reason" AS ENUM (
    'forgot',
    'system_down'
);

CREATE TYPE "token_scope" AS ENUM (
    'activation',
    'authentication'
);

CREATE TABLE "users" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "name" varchar NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "status" user_status NOT NULL DEFAULT 'created',
    "version" int NOT NULL DEFAULT 1,
    "updated_at" timestamp NOT NULL DEFAULT (now()),
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "credentials" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "user_id" uuid NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "hashed_password" varchar NOT NULL,
    "login_attempts" smallint NOT NULL DEFAULT 0,
    "version" int NOT NULL DEFAULT 1,
    "updated_at" timestamp NOT NULL DEFAULT (now()),
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "profiles" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "user_id" uuid NOT NULL,
    "nickname" varchar NOT NULL,
    "picture" varchar DEFAULT null,
    "bio" varchar DEFAULT null,
    "version" int NOT NULL DEFAULT 1,
    "updated_at" timestamp NOT NULL DEFAULT (now()),
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "tokens" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "hash" bytea NOT NULL,
    "user_id" uuid NOT NULL,
    "expiry" timestamp NOT NULL,
    "scope" token_scope
);

CREATE TABLE "attendance_records" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "employee" uuid NOT NULL,
    "puch_time" timestamp NOT NULL DEFAULT (now()),
    "ticket_id" uuid DEFAULT null,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "tickets" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "requester" uuid NOT NULL,
    "attendant" uuid DEFAULT null,
    "status" ticket_status NOT NULL,
    "punch_time" timestamp NOT NULL,
    "reason" ticket_reason NOT NULL,
    "updated_at" timestamp NOT NULL DEFAULT (now()),
    "created_at" timestamp NOT NULL DEFAULT (now())
);

ALTER TABLE "credentials" ADD CONSTRAINT "user_credential" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");

ALTER TABLE "profiles" ADD CONSTRAINT "user_profile" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");

ALTER TABLE "tokens" ADD CONSTRAINT "token_user" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");

ALTER TABLE "attendance_records" ADD CONSTRAINT
"attendance_record_employee" FOREIGN KEY (
    "employee"
) REFERENCES "users" ("id");

ALTER TABLE "tickets" ADD CONSTRAINT "ticket_requester" FOREIGN KEY (
    "requester"
) REFERENCES "users" ("id");

ALTER TABLE "tickets" ADD CONSTRAINT "ticket_attendant" FOREIGN KEY (
    "attendant"
) REFERENCES "users" ("id");
