# Main Database

Use PostgreSql. Database: `austrody`

![Entity Relation Diagram](./ER.svg)

## TYPES

- vsb *ENUM*: visibility of post
- img: image

```sql
CREATE TYPE "vsb" AS ENUM ('public', 'private', 'direct');

CREATE TYPE "img" AS (
  "type" text,
  "url" text,
  "alt" text
);
```

## INDEX

- `<this>_out`: for constructing timeline
- `<this>_ntf`: for constructing notification
- `<other>_<field>`: foreign primary for joining
- `<this>_<field>`: self field for query

## TABLE: `users`

- username *PRIMARY*: `text`
- foreign: `boolean` marking whether a foreign user
- fid: `text` as foreign user's federal id
- nickname: `text`
- summary: `text`
- avatar *NULLABLE*: `img`
- lock: `boolean` indicating whether a locked account
- pub_key: `text` as user's public key (RSA)
- pri_key: `text` as local user's private key (RSA)

CONSTRAINT:

- if not `foreign`, `pri_key` must not null.

> NOTE: for foreign user, `username` syntax is `username@domain`.

```sql
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

CREATE INDEX "user_fid" ON "users"("fid");

CREATE VIEW "user_content" AS 
    SELECT "user", "tgt" AS "fid", "date", "vsb", "user" AS "shared_by"
    FROM "share"
  UNION
    SELECT "user", "fid", "date", "vsb", NULL AS "shared_by"
    FROM "posts"
  ORDER BY "date" DESC;
```

### Query all status of a user

> It's necessary to filter by `vsb`

```sql
WITH "l" AS (
  SELECT
    "tgt" AS "id", "date" AS "act",
    ${user} AS "shared_by", NULL AS "reply_to"
  FROM "share" WHERE "user" = ${user}
UNION
  SELECT
    "posts"."id", "posts"."date" AS "act",
    NULL AS "shared_by", "reply"."tgt_user" AS "reply_to"
  FROM "posts" LEFT JOIN "reply"
    ON "posts"."user" = ${user} AND "reply"."id" = "posts"."id"
)
SELECT * FROM "l"
-- filtered by date
ORDER BY "act" DESC
-- limit size
;
```

## TABLE: `user_preferences`

- username *PRIMARY, FOREIGN*: `text`, referencing `users.username`
- password: `text` as encrypted password
- preferences: `jsonb`

> NOTE: Only local users

```sql
CREATE TABLE IF NOT EXISTS "user_preferences" (
  "username" text,
  "password" text,
  "preferences" jsonb DEFAULT '{"postVsb":"public","shareVsb":"public"}',
  PRIMARY KEY ("username"),
  FOREIGN KEY ("username") REFERENCES "users"("username")
);
```

## TABLE: `foreign_inboxes`

- username *PRIMARY, FOREIGN*: `text`, referencing `users.username`
- inbox: `text` as user inbox url
- shared *NULLABLE*: `text` as instance inbox url

```sql
CREATE TABLE IF NOT EXISTS "foreign_inboxes" (
  "username" text,
  "inbox" text NOT NULL,
  "shared" text,
  PRIMARY KEY ("username"),
  FOREIGN KEY ("username") REFERENCES "users"("username")
);
```

## TABLE: `follow`

- from *PRIMARY, FOREIGN*: `text`, referencing `users.username`
- to *PRIMARY, FOREIGN*: `text`, referencing `users.username`
- date: `timestamp`
- fid *UNIQUE*: `text` as federal id of this activity
- is_req: `bool` whether is still a unhandled request

CONSTRAINT:

- `from` != `to`

```sql
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

CREATE INDEX "user_follows" ON "follow" ("from");
CREATE INDEX "user_be_followed" ON "follow" ("to");
CREATE INDEX "follow_ntf" ON "follow"("to", "date");
CREATE INDEX "follow_fid" ON "follow"("fid");

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
```

## TABLE: `posts`

- id *PRIMARY*: `text` as uuid
- tombstone: `boolean` as whether a removed post
- fid *UNIQUE*: `text` as federal id of this post.
- url: `text` as url
- user *FOREIGN*: `text`, referencing `user.username`
- date: `timestap`
- vsb: `vsb` as visibility of post
- content: `text`
- media: `img[]` as images attaching to this post

```sql
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

CREATE INDEX "post_out" ON "posts" ("user", "date");
CREATE INDEX "post_fid" ON "posts"("fid");

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
```

## TABLE: `reply`

- id *PRIMARY, FOREIGN*: `text` as uuid of the replying post, referencing to `posts.id`
- tgt *PRIMARY, FOREIGN*: `text` as uuid of the post replied, referencing to `posts.id`
- tgt_user *FOREIGN*: `text` as publisher of the target post, referencing to `users.username`
- date: `timestamp`

```sql
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

CREATE INDEX "post_replies" ON "reply"("tgt");
CREATE INDEX "reply_ntf" ON "reply"("tgt_user", "date");
```

### Query

```sql
-- replying
WITH RECURSIVE "rs" AS (
    SELECT "id", "tgt" FROM "reply" WHERE "id" = ${postID}
  UNION ALL
    SELECT "reply"."id", "reply"."tgt" FROM "reply"
    JOIN "rs" ON "rs"."tgt" = "reply"."id"
)
SELECT "id", "tgt" FROM "rs";

-- replies
WITH RECURSIVE "rs" AS (
    SELECT "id", "tgt" FROM "reply" WHERE "tgt" = ${postID}
  UNION ALL
    SELECT "reply"."id", "reply"."tgt" FROM "reply"
    JOIN "rs" ON "rs"."id" = "reply"."tgt"
)
SELECT "id", "tgt" FROM "rs";
```

## TABLE: `like`

- user *PRIMARY*: `text`, referencing `user.username`
- tgt *PRIMARY, FOREIGN, INDEX*: `text` as uuid, referencing to `posts.id`
- tgt_user *FOREIGN*: `text` as publisher of the target post, referencing `user.username`
- date: `timestamp`
- fid *UNIQUE*: `text` as federal id of this activity

```sql
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

CREATE INDEX "post_likes" ON "like"("tgt");
CREATE INDEX "like_ntf" ON "like"("tgt_user", "date");
CREATE INDEX "like_fid" ON "like"("fid");
```

## TABLE: `share`

- user *PRIMARY, FOREIGN*: `text` as username
- tgt *PRIMARY, FOREIGN*: `text` as uuid, referencing to `posts."id"`
- tgt_user *FOREIGN*: `text` as publisher of the target post, referencing `user.username`
- vsb: `vsb` as visibility of sharing
- date: `timestamp`
- fid *UNIQUE*: `text` as federal id of this activity

```sql
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

CREATE INDEX "post_shares" ON "share"("tgt");
CREATE INDEX "share_fid" ON "share"("fid");
CREATE INDEX "share_out" ON "share"("user", "date");
CREATE INDEX "share_ntf" ON "share"("tgt_user", "date");
```