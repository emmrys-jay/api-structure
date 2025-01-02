CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE "users_role_enum" AS ENUM ('admin', 'user');

CREATE TABLE "users" (
    "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "first_name" varchar NOT NULL,
    "last_name" varchar NOT NULL,
    "email" varchar NOT NULL,
    "password" varchar NOT NULL,
    "role" users_role_enum DEFAULT 'user',
    "is_active" bool NOT NULL DEFAULT false,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "updated_at" timestamptz NOT NULL DEFAULT (now()),
    "deleted_at" timestamptz
);

CREATE UNIQUE INDEX "email" ON "users" ("email");