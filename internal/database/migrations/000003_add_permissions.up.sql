CREATE TABLE "permissions" (
    "id" uuid PRIMARY KEY DEFAULT (uuid_generate_v4()),
    "code" varchar UNIQUE NOT NULL,
    "description" varchar NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT (now())
);

CREATE TABLE "permissions_roles" (
    "role_id" uuid NOT NULL,
    "permission_id" uuid NOT NULL,
    "deleted_at" timestamp DEFAULT null,
    "created_at" timestamp NOT NULL DEFAULT (now()),
    PRIMARY KEY ("role_id", "permission_id")
);

ALTER TABLE "permissions_roles" ADD CONSTRAINT "permission_role" FOREIGN KEY (
    "permission_id"
) REFERENCES "permissions" ("id");

ALTER TABLE "permissions_roles" ADD CONSTRAINT "role_permission" FOREIGN KEY (
    "role_id"
) REFERENCES "roles" ("id");
