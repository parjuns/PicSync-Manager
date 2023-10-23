CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "mobile" numeric NOT NULL,
  "latitude" decimal NOT NULL,
  "longitude" decimal NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz
);

CREATE TABLE "products" (
  "id" bigserial PRIMARY KEY,
  "name" varchar NOT NULL,
  "description" text NOT NULL,
  "images" text [] NOT NULL,
  "price" decimal NOT NULL,
  "user_id" bigserial NOT NULL,
  "compressed_images" text [],
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz 
);

ALTER TABLE "products" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

INSERT INTO "users" (id, name, mobile,latitude,longitude,created_at) VALUES (
  generate_series(1, 100),
  upper(substr(md5(random()::text), 0, 8)),
  CAST(1000000000 + floor(random() * 9000000000) AS bigint),
  random() * 100,
  random() * 100,
  '2000-01-01'::date + trunc(random() * 366 * 10)::int
  );