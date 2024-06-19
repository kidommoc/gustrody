package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
)

// models

type Post struct {
	ID       string    `json:"id"`
	User     string    `json:"user"`
	Date     time.Time `json:"date"`
	Content  string    `json:"content"`
	Media    []string  `json:"media"`    // attached media url
	Replying string    `json:"replying"` // post id
	ReplyTo  string    `json:"replyTo"`  // user id, temporary field
	SharedBy string    `json:"sharedBy"` // user id, temporary field
	Likes    int64     `json:"likes"`    // count, temporary field
	Shares   int64     `json:"shares"`   // count, temporary field
	ActDate  string    `json:"actDate"`  // temporary field, used in sort
	Level    int       `json:"level"`    // temporary field, used in replying and replies
}

// db

type IPostDb interface {
	IsPostExist(id string) bool
	QueryPostByID(id string) (post Post, err utils.Error)
	QueryPostReplies(id string) (replyings []*Post, replies []*Post, err utils.Error)
	QueryPostsAndSharesByUser(user string, asec bool) (list []*Post, err utils.Error)
	SetPost(id string, user string, replying string, content string, attachments []string) utils.Error
	UpdatePost(id string, content string, attachments []string) utils.Error
	RemovePost(id string) utils.Error
	QueryLikes(id string) (list []string, err utils.Error)
	SetLike(user string, id string) utils.Error
	RemoveLike(user string, id string) utils.Error
	QueryShares(id string) (list []string, err utils.Error)
	SetShare(user string, id string) utils.Error
	RemoveShare(user string, id string) utils.Error
}

type PostDb struct {
	pool *_db.ConnPool[*_db.PqConn]
}

var postsIns *PostDb = nil

func PostInstance() *PostDb {
	if postsIns == nil {
		postsIns = &PostDb{
			pool: _db.MainPool(),
		}
	}
	return postsIns
}

// functions

func insertArray(q string, l int, start int) string {
	if l == 0 {
		return fmt.Sprintf(q, "NULL")
	}
	s := make([]string, 0, l)
	for i := 0; i < l; i += 1 {
		s = append(s, fmt.Sprintf("$%d", i+start))
	}
	return fmt.Sprintf(q, fmt.Sprintf("ARRAY[%s]", strings.Join(s, ", ")))
}

func (db *PostDb) IsPostExist(id string) bool {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return false
	}
	defer conn.Close()

	qs := `SELECT 1
		   FROM posts
		   WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)
	var n int
	if e := r.Scan(&n); e != nil {
		switch e {
		case sql.ErrNoRows:
			return false
		default:
			logger.Error("[Model.Posts] Cannot query", newErr(ErrDbInternal, e.Error()))
			return false
		}
	}
	return true
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryPostByID(id string) (post Post, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return post, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := `SELECT
		     "id", "user", "date", "content", "media"
		     CARDINALITY("likes") as "likes",
		     CARDINALITY("shares") as "shares"
		   FROM posts
		   WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)
	post = Post{}
	var media []string
	if e := r.Scan(
		&post.ID, &post.User, &post.Date, &post.Content, pq.Array(&media),
		&post.Likes, &post.Shares,
	); e != nil {
		switch e {
		case sql.ErrNoRows:
			return post, newErr(ErrNotFound, "post")
		default:
			err = newErr(ErrDbInternal, e.Error())
			logger.Error("[Model.Posts] Cannot scan row", err)
			return post, err
		}
	}
	if media == nil {
		media = []string{}
	}
	post.Media = media
	return post, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryPostReplies(id string) (replyings []*Post, replies []*Post, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Reply] Failed to open a connection", err)
		return nil, nil, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return nil, nil, newErr(ErrNotFound, "post")
	}

	qs := ` WITH RECURSIVE rt AS (
			    SELECT "id", "replying", 0 AS "level" FROM posts
			    WHERE "id" = $1
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
	`
	r, e := conn.Query(qs, id)
	if e != nil {
		return nil, nil, newErr(ErrDbInternal, e.Error())
	}
	replyings = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpy sql.NullString
		var media []string
		if e := r.Scan(
			&p.ID, &p.User, &p.Date,
			&p.Content, pq.Array(&media),
			&p.Likes, &p.Shares,
			&rpy, &p.Level,
		); e != nil {
			logger.Error("[Model.Reply] Cannot scan row", newErr(ErrDbInternal, e.Error()))
			continue
		}
		if rpy.Valid {
			p.Replying = rpy.String
		}
		if media == nil {
			media = []string{}
		}
		p.Media = media
		replyings = append(replyings, &p)
	}

	qs = `  WITH RECURSIVE rs AS (
			    SELECT "id", "replying", 0 AS "level" FROM posts
			    WHERE "id" = $1
			  UNION ALL
			    SELECT posts."id", posts."replying", rs."level" + 1
			    FROM posts
			      JOIN rs ON rs."id" = posts."replying"
			)
			SELECT
			  posts."id", posts."user", posts."date",
			  posts."content", posts."media"
			  CARDINALITY(posts."likes") as "likes",
			  CARDINALITY(posts."shares") as "shares",
			  posts."replying", rs."level"
			FROM posts
			  JOIN rs ON posts."id" = rs."id"
			ORDER BY "level" ASC, "date" DESC;
	`
	r, e = conn.Query(qs, id)
	if e != nil {
		return nil, nil, newErr(ErrDbInternal, e.Error())
	}
	replies = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpy sql.NullString
		var media []string
		if e := r.Scan(
			&p.ID, &p.User, &p.Date,
			&p.Content, pq.Array(&media),
			&p.Likes, &p.Shares,
			&rpy, &p.Level,
		); e != nil {
			logger.Error("[Model.Reply] Cannot scan row", newErr(ErrDbInternal, e.Error()))
			continue
		}
		if rpy.Valid {
			p.Replying = rpy.String
		}
		if media == nil {
			media = []string{}
		}
		p.Media = media
		replies = append(replies, &p)
	}

	return replyings, replies, nil
}

// ERRORS
//
//   - DbInternal
func (db *PostDb) QueryPostsAndSharesByUser(user string, asc bool) (list []*Post, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return nil, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := `   WITH rr AS (
			    SELECT p1."id", p2."user" as "user"
			    FROM posts AS p1, posts AS p2
			    WHERE p1."user" = $1 AND p2."id" = p1."replying"
			  )
			  SELECT
			    posts."id", posts."user", posts."date",
				posts."content", posts."media",
			    CARDINALITY("likes") as "likes",
			    CARDINALITY("shares") as "shares",
			    rr."user" AS "replyTo", NULL AS "sharedBy",
			    posts."date" AS "act"
			  FROM posts, rr
			  WHERE posts."user" = $1 AND posts."id" = rr."id"
			UNION ALL
			  SELECT
			    "id", "user", "date", "content", "media",
			    CARDINALITY("likes") as "likes",
			    CARDINALITY("shares") as "shares",
			    NULL AS "replyTo", NULL AS "sharedBy",
			    "date" AS "act"
			  FROM posts
			  WHERE "user" = $1 AND "replying" IS NULL
			UNION ALL
			  SELECT
			    posts."id", posts."user", posts."date",
				posts."content", posts."media",
			    CARDINALITY("likes") as "likes",
			    CARDINALITY("shares") as "shares",
			    NULL AS "replyTo", shares."user" as "sharedBy",
			    shares."date" AS "act"
			  FROM posts, shares
			  WHERE shares."user" = $1 AND posts."id" = shares."id"
			ORDER BY "act" %s;`
	if asc {
		qs = fmt.Sprintf(qs, "ASC")
	} else {
		qs = fmt.Sprintf(qs, "DESC")
	}
	r, e := conn.Query(qs, user)
	if e != nil {
		return nil, newErr(ErrDbInternal, e.Error())
	}
	list = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpt sql.NullString
		var shb sql.NullString
		var media []string
		if e := r.Scan(
			&p.ID, &p.User, &p.Date, &p.Content, pq.Array(&media),
			&p.Likes, &p.Shares,
			&rpt, &shb, &p.ActDate,
		); e != nil {
			logger.Error("[Model.Posts] Cannot scan row", newErr(ErrDbInternal, e.Error()))
			continue
		}
		if rpt.Valid {
			p.ReplyTo = rpt.String
		}
		if shb.Valid {
			p.SharedBy = shb.String
		}
		if media == nil {
			media = []string{}
		}
		p.Media = media
		list = append(list, &p)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - Dunplicate "post"
func (db *PostDb) SetPost(id string, user string, replying string, content string, attachments []string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if replying != "" && !db.IsPostExist(replying) {
		return newErr(ErrNotFound, "replying")
	}

	qs := ` INSERT INTO posts(
			  "id", "user", "date",
			  "replying", "content", "media"
			)
			VALUES (
			  $1, $2, NOW(),
			  $3, $4, %s
			);`
	qs = insertArray(qs, len(attachments), 5)
	args := make([]interface{}, 0, 4+len(attachments))
	args = append(args, id)
	args = append(args, user)
	if replying == "" {
		args = append(args, nil)
	} else {
		args = append(args, replying)
	}
	args = append(args, content)
	for _, v := range attachments {
		args = append(args, v)
	}
	r, e := conn.Exec(qs, args...)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrDunplicate, "post")
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) UpdatePost(id string, content string, attachments []string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := ` UPDATE posts
			SET
			  "date" = NOW(), "content" = $1,
			  "media" = %s
			WHERE "id" = $2;`
	qs = insertArray(qs, len(attachments), 3)
	args := make([]interface{}, 0, len(attachments))
	args = append(args, content)
	args = append(args, id)
	for _, v := range attachments {
		args = append(args, v)
	}
	r, e := conn.Exec(qs, args...)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrNotFound, "post")
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) RemovePost(id string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return newErr(ErrNotFound, "post")
	}

	qs := ` DELETE FROM posts
			WHERE "id" = $1;`
	r, e := conn.Exec(qs, id)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrNotFound, "post")
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryLikes(id string) (list []string, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := ` SELECT "likes"
			FROM posts
			WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)
	if e := r.Scan(pq.Array(&list)); e != nil {
		switch e {
		case sql.ErrNoRows:
			return nil, newErr(ErrNotFound, "post")
		default:
			err = newErr(ErrDbInternal, e.Error())
			logger.Error("[Model.Like] Cannot scan row", err)
			return nil, err
		}
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
//   - Dunplicate "like"
func (db *PostDb) SetLike(user string, id string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Like] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return newErr(ErrNotFound, "post")
	}

	qs := ` UPDATE posts
			SET "likes" = ARRAY_APPEND("likes", $1)
			WHERE
			  "id" = $2
  			  AND ARRAY_POSITION("likes", $1) IS NULL;`
	r, e := conn.Exec(qs, user, id)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrDunplicate, "like")
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) RemoveLike(user string, id string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Like] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := ` UPDATE posts
			SET "likes" = ARRAY_REMOVE("likes", $1)
			WHERE "id" = $2;`
	r, e := conn.Exec(qs, user, id)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrNotFound, "post")
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryShares(id string) (list []string, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Share] Failed to open a connection", err)
		return nil, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := ` SELECT "shares"
			FROM posts
			WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)
	if e := r.Scan(pq.Array(&list)); e != nil {
		switch e {
		case sql.ErrNoRows:
			return nil, newErr(ErrNotFound)
		default:
			err = newErr(ErrDbInternal, e.Error())
			logger.Error("[Model.Like] Cannot scan row", err)
			return nil, err
		}
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
//   - Dunplicate "share"
func (db *PostDb) SetShare(user string, id string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Share] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return newErr(ErrNotFound, "post")
	}

	tx, e := conn.BeginTx()
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}

	// update posts.shares
	qs := ` UPDATE posts
			SET "shares" = ARRAY_APPEND("shares", $1)
			WHERE
			  "id" = $2 AND
			  AND ARRAY_POSITION("shares", $1) IS NULL;
	`
	r, e := tx.Exec(qs, user, id)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrDunplicate, "share")
	}

	// insert into shares
	qs = `  INSERT INTO shares
			VALUES ($1, $2, NOW());
	`
	_, e = tx.Exec(qs, user, id)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrDunplicate, "share")
	}

	if e := tx.Commit(); e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post", "share"
func (db *PostDb) RemoveShare(user string, id string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Share] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return newErr(ErrNotFound, "post")
	}

	tx, e := conn.BeginTx()
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}

	// use 2 annoymous func to ensure completely deletion

	err1 := func() utils.Error {
		qs := ` UPDATE posts
				SET "shares" = ARRAY_REMOVE("shares", $1)
				WHERE "id" = $2;
		`
		r, e := tx.Exec(qs, user, id)
		if e != nil {
			logger.Error("[Model.Share] Failed to exec", newErr(ErrDbInternal, e.Error()))
			return newErr(ErrDbInternal, e.Error())
		}
		if r == 0 {
			return newErr(ErrNotFound)
		}
		return nil
	}()

	err2 := func() utils.Error {
		qs := ` DELETE FROM shares
				WHERE "id" = $1 and "user" = $2;`
		r, e := tx.Exec(qs, id, user)
		if e != nil {
			logger.Error("[Model.Share] Failed to exec", newErr(ErrDbInternal, e.Error()))
			return newErr(ErrDbInternal, e.Error())
		}
		if r == 0 {
			return newErr(ErrNotFound)
		}
		return nil
	}()

	if e := tx.Commit(); e != nil {
		return newErr(ErrDbInternal, e.Error())
	}

	if err1 != nil || err2 != nil {
		return newErr(ErrNotFound, "share")
	} else {
		return nil
	}
}
