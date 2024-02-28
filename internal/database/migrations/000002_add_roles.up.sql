CREATE TABLE "roles" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "code" varchar UNIQUE NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "users_roles" (
    "user_id" uuid NOT NULL,
    "role_id" uuid NOT NULL,
    "grantor" uuid NOT NULL,
    "deleted_at" timestamp DEFAULT null,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    PRIMARY KEY ("user_id", "role_id")
);

ALTER TABLE "users_roles" ADD CONSTRAINT "user_role" FOREIGN KEY (
    "user_id"
) REFERENCES "users" ("id");

ALTER TABLE "users_roles" ADD CONSTRAINT "role_user" FOREIGN KEY (
    "role_id"
) REFERENCES "roles" ("id");

ALTER TABLE "users_roles" ADD CONSTRAINT "role_grantor" FOREIGN KEY (
    "grantor"
) REFERENCES "users" ("id");
