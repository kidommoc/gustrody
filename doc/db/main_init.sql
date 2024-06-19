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

CREATE TABLE IF NOT EXISTS users (
    "username" varchar(20) PRIMARY KEY,
    "nickname" text,
    "summary" text,
    "createdAt" timestamp NOT NULL,
    "icon" text
);

CREATE TABLE IF NOT EXISTS follow (
    "from" text,
    "to" text,
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
    "id" text PRIMARY KEY,
    "user" text NOT NULL,
    "date" timestamp NOT NULL,
    "replying" text,
    "content" text,
    "media" text[] DEFAULT array[]::text[],
    "likes" text[] DEFAULT array[]::text[],
    "shares" text[] DEFAULT array[]::text[]
);

CREATE INDEX posters ON posts ("user");

CREATE TABLE IF NOT EXISTS shares (
  "id" text NOT NULL,
  "user" text NOT NULL,
  "date" timestamp NOT NULL,
  PRIMARY KEY ("id", "user"),
  FOREIGN KEY ("id") REFERENCES posts("id")
);

CREATE INDEX sharers ON shares ("user");