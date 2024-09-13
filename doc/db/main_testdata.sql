-- drop old tables
-- must init tables afterwards

DROP TABLE "reply" CASCADE;
DROP TABLE "like" CASCADE;
DROP TABLE "share" CASCADE;
DROP TABLE "posts" CASCADE;
DROP TABLE "foreign_inboxes" CASCADE;
DROP TABLE "user_preferences" CASCADE;
DROP TABLE "follow" CASCADE;
DROP TABLE "users" CASCADE;
DROP TYPE img;
DROP TYPE vsb;

-- users

INSERT INTO "users"(
  "username", "id", "foreign", "nickname", "summary", "pub_key", "pri_key"
)
VALUES (
  'u1', 'http://austrody.sns/users/u1', FALSE, 'User 1', 'I am u1.', 'pub', 'pri'
);

INSERT INTO "users"(
  "username", "id", "foreign", "nickname", "summary", "pub_key", "pri_key"
)
VALUES (
  'u2', 'http://austrody.sns/users/u2', FALSE, 'User 2', 'I am u2.', 'pub', 'pri'
);

INSERT INTO "users"(
  "username", "id", "foreign", "nickname", "summary", "pub_key", "pri_key"
)
VALUES (
  'u3', 'http://austrody.sns/users/u3', FALSE, 'User 3', 'I am u3.', 'pub', 'pri'
);

-- follow

INSERT INTO "follow"("from", "to")
VALUES ('u1', 'u3');

INSERT INTO "follow"("from", "to")
VALUES ('u2', 'u1');

INSERT INTO "follow"("from", "to")
VALUES ('u2', 'u3');

INSERT INTO "follow"("from", "to")
VALUES ('u3', 'u2');

-- posts; reply and share

INSERT INTO "posts"(
  "id", "fid", "user", "date", "vsb", "content"
) VALUES (
  '90344833-7ecc-4ae2-b4ee-1eb1b2f335d1',
  'http://austrody.sns/posts/90344833-7ecc-4ae2-b4ee-1eb1b2f335d1',
  'u1', NOW(), 'public', 'p:u1-1'
);

INSERT INTO "posts"(
  "id", "fid", "user", "date", "vsb", "content"
) VALUES (
  'f56b1307-959a-4f03-9422-81eb3d150071',
  'http://austrody.sns/posts/f56b1307-959a-4f03-9422-81eb3d150071',
  'u2', NOW(), 'public', 'p:u2-1'
);

-- p:u2-1
INSERT INTO "share"("user", "date", "vsb", "tgt")
VALUES (
  'u1', NOW(), 'public',
  'f56b1307-959a-4f03-9422-81eb3d150071'
);

INSERT INTO "posts"(
  "id", "fid", "user", "date", "vsb", "content"
) VALUES (
  'dcbd25aa-3610-4e24-8e86-34105b96359b',
  'http://austrody.sns/posts/dcbd25aa-3610-4e24-8e86-34105b96359b',
  'u1', NOW(), 'public', 'p:u1-2'
);

BEGIN TRANSACTION;

  INSERT INTO "posts"(
    "id", "fid", "user", "date", "vsb", "content"
  ) VALUES (
    '6ae0210e-00b2-4837-abb8-b3a4242bacac',
    'http://austrody.sns/posts/6ae0210e-00b2-4837-abb8-b3a4242bacac',
    'u2', NOW(), 'public', 'r:u1-1'
  );
  
  INSERT INTO "reply"("id", "tgt", "tgt_user")
  VALUES('6ae0210e-00b2-4837-abb8-b3a4242bacac', '90344833-7ecc-4ae2-b4ee-1eb1b2f335d1', 'u1');

COMMIT;

BEGIN TRANSACTION;

  INSERT INTO "posts"(
    "id", "fid", "user", "date", "vsb", "content"
  ) VALUES (
    'b5078927-46a7-482e-9b60-8b602f6e2fe9',
    'http://austrody.sns/posts/b5078927-46a7-482e-9b60-8b602f6e2fe9',
    'u1', NOW(), 'public', 'r:u2-r:u1-1'
  );

  INSERT INTO "reply"("id", "tgt", "tgt_user")
  VALUES('b5078927-46a7-482e-9b60-8b602f6e2fe9', '6ae0210e-00b2-4837-abb8-b3a4242bacac', 'u2');

COMMIT;

INSERT INTO "posts"(
  "id", "fid", "user", "date", "vsb", "content"
) VALUES (
  '70d94cf7-59f6-466b-bd79-c87512ffc86e',
  'http://austrody.sns/posts/70d94cf7-59f6-466b-bd79-c87512ffc86e',
  'u3', NOW(), 'public', 'p:u3-1'
);

BEGIN TRANSACTION;

  INSERT INTO "posts"(
    "id", "fid", "user", "date", "vsb", "content"
  ) VALUES (
    'a421c6a6-106f-4f34-a655-4a25fcc3a74c',
    'http://austrody.sns/posts/a421c6a6-106f-4f34-a655-4a25fcc3a74c',
    'u3', NOW(), 'public', 'r:u1-1'
  );

  INSERT INTO "reply"("id", "tgt", "tgt_user")
  VALUES('a421c6a6-106f-4f34-a655-4a25fcc3a74c', '90344833-7ecc-4ae2-b4ee-1eb1b2f335d1', 'u1');

COMMIT