package models

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
)

type IPostQuery interface {
	IsPostExist(id string) bool
	QueryPost(id string) (post Post, err error)
	QueryPostReplies(id string) (replyings []*Post, replies []*Post, err error)
	QueryPostsAndSharesByUser(user string, asec bool) (list []*Post, err error)
}

type IPostSet interface {
	SetPost(p *Post) error
	UpdatePost(p *Post) error
	RemovePost(id string) error
}

func (db *PostDb) IsPostExist(id string) bool {
	logger := db.lg
	qs := `SELECT 1 FROM posts WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)
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
func (db *PostDb) QueryPost(id string) (post Post, err error) {
	logger := db.lg
	if id == "" {
		return post, ErrFormat
	}

	// try query cache
	ss, err := db.cache.CacheQueryString([]string{"post:" + id})
	if err == nil && ss["post:"+id] != "" {
		err = json.Unmarshal([]byte(ss["post:"+id]), &post)
		if err == nil {
			// cache hit
			return post, nil
		}
		// cache corrupted. clear cache
		db.cache.CacheRemoveString("post:" + id)
	} else if err != nil && err != ErrNotFound {
		// don't return
		logger.Warning("[Models.QueryPost] Failed to query cache.", "error", err)
	}

	qs := ` SELECT
			  "id", "url", "user", "date", "replying", "vsb", "content", "media",
			  CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares"
			FROM posts
			WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)

	post = Post{}
	var vsb string
	var rpy sql.NullString
	if e := r.Scan(
		&post.ID, &post.Url, &post.User, &post.Date,
		&rpy, &vsb, &post.Content, pq.Array(&post.Media),
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
	if rpy.Valid {
		qs = `SELECT "user" FROM posts WHERE "id" = $1;`
		r := db.client.QueryRow(qs, rpy.String)
		var u UD
		if err := r.Scan(&u); err == nil {
			post.Replying = rpy.String
			post.ReplyTo = u.String()
		}
	}

	// cache
	s, _ := json.Marshal(post)
	go db.cache.CacheSetString("post:"+post.ID, string(s), false)

	return post, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryPostReplies(id string) (replyings []*Post, replies []*Post, err error) {
	logger := db.lg
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
	r, e := db.client.Query(qs, id)
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
			&vsb, &p.Content, pq.Array(&p.Media),
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
	r, e = db.client.Query(qs, id)
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
			&vsb, &p.Content, pq.Array(&p.Media),
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
	r, e := db.client.Query(qs, user)
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
			&vsb, &p.Content, pq.Array(&p.Media),
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
	// !!cache
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - Dunplicate
func (db *PostDb) SetPost(p *Post) error {
	logger := db.lg
	if p.Replying != "" && !db.IsPostExist(p.Replying) {
		return ErrNotFound
	}

	qs := ` INSERT INTO posts("id", "url", "user", "date", "replying", "vsb", "content", "media")
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`
	p.Date = p.Date.UTC()
	var x *string
	if p.Replying != "" {
		x = &p.Replying
	}
	r, e := sqlExec(db.client.Exec(qs,
		p.ID, p.Url, p.User, p.Date,
		x, p.Vsb.String(), p.Content,
		pq.Array(p.Media),
	))
	if e != nil {
		logger.Error("[Model.Posts] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}

	// cache
	p.ReplyTo = ""
	if p.Replying != "" {
		qs = `SELECT "user" FROM posts WHERE "id" = $1;`
		r := db.client.QueryRow(qs, p.Replying)
		var u UD
		if err := r.Scan(&u); err == nil {
			p.ReplyTo = u.String()
		} else {
			p.Replying = ""
		}
	}
	s, _ := json.Marshal(Post{
		ID: p.ID, Url: p.Url, User: p.User,
		Date: p.Date, Vsb: p.Vsb,
		Content: p.Content, Media: p.Media,
		Replying: p.Replying, ReplyTo: p.ReplyTo,
		Likes: 0, Shares: 0,
	})
	go db.cache.CacheSetString("post:"+p.ID, string(s), false)
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound
func (db *PostDb) UpdatePost(p *Post) error {
	logger := db.lg
	qs := ` UPDATE posts
			SET "date" = $2, "content" = $3, "media" = $4
			WHERE "id" = $1;`
	p.Date = p.Date.UTC()
	r, e := sqlExec(db.client.Exec(qs, p.ID, p.Date, p.Content, pq.Array(p.Media)))
	if e != nil {
		logger.Error("[Model.Posts] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}

	// cache
	go db.cache.CacheUpdateJson("post:"+p.ID, Post{
		Date: p.Date, Content: p.Content, Media: p.Media,
	})
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound
func (db *PostDb) RemovePost(id string) error {
	logger := db.lg
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	qs := ` DELETE FROM posts
			WHERE "id" = $1;`
	r, e := sqlExec(db.client.Exec(qs, id))
	if e != nil {
		logger.Error("[Model.Posts] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}

	// cache
	go db.cache.CacheRemoveString("post:" + id)
	return nil
}
