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

CREATE TYPE img AS (
  "type" text,
  "url" text,
  "alt" text
);

CREATE TABLE IF NOT EXISTS "users" (
  "username" text,
  "foreign" boolean NOT NULL,
  "id" text NOT NULL UNIQUE,
  "nickname" text NOT NULL,
  "summary" text DEFAULT '',
  "avatar" img,
  "lock" boolean DEFAULT FALSE,
  "pub_key" text NOT NULL,
  "pri_key" text,
  PRIMARY KEY ("username"),
  CONSTRAINT "foreign_user" CHECK (
    ("foreign" = FALSE AND "pri_key" <> NULL) OR "foreign" = TRUE
  )
);

CREATE INDEX "user_id" ON "users"("id");

CREATE TABLE IF NOT EXISTS "user_preferences" (
  "username" text,
  "password" text,
  "preferences" jsonb DEFAULT '{"postVsb":"public","shareVsb":"public"}',
  PRIMARY KEY ("username"),
  FOREIGN KEY ("username") REFERENCES "users"("username")
);

CREATE TABLE IF NOT EXISTS "foreign_inboxes" (
  "username" text,
  "inbox" text NOT NULL,
  "shared" text,
  PRIMARY KEY ("username"),
  FOREIGN KEY ("username") REFERENCES "users"("username")
);

CREATE TABLE IF NOT EXISTS "follow" (
  "from" text,
  "to" text CHECK ("to" <> "from"),
  PRIMARY KEY ("from", "to"),
  FOREIGN KEY ("from") REFERENCES "users"("username"),
  FOREIGN KEY ("to") REFERENCES "users"("username")
);

CREATE INDEX "user_follows" ON "follow" ("from");
CREATE INDEX "user_be_followed" ON "follow" ("to");

CREATE VIEW "follow_data" ("user", "followings", "followers") AS
  WITH "followings" AS (
    SELECT "from" AS "u", COUNT(*) AS "c"
    FROM "follow" GROUP BY "u"
  ), "followers" AS (
    SELECT "to" AS "u", COUNT(*) AS "c"
    FROM "follow" GROUP BY "u"
  )
  SELECT
    "users"."username" AS "user",
    COALESCE("followings"."c", 0) AS "followings",
    COALESCE("followers"."c", 0) AS "followers"
  FROM "users"
  FULL JOIN "followings"
    ON "users"."username" = "followings"."u"
  FULL JOIN "followers"
    ON "users"."username" = "followers"."u";

CREATE TABLE IF NOT EXISTS posts (
  "id" varchar(36) NOT NULL,
  "tombstone" boolean DEFAULT FALSE,
  "fid" text NOT NULL UNIQUE,
  "user" text NOT NULL,
  "date" timestamp NOT NULL,
  "vsb" vsb NOT NULL,
  "content" text NOT NULL,
  "media" img[] DEFAULT array[]::img[],
  PRIMARY KEY ("id"),
  FOREIGN KEY ("user") REFERENCES "users"("username")
);

CREATE INDEX "posters" ON "posts" ("user");

CREATE TABLE IF NOT EXISTS "reply" (
  "id" varchar(36) UNIQUE,
  "tgt" varchar(36),
  "tgt_user" text,
  PRIMARY KEY ("id", "tgt"),
  FOREIGN KEY ("id") REFERENCES "posts"("id"),
  FOREIGN KEY ("tgt") REFERENCES "posts"("id"),
  FOREIGN KEY ("tgt_user") REFERENCES "users"("username")
);

CREATE INDEX "post_replies" ON "reply"("tgt");

CREATE TABLE IF NOT EXISTS "like" (
  "user" text,
  "tgt" varchar(36),
  PRIMARY KEY ("user", "tgt"),
  FOREIGN KEY ("user") REFERENCES "users"("username"),
  FOREIGN KEY ("tgt") REFERENCES "posts"("id")
);

CREATE INDEX "post_likes" ON "like"("tgt");

CREATE TABLE IF NOT EXISTS "share" (
  "user" text NOT NULL,
  "tgt" varchar(36) NOT NULL,
  "date" timestamp NOT NULL,
  "vsb" vsb NOT NULL,
  PRIMARY KEY ("user", "tgt"),
  FOREIGN KEY ("user") REFERENCES "users"("username"),
  FOREIGN KEY ("tgt") REFERENCES "posts"("id")
);

CREATE INDEX "post_shares" ON "share"("tgt");
CREATE INDEX "sharer" ON "share"("user");

CREATE VIEW "post_data" AS
  WITH "r" AS (
    SELECT "tgt", COUNT(*) AS "c"
    FROM "reply" GROUP BY "tgt"
  ),
  "s" AS (
    SELECT "tgt", COUNT(*) AS "c"
    FROM "share" GROUP BY "tgt"
  ),
  "l" AS (
    SELECT "tgt", COUNT(*) AS "c"
    FROM "like" GROUP BY "tgt"
  ), "ps" AS (
    SELECT
      "posts".*,
      COALESCE("r"."c", 0) AS "replies",
      COALESCE("s"."c", 0) AS "shares",
      COALESCE("l"."c", 0) AS "likes"
    FROM "posts"
    FULL JOIN "r" ON "posts"."id" = "r"."tgt"
    FULL JOIN "s" ON "posts"."id" = "s"."tgt"
    FULL JOIN "l" ON "posts"."id" = "l"."tgt"
  )
  SELECT "ps".*, "reply"."tgt" AS "replying", "reply"."tgt_user" AS "reply_to"
  FROM "ps" LEFT JOIN "reply" ON "ps"."id" = "reply"."id";

CREATE VIEW "user_content" AS 
    SELECT "user", "tgt" AS "id", "date", "vsb", "user" AS "shared_by"
    FROM "share"
  UNION
    SELECT "user", "id", "date", "vsb", NULL AS "shared_by"
    FROM "posts"
  ORDER BY "date" DESC;