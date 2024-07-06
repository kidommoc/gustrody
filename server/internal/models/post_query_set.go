package models

import (
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/utils"
)

type IPostQuery interface {
	IsPostExist(id string) bool
	QueryPostByID(id string) (post Post, err error)
	QueryPostReplies(id string) (replyings []*Post, replies []*Post, err error)
	QueryPostsAndSharesByUser(user string, asec bool) (list []*Post, err error)
}

type IPostSet interface {
	SetPost(p *Post, attachments []Img) error
	UpdatePost(p *Post, attachments []Img) error
	RemovePost(id string) error
}

func (db *PostDb) IsPostExist(id string) bool {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return false
	}
	defer conn.Close()

	qs := `SELECT 1 FROM posts WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)
	var n int
	if err := r.Scan(&n); err != nil {
		switch err {
		case sql.ErrNoRows:
		default:
			logger.Error("[Model.Posts] Cannot query", err)
		}
		return false
	}
	return true
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryPostByID(id string) (post Post, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return post, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT
			  "id", "url", "user", "date", "vsb", "content", "media",
			  CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares"
			FROM posts
			WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)

	post = Post{}
	var vsb string
	if e := r.Scan(
		&post.ID, &post.Url, &post.User, &post.Date,
		&vsb, &post.Content, post.Media.ScanArray(),
		&post.Likes, &post.Shares,
	); e != nil {
		switch e {
		case sql.ErrNoRows:
			return post, ErrNotFound
		default:
			logger.Error("[Model.Posts] Cannot scan row", e)
			return post, ErrDbInternal
		}
	}
	post.Vsb, _ = utils.GetVsb(vsb)
	return post, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryPostReplies(id string) (replyings []*Post, replies []*Post, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return nil, nil, ErrDbInternal
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return nil, nil, ErrNotFound
	}

	qs := ` WITH RECURSIVE rt AS (
			    SELECT "id", "replying", 0 AS "level" FROM posts
			    WHERE "id" = $1
			  UNION ALL
			    SELECT posts."id", posts."replying", rt."level" + 1
			    FROM posts JOIN rt ON posts."id" = rt."replying"
			)
			SELECT
  			  posts."id", posts."url", posts."user", posts."date",
  			  posts."vsb", posts."content", posts."media",
			  CARDINALITY(posts."likes") as "likes", CARDINALITY(posts."shares") as "shares",
			  posts."replying", rt."level"
			FROM posts JOIN rt ON posts."id" = rt."id"
			ORDER BY "level" ASC, "date" DESC;`
	r, e := conn.Query(qs, id)
	if e != nil {
		logger.Error("[Model.Posts] Cannot query", e)
		return nil, nil, ErrDbInternal
	}

	replyings = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpy sql.NullString
		var vsb string
		if e := r.Scan(
			&p.ID, &p.Url, &p.User, &p.Date,
			&vsb, &p.Content, p.Media.ScanArray(),
			&p.Likes, &p.Shares,
			&rpy, &p.Level,
		); e != nil {
			logger.Error("[Model.Posts] Cannot scan row", e)
			continue
		}
		if rpy.Valid {
			p.Replying = rpy.String
		}
		p.Vsb, _ = utils.GetVsb(vsb)
		replyings = append(replyings, &p)
	}

	qs = `  WITH RECURSIVE rs AS (
			    SELECT "id", "replying", 0 AS "level" FROM posts
			    WHERE "id" = $1
			  UNION ALL
			    SELECT posts."id", posts."replying", rs."level" + 1
			    FROM posts JOIN rs ON rs."id" = posts."replying"
			)
			SELECT
  			  posts."id", posts."url", posts."user", posts."date",
  			  posts."vsb", posts."content", posts."media",
			  CARDINALITY(posts."likes") as "likes", CARDINALITY(posts."shares") as "shares",
			  posts."replying", rs."level"
			FROM posts JOIN rs ON posts."id" = rs."id"
			ORDER BY "level" ASC, "date" DESC;`
	r, e = conn.Query(qs, id)
	if e != nil {
		logger.Error("[Model.Posts] Cannot query", e)
		return nil, nil, ErrDbInternal
	}
	replies = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpy sql.NullString
		var vsb string
		if e := r.Scan(
			&p.ID, &p.Url, &p.User, &p.Date,
			&vsb, &p.Content, p.Media.ScanArray(),
			&p.Likes, &p.Shares,
			&rpy, &p.Level,
		); e != nil {
			logger.Error("[Model.Posts] Cannot scan row", e)
			continue
		}
		if rpy.Valid {
			p.Replying = rpy.String
		}
		p.Vsb, _ = utils.GetVsb(vsb)
		replies = append(replies, &p)
	}

	return replyings, replies, nil
}

// ERRORS
//
//   - DbInternal
func (db *PostDb) QueryPostsAndSharesByUser(user string, asc bool) (list []*Post, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return nil, ErrDbInternal
	}
	defer conn.Close()

	qs := `   WITH rr AS (
			    SELECT p1."id", p2."user" as "user"
			    FROM posts AS p1, posts AS p2
			    WHERE p1."user" = $1 AND p2."id" = p1."replying"
			  )
			  SELECT
    		    posts."id", posts."url", posts."user", posts."date",
    		    posts."vsb", posts."content", posts."media",
			    CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares",
			    rr."user" AS "replyTo", NULL AS "sharedBy", posts."date" AS "act"
			  FROM posts, rr
			  WHERE posts."user" = $1 AND posts."id" = rr."id"
			UNION ALL
			  SELECT
    		    "id", "url", "user", "date", "vsb", "content", "media",
			    CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares",
			    NULL AS "replyTo", NULL AS "sharedBy", "date" AS "act"
			  FROM posts
			  WHERE "user" = $1 AND "replying" IS NULL
			UNION ALL
			  SELECT
  			    posts."id", posts."url", posts."user", posts."date",
  			    shares."vsb", posts."content", posts."media",
			    CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares",
			    NULL AS "replyTo", shares."user" as "sharedBy", shares."date" AS "act"
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
		logger.Error("[Model.Posts] Cannot query", e)
		return nil, ErrDbInternal
	}
	list = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpt sql.NullString
		var shb sql.NullString
		var vsb string
		if e := r.Scan(
			&p.ID, &p.Url, &p.User, &p.Date,
			&vsb, &p.Content, p.Media.ScanArray(),
			&p.Likes, &p.Shares,
			&rpt, &shb, &p.ActDate,
		); e != nil {
			logger.Error("[Model.Posts] Cannot scan row", e)
			continue
		}
		if rpt.Valid {
			p.ReplyTo = rpt.String
		}
		if shb.Valid {
			p.SharedBy = shb.String
		}
		p.Vsb, _ = utils.GetVsb(vsb)
		list = append(list, &p)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - Dunplicate "post"
func (db *PostDb) SetPost(p *Post, attachments []Img) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()
	if p.Replying != "" && !db.IsPostExist(p.Replying) {
		return ErrNotFound
	}

	qs := ` INSERT INTO posts("id", "url", "user", "date", "replying", "vsb", "content", "media")
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	p.Date = p.Date.UTC()
	r, e := conn.Exec(qs,
		p.ID, p.Url, p.User, p.Date,
		p.Replying, p.Vsb.String(), p.Content,
		NewArray(attachments).ValueArray(),
	)
	if e != nil {
		logger.Error("[Model.Posts] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) UpdatePost(p *Post, attachments []Img) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` UPDATE posts
			SET "date" = $2, "content" = $3, "media" = $4
			WHERE "id" = $1;`
	p.Date = p.Date.UTC()
	r, e := conn.Exec(qs, p.ID, p.Date, p.Content, NewArray(attachments).ValueArray())
	if e != nil {
		logger.Error("[Model.Posts] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) RemovePost(id string) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Posts] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	qs := ` DELETE FROM posts
			WHERE "id" = $1;`
	r, e := conn.Exec(qs, id)
	if e != nil {
		logger.Error("[Model.Posts] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}
