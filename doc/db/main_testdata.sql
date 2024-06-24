-- drop old tables
-- must init tables afterwards

DROP TABLE users CASCADE;
DROP TABLE follow CASCADE;
DROP TABLE posts CASCADE;
DROP TABLE shares CASCADE;

-- users

INSERT INTO users("username", "nickname", "createdAt", "summary")
VALUES ('u1', 'User 1', NOW(), 'I am u1.');

INSERT INTO users("username", "nickname", "createdAt", "summary")
VALUES ('u2', 'User 2', NOW(), 'I am u2.');

INSERT INTO users("username", "nickname", "createdAt", "summary")
VALUES ('u3', 'User 3', NOW(), 'I am u3.');

-- follow

INSERT INTO follow
VALUES ('u1', 'u3');

INSERT INTO follow
VALUES ('u2', 'u1');

INSERT INTO follow
VALUES ('u2', 'u3');

INSERT INTO follow
VALUES ('u3', 'u2');

-- posts and shares

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  '90344833-7ecc-4ae2-b4ee-1eb1b2f335d1',
  'austrody.sns/posts/90344833-7ecc-4ae2-b4ee-1eb1b2f335d1',
  'u1', NOW(),
  'public', 'p:u1-1', NULL
);

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  'f56b1307-959a-4f03-9422-81eb3d150071',
  'austrody.sns/posts/f56b1307-959a-4f03-9422-81eb3d150071',
  'u2', NOW(),
  'public', 'p:u2-1', NULL
);

-- p:u2-1
INSERT INTO shares("user", "date", "vsb", "id")
VALUES (
  'u1', NOW(), 'public',
  'f56b1307-959a-4f03-9422-81eb3d150071'
);

UPDATE posts
SET "shares" = ARRAY_APPEND("shares", 'u1')
WHERE
  "id" = 'f56b1307-959a-4f03-9422-81eb3d150071'
  AND ARRAY_POSITION("shares", 'u1') IS NULL;

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  'dcbd25aa-3610-4e24-8e86-34105b96359b',
  'austrody.sns/posts/dcbd25aa-3610-4e24-8e86-34105b96359b',
  'u1', NOW(),
  'public', 'p:u1-2', NULL
);

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  '6ae0210e-00b2-4837-abb8-b3a4242bacac',
  'austrody.sns/posts/6ae0210e-00b2-4837-abb8-b3a4242bacac',
  'u2', NOW(),
  'public', 'r:u1-1',
  '90344833-7ecc-4ae2-b4ee-1eb1b2f335d1'
);

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  'b5078927-46a7-482e-9b60-8b602f6e2fe9',
  'austrody.sns/posts/b5078927-46a7-482e-9b60-8b602f6e2fe9',
  'u1', NOW(),
  'public', 'r:u2-r:u1-1',
  '6ae0210e-00b2-4837-abb8-b3a4242bacac'
);

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  '70d94cf7-59f6-466b-bd79-c87512ffc86e',
  'austrody.sns/posts/70d94cf7-59f6-466b-bd79-c87512ffc86e',
  'u3', NOW(),
  'public', 'p:u3-1', NULL
);

INSERT INTO posts(
  "id", "url", "user", "date",
  "vsb", "content", "replying" 
) VALUES (
  'a421c6a6-106f-4f34-a655-4a25fcc3a74c',
  'austrody.sns/posts/a421c6a6-106f-4f34-a655-4a25fcc3a74c',
  'u3', NOW(),
  'public', 'r:u1-1',
  '90344833-7ecc-4ae2-b4ee-1eb1b2f335d1'
);