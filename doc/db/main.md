# Main Database

Use PostgreSql. Database: `austrody`

## TABLE: user_info

- username *PRIMARY*: `varchar(20)`
- nickname *NULLABLE*: `text`
- summary *NULLABLE*: `text`
- createdAt: `timestamp`
- icon *NULLABLE*: `text` as url

```sql
CREATE TABLE IF NOT EXISTS users (
    "username" varchar(20) PRIMARY KEY,
    "nickname" text,
    "summary" text,
    "createdAt" timestamp NOT NULL,
    "icon" text
);
```

### Queries

- query a user's information

```sql
SELECT
  "username", "nickname", "summary"
FROM users
WHERE "username" = ${username};
```

- insert a new user

```sql
INSERT INTO users("username", "createAt")
VALUES (${username}, NOW());
```

- set user information

```sql
UPDATE users
SET
  "nickname" = ${nickname},
  "summary" = ${summary},
WHERE "username" = ${username};
```

## TABLE: follow

- user *PRIMARY, INDEX*: `text`
- follows *PRIMARY, INDEX*: `text`

CONSTRAINT:

- `"user"` != `follows`

```sql
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

- id *PRIMARY*: `text` as url (generated with uuid)
- user *INDEX*: `text`
- replying *NULLABLE*: `text` as id of the post replied
- date: `timestamp`
- content: `text`
- likes: `text[]` as id of the users liking this post
- shares: `text[]` as id of the users sharing this post

```sql
CREATE TABLE IF NOT EXISTS posts (
    "id" text PRIMARY KEY,
    "user" text NOT NULL,
    "date" timestamp NOT NULL,
    "replying" text,
    "content" text,
    "media" text[4] DEFAULT array[4]::text[4],
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
  "id", "user", "date",
  "replying", "content",
  "media"
)
VALUES (
  ${postID}, ${username}, NOW(),
  ${replying}, ${content},
  ARRAY[${mediaUrl}, ...]
);
```

- update a post

```sql
UPDATE posts
SET
  "content" = ${content}, "date" = NOW(),
  "media" = ARRAY[${mediaUrl}, ...]
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
  "id", "user", "date", "content", "media",
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
  posts."id", posts."user", posts."date",
  posts."content", posts."media",
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
  posts."id", posts."user", posts."date",
  posts."content", posts."media",
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
    posts."id", posts."user", posts."date",
    posts."content", posts."media",
    CARDINALITY("likes") as "likes",
    CARDINALITY("shares") as "shares",
    rr."user" AS "replyTo", NULL AS "sharedBy",
    posts."date" AS "act"
  FROM posts, rr
  WHERE posts."user" = ${username} AND posts."id" = rr."id"
UNION ALL
  SELECT
    "id", "user", "date", "content", "media",
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

- id *PRIMARY, FOREIGN*: `text` as url, referencing to `posts."id"`
- user *PRIMARY, INDEX*: `text`
- date: `timestamp`

```sql
CREATE TABLE IF NOT EXISTS shares (
  "id" text NOT NULL,
  "user" text NOT NULL,
  "date" timestamp NOT NULL,
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
  posts."id", posts."user", posts."date",
  posts."content", posts."media",
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
INSERT INTO shares
VALUES (${postID}, ${username}, NOW());

-- UNSET
DELETE FROM shares
WHERE "id" = ${postID} and "user" = ${username};
```