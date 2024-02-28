ALTER TABLE "permissions_roles" DROP CONSTRAINT "permission_role";

ALTER TABLE "permissions_roles" DROP CONSTRAINT "role_permission";

ALTER TABLE "users_roles" DROP CONSTRAINT "role_user";

TRUNCATE TABLE "permissions_roles";

TRUNCATE TABLE "roles";

TRUNCATE TABLE "permissions";

ALTER TABLE "users_roles" ADD CONSTRAINT "role_user" FOREIGN KEY (
    "role_id"
) REFERENCES "roles" ("id");
