CREATE TYPE "users_role_enum" AS ENUM ('admin', 'user');

CREATE TABLE "users" (
    "id" BIGSERIAL PRIMARY KEY,
    "name" varchar NOT NULL,
    "email" varchar NOT NULL,
    "password" varchar NOT NULL,
    "role" users_role_enum DEFAULT 'user',
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE UNIQUE INDEX "email" ON "users" ("email");