# Main Database

Use PostgreSql. Database: `austrody`

## TYPEs

- vsb: visibility of post
- kp: user's encryption key pair (RSA)
- img: image

```sql
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
```

## TABLE: users

- username *PRIMARY*: `varchar(20)`
- nickname: `text`
- summary *NULLABLE*: `text`
- createdAt: `timestamp`
- avatar *NULLABLE*: `text` as url
- keys: `kp` as user's key pair
- preference: `json`

```sql
CREATE TABLE IF NOT EXISTS users (
  "username" varchar(20) PRIMARY KEY,
  "nickname" text NOT NULL,
  "summary" text,
  "createdAt" timestamp NOT NULL,
  "avatar" text,
  "keys" kp NOT NULL,
  "preference" json DEFAULT '{"postVsb":"public","shareVsb":"public"}'
);
```

### Queries

- query a user's information

```sql
SELECT
  "username", "nickname", "summary", "avatar"
FROM users
WHERE "username" = ${username};
```

- insert a new user

```sql
INSERT INTO users(
  "username", "nickname", "createAt", "keys"
)
VALUES (
  ${username}, ${nickname}, NOW(),
  ROW(${pub_key}, ${pri_key})
);
```

- set user information

```sql
UPDATE users
SET
  "nickname" = ${nickname},
  "summary" = ${summary},
  "avatar" = ${avatar_url}
WHERE "username" = ${username};
```

- query a user's encryption key pair

```sql
SELECT ("keys")."pub", ("keys")."pri"
FROM users
WHERE "username" = ${username};
```

- query a user's preference

```sql
SELECT "preference"
FROM users
WHERE "username" = ${username};
```

- set a user's preference

```sql
UPDATE users
SET "preference" = ${preference_json}
WHERE "username" = ${username};
```

## TABLE: foreign_users

- username *PRIMARY, INDEX*: `varchar(60)`
- inbox: `text` as url
- pub: `text` as RSA public key

```sql
CREATE TABLE IF NOT EXISTS foreign_users (
  "username" varchar(60) PRIMARY KEY,
  "inbox" text NOT NULL,
  "pub" text NOT NULL
);
```

### Queries

- insert a foreign user

```sql
INSERT INTO foreign_users("username", "inbox", "pub")
VALUES (${username}, ${inbox_url}, ${pub_key});
```

- query a foreign user

```sql
```

## TABLE: follow

- from *PRIMARY, INDEX*: `varchar(60)`
- to *PRIMARY, INDEX*: `varchar(60)`

CONSTRAINT:

- `"from"` != `to`

```sql
CREATE TABLE IF NOT EXISTS follow (
  "from" varchar(60),
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
```

### Queries

- query a user's followings:

``` sql
SELECT "to" AS "following"
FROM follow
WHERE "from" = ${username};
```

- query a user's followers:

``` sql
SELECT "from" AS "follower"
FROM follow
WHERE "to" = ${username};
```

- query a user's follow data

```sql
SELECT "followings", "followers"
FROM follow_info
WHERE "user" = ${username};
```

- set follow relationship

```sql
-- SET
INSERT INTO follow
VALUES (${from}, ${to});

-- UNSET
DELETE FROM follow
WHERE "from" = ${from} AND "to" = ${to};
```

## TABLE: posts

- id *PRIMARY*: `text` as uuid
- url: `text` as url
- date: `timestap`
- user *INDEX*: `varchar(60)`
- replying *NULLABLE*: `text` as id of the post replied
- vsb: `vsb` as visibility of post
- content: `text`
- media: `img[]` as images attaching to this post
- likes: `text[]` as id of the users liking this post
- shares: `text[]` as id of the users sharing this post

```sql
CREATE TABLE IF NOT EXISTS posts (
  "id" varchar(36) NOT NULL,
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
```

*Note*: `posts."user"` is not a foreign key to `users."username"`. After applying federal protocol, there will be posts from foreign sites storing in `posts` table, which cannot refer to a user in `users`.

### Queries

- insert a post

```sql
INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying",
  "media"
)
VALUES (
  ${postID}, ${url}, ${username}, ${date},
  ${replying}, ${vsb}, ${content},
  ARRAY[
    ROW(${mediaUrl}, ${alt_text}), ...
  ]
);
```

- update a post

```sql
UPDATE posts
SET
  "date" = ${date}, "content" = ${content}, 
  "media" = ARRAY[
    ROW(${mediaUrl}, ${alt_text}), ...
  ]
WHERE "id" = ${postID};
```

- delete a post

```sql
DELETE FROM posts
WHERE "id" = ${postID};
```

- query a post and it's replyings and replies

```sql
-- QUERY post
SELECT
  "id", "url", "user", "date",
  "vsb", "content", "media",
  CARDINALITY("likes") as "likes",
  CARDINALITY("shares") as "shares"
FROM posts
WHERE "id" = {postID};

-- QUERY replyings
WITH RECURSIVE rt AS (
    SELECT "id", "replying", 0 AS "level" FROM posts
    WHERE "id" = ${postID}
  UNION ALL
    SELECT posts."id", posts."replying", rt."level" + 1
    FROM posts
      JOIN rt ON posts."id" = rt."replying"
)
SELECT
  posts."id", posts."url", posts."user", posts."date",
  posts."vsb", posts."content", posts."media",
  CARDINALITY(posts."likes") as "likes",
  CARDINALITY(posts."shares") as "shares",
  posts."replying", rt."level"
FROM posts
  JOIN rt ON posts."id" = rt."id"
ORDER BY "level" ASC, "date" DESC;

-- QUERY replies
WITH RECURSIVE rs AS (
    SELECT "id", "replying", 0 AS "level" FROM posts
    WHERE "id" = ${postID}
  UNION ALL
    SELECT posts."id", posts."replying", rs."level" + 1
    FROM posts
      JOIN rs ON rs."id" = posts."replying"
)
SELECT
  posts."id", posts."url", posts."user", posts."date",
  posts."vsb", posts."content", posts."media",
  CARDINALITY(posts."likes") as "likes",
  CARDINALITY(posts."shares") as "shares",
  posts."replying", rs."level"
FROM posts
  JOIN rs ON posts."id" = rs."id"
ORDER BY "level" ASC, "date" DESC;
```

- query all posts of a user

```sql
  WITH rr AS (
    SELECT p1."id", p2."user" as "user"
    FROM posts AS p1, posts AS p2
    WHERE p1."user" = ${username} AND p2."id" = p1."replying"
  )
  SELECT
    posts."id", posts."url", posts."user", posts."date",
    posts."vsb", posts."content", posts."media",
    CARDINALITY("likes") as "likes",
    CARDINALITY("shares") as "shares",
    rr."user" AS "replyTo", NULL AS "sharedBy",
    posts."date" AS "act"
  FROM posts, rr
  WHERE posts."user" = ${username} AND posts."id" = rr."id"
UNION ALL
  SELECT
    "id", "url", "user", "date",
    "vsb", "content", "media",
    CARDINALITY("likes") as "likes",
    CARDINALITY("shares") as "shares",
    NULL AS "replyTo", NULL AS "sharedBy",
    "date" AS "act"
  FROM posts
  WHERE "user" = ${username} AND "replying" IS NULL
ORDER BY "act" DESC;
```

- query a post's likes

```sql
SELECT "likes"
FROM posts
WHERE "id" = ${postID};
```

- set liking of a post

```sql
-- SET
UPDATE posts
SET "likes" = ARRAY_APPEND("likes", ${username})
WHERE
  "id" = ${postID}
  AND ARRAY_POSITION("likes", ${username}) IS NULL;

-- UNSET
UPDATE posts
SET "likes" = ARRAY_REMOVE("likes", ${username})
WHERE "id" = ${postID};
```

- query a post's shares

```sql
SELECT "shares"
FROM posts
WHERE "id" = ${postID};
```
- set sharing of a post

*NOTE*: These statements should be part of a transaction

```sql
-- SET
UPDATE posts
SET "shares" = ARRAY_APPEND("shares", ${username})
WHERE
  "id" = ${postID}
  AND ARRAY_POSITION("shares", ${username}) IS NULL;

-- UNSET
UPDATE posts
SET "shares" = ARRAY_REMOVE("shares", ${username})
WHERE "id" = ${postID};
```

## TABLE: shares

- id *PRIMARY, FOREIGN*: `text` as uuid, referencing to `posts."id"`
- user *PRIMARY, INDEX*: `varchar(60)`
- date: `timestamp`

```sql
CREATE TABLE IF NOT EXISTS shares (
  "user" varchar(60) NOT NULL,
  "id" varchar(36) NOT NULL,
  "date" timestamp NOT NULL,
  "vsb" vsb NOT NULL,
  PRIMARY KEY ("id", "user"),
  FOREIGN KEY ("id") REFERENCES posts("id")
);

CREATE INDEX sharers ON shares ("user");
```

*Note*: `shares."user"` is not a foreign key to `users."username"`. After applying federal protocol, there will be shares from foreign sites storing in `shares` table, which cannot refer to a user in `users`.

### Queries

- query all shares of a user

```sql
SELECT
  posts."id", posts."url", posts."user", posts."date",
  shares."vsb", posts."content", posts."media",
  CARDINALITY("likes") as "likes",
  CARDINALITY("shares") as "shares",
  NULL AS "replyTo", shares."user" as "sharedBy",
  shares."date" AS "act"
FROM posts, shares
WHERE shares."user" = ${username} AND posts."id" = shares."id"
ORDER BY "act" DESC;
```

- set sharing of a post:

*NOTE*: These statements should be part of a transaction

```sql
-- SET
INSERT INTO shares("user", "id", "date", "vsb")
VALUES (${username}, ${postID}, ${date}, ${vsb});

-- UNSET
DELETE FROM shares
WHERE "user" = ${username} and "id" = ${postID};
```