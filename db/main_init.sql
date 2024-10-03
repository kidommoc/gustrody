CREATE TYPE vsb AS ENUM ('public', 'private', 'direct');

CREATE TYPE img AS (
  "type" text,
  "url" text,
  "alt" text
);

-- USER

CREATE TABLE IF NOT EXISTS "users" (
  "username" text,
  "foreign" boolean NOT NULL,
  "fid" text NOT NULL UNIQUE,
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
  "date" timestamp,
  "fid" text UNIQUE,
  "is_req" boolean DEFAULT FALSE,
  PRIMARY KEY ("from", "to"),
  FOREIGN KEY ("from") REFERENCES "users"("username"),
  FOREIGN KEY ("to") REFERENCES "users"("username")
);

-- USER: INDEX and VIEW

CREATE INDEX "user_fid" ON "users"("fid");
CREATE INDEX "follow_fid" ON "follow"("fid");

CREATE INDEX "user_follows" ON "follow" ("from");
CREATE INDEX "user_be_followed" ON "follow" ("to");

CREATE INDEX "follow_ntf" ON "follow"("to", "date");

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

-- POST

CREATE TABLE IF NOT EXISTS "posts" (
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

CREATE TABLE IF NOT EXISTS "reply" (
  "id" varchar(36) UNIQUE,
  "tgt" varchar(36),
  "tgt_user" text,
  "date" timestamp,
  PRIMARY KEY ("id", "tgt"),
  FOREIGN KEY ("id") REFERENCES "posts"("id"),
  FOREIGN KEY ("tgt") REFERENCES "posts"("id"),
  FOREIGN KEY ("tgt_user") REFERENCES "users"("username")
);

CREATE TABLE IF NOT EXISTS "like" (
  "user" text,
  "tgt" varchar(36),
  "tgt_user" text,
  "date" timestamp,
  "fid" text UNIQUE,
  PRIMARY KEY ("user", "tgt"),
  FOREIGN KEY ("user") REFERENCES "users"("username"),
  FOREIGN KEY ("tgt") REFERENCES "posts"("id"),
  FOREIGN KEY ("tgt_user") REFERENCES "users"("username")
);

CREATE TABLE IF NOT EXISTS "share" (
  "user" text NOT NULL,
  "tgt" varchar(36) NOT NULL,
  "tgt_user" text,
  "vsb" vsb NOT NULL,
  "date" timestamp NOT NULL,
  "fid" text UNIQUE,
  PRIMARY KEY ("user", "tgt"),
  FOREIGN KEY ("user") REFERENCES "users"("username"),
  FOREIGN KEY ("tgt") REFERENCES "posts"("id"),
  FOREIGN KEY ("tgt_user") REFERENCES "users"("username")
);

-- POST: INDEX and VIEW

CREATE INDEX "post_replies" ON "reply"("tgt");
CREATE INDEX "post_likes" ON "like"("tgt");
CREATE INDEX "post_shares" ON "share"("tgt");

CREATE INDEX "post_fid" ON "posts"("fid");
CREATE INDEX "like_fid" ON "like"("fid");
CREATE INDEX "share_fid" ON "share"("fid");

CREATE INDEX "post_out" ON "posts" ("user", "date");
CREATE INDEX "share_out" ON "share"("user", "date");

CREATE INDEX "reply_ntf" ON "reply"("tgt_user", "date");
CREATE INDEX "like_ntf" ON "like"("tgt_user", "date");
CREATE INDEX "share_ntf" ON "share"("tgt_user", "date");

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
      "posts"."id",
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