CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- noqa: L057

CREATE TABLE IF NOT EXISTS "employees" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "name" varchar NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "hashed_password" varchar DEFAULT NULL,
    "status" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "roles" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "code" varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS "employees_roles" (
    "employee_id" uuid NOT NULL,
    "role_id" uuid NOT NULL,
    "grantor" uuid NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    PRIMARY KEY ("employee_id", "role_id")
);

CREATE TABLE IF NOT EXISTS "attendance_records" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "employee" uuid NOT NULL,
    "created_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "tickets" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "requester" uuid NOT NULL,
    "attendant" uuid DEFAULT NULL,
    "status" varchar NOT NULL,
    "updated_at" timestamp NOT NULL DEFAULT (now()),
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE IF NOT EXISTS "absences" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "employee" uuid NOT NULL,
    "reason" varchar NOT NULL,
    "from_date" date NOT NULL,
    "to_date" date NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

ALTER TABLE "employees_roles"
ADD CONSTRAINT "user_role" FOREIGN KEY ("employee_id") REFERENCES "employees" (
    "id"
);

ALTER TABLE "employees_roles"
ADD CONSTRAINT "role_user" FOREIGN KEY ("role_id") REFERENCES "roles" ("id");

ALTER TABLE "employees_roles"
ADD CONSTRAINT "role_grantor" FOREIGN KEY ("grantor") REFERENCES "employees" (
    "id"
);

ALTER TABLE "attendance_records"
ADD CONSTRAINT "attendance_record_employee" FOREIGN KEY (
    "employee"
) REFERENCES "employees" ("id");

ALTER TABLE "tickets"
ADD CONSTRAINT "ticket_requester" FOREIGN KEY (
    "requester"
) REFERENCES "employees" ("id");

ALTER TABLE "tickets"
ADD CONSTRAINT "ticket_attendant" FOREIGN KEY (
    "attendant"
) REFERENCES "employees" ("id");

ALTER TABLE "absences"
ADD CONSTRAINT "absence_employee" FOREIGN KEY (
    "employee"
) REFERENCES "employees" ("id");
