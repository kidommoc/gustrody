-- DROP DATABASE IF EXISTS austrody;
-- 
-- CREATE DATABASE austrody
--     WITH
--     OWNER = penguin
--     ENCODING = 'UTF8'
--     LC_COLLATE = 'en_US.utf8'
--     LC_CTYPE = 'en_US.utf8'
--     LOCALE_PROVIDER = 'libc'
--     TABLESPACE = pg_default
--     CONNECTION LIMIT = -1
--     IS_TEMPLATE = False;

CREATE TYPE vsb AS ENUM (
  'public', 'follower', 'direct'
);

CREATE TYPE kp AS (
  "pub" text,
  "pri" text
);

CREATE TYPE img AS (
  "url" text,
  "alt" text
);

CREATE TABLE IF NOT EXISTS users (
  "username" varchar(20) PRIMARY KEY,
  "nickname" text NOT NULL,
  "summary" text,
  "createdAt" timestamp NOT NULL,
  "avatar" text,
  "keys" kp, -- NOT NULL
  "preferences" jsonb DEFAULT '{"postVsb":"public","shareVsb":"public"}'
);

CREATE INDEX user_pf_postVsb ON users USING gin(("preferences"->'postVsb'));
CREATE INDEX user_pf_shareVsb ON users USING gin(("preferences"->'shareVsb'));

CREATE TABLE IF NOT EXISTS follow (
  "from" text,
  "to" varchar(60) CHECK ("to" <> "from"),
  PRIMARY KEY ("from", "to")
);

CREATE INDEX user_follows ON follow ("from");
CREATE INDEX user_be_followed ON follow ("to");

CREATE VIEW follow_info ("user", "followings", "followers") AS
  WITH followings AS (
    SELECT "from" AS u, COUNT(*) AS c
    FROM follow
    GROUP BY u
  ), followers AS (
    SELECT "to" AS u, COUNT(*) AS c
    FROM follow
    GROUP BY u
  )
  SELECT
    users."username" AS "user",
    COALESCE(followings.c, 0) AS "followings",
    COALESCE(followers.c, 0) AS "followers"
  FROM users
  FULL JOIN followings
    ON users."username" = followings.u
  FULL JOIN followers
    ON users."username" = followers.u;

CREATE TABLE IF NOT EXISTS posts (
  "id" varchar(36) PRIMARY KEY,
  "url" text NOT NULL,
  "date" timestamp NOT NULL,
  "user" varchar(60) NOT NULL,
  "replying" text,
  "vsb" vsb NOT NULL,
  "content" text NOT NULL,
  "media" img[] DEFAULT array[]::img[],
  "likes" text[] DEFAULT array[]::text[],
  "shares" text[] DEFAULT array[]::text[]
);

CREATE INDEX posters ON posts ("user");

CREATE TABLE IF NOT EXISTS shares (
  "id" varchar(36) NOT NULL,
  "user" varchar(60) NOT NULL,
  "date" timestamp NOT NULL,
  "vsb" vsb NOT NULL,
  PRIMARY KEY ("id", "user"),
  FOREIGN KEY ("id") REFERENCES posts("id")
);

CREATE INDEX sharers ON shares ("user");