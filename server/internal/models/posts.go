package models

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
)

// models

type Img struct {
	Url string `json:"url"`
	Alt string `json:"alt,omitempty"`
}

func (img Img) Value() (driver.Value, error) {
	if img.Alt == "" {
		return fmt.Sprintf("\"(%s,)\"", img.Url), nil
	} else {
		return fmt.Sprintf("\"(%s,%s)\"", img.Url, img.Alt), nil
	}
}

func (img *Img) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Scan img: src cannot cast to []byte")
	}
	fields := strings.Split(strings.Trim(string(b), "()"), ",")
	if fields[0] == "" {
		return fmt.Errorf("Scan img: empty url")
	}
	img.Url = fields[0]
	if len(fields) > 1 {
		img.Alt = fields[1]
	}
	return nil
}

type Post struct {
	ID       string           `json:"id"`
	Url      string           `json:"url"`
	User     string           `json:"user"`
	Date     time.Time        `json:"date"`
	Vsb      utils.Vsb        `json:"vsb"`
	Content  string           `json:"content"`
	Media    Array[Img, *Img] `json:"media"`    // magic but sucks
	Replying string           `json:"replying"` // post id
	ReplyTo  string           `json:"replyTo"`  // user id, temporary field
	SharedBy string           `json:"sharedBy"` // user id, temporary field
	Likes    int64            `json:"likes"`    // count, temporary field
	Shares   int64            `json:"shares"`   // count, temporary field
	ActDate  string           `json:"actDate"`  // temporary field, used in sort
	Level    int              `json:"level"`    // temporary field, used in replying and replies
}

// db

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

type IPostLike interface {
	QueryLikes(id string) (list []string, owner string, vsb utils.Vsb, err error)
	SetLike(user, id string) error
	RemoveLike(user, id string) error
}

type IPostShare interface {
	QueryShares(id string) (list []string, owner string, vsb utils.Vsb, err error)
	SetShare(user, id string, date time.Time, vsb utils.Vsb) error
	RemoveShare(user, id string) error
}

type PostDb struct {
	lg   logging.Logger
	pool *_db.ConnPool[*_db.PqConn]
}

var postIns *PostDb = nil

func PostInstance(lg logging.Logger) *PostDb {
	if postIns == nil {
		postIns = &PostDb{
			lg:   lg,
			pool: _db.MainPool(nil, nil),
		}
	}
	return postIns
}

// functions

func (db *PostDb) IsPostExist(id string) bool {
	logger := db.lg
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
			logger.Error("[Model.Posts] Cannot query", e)
			return false
		}
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
			  "id", "url", "user", "date",
			  "vsb", "content", "media",
			  CARDINALITY("likes") as "likes",
			  CARDINALITY("shares") as "shares"
			FROM posts
			WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)

	post = Post{}
	var vsb string
	if e := r.Scan(
		&post.ID, &post.Url, &post.User, &post.Date,
		&vsb, &post.Content, post.Media.ToPqArray(),
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
		logger.Error("[Model.Reply] Failed to open a connection", err)
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
	`
	r, e := conn.Query(qs, id)
	if e != nil {
		logger.Error("[Model.Reply] Cannot query", e)
		return nil, nil, ErrDbInternal
	}

	replyings = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpy sql.NullString
		var vsb string
		if e := r.Scan(
			&p.ID, &p.Url, &p.User, &p.Date,
			&vsb, &p.Content, p.Media.ToPqArray(),
			&p.Likes, &p.Shares,
			&rpy, &p.Level,
		); e != nil {
			logger.Error("[Model.Reply] Cannot scan row", e)
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
	`
	r, e = conn.Query(qs, id)
	if e != nil {
		logger.Error("[Model.Reply] Cannot query", e)
		return nil, nil, ErrDbInternal
	}
	replies = make([]*Post, 0)
	for r.Next() {
		p := Post{}
		var rpy sql.NullString
		var vsb string
		if e := r.Scan(
			&p.ID, &p.Url, &p.User, &p.Date,
			&vsb, &p.Content, p.Media.ToPqArray(),
			&p.Likes, &p.Shares,
			&rpy, &p.Level,
		); e != nil {
			logger.Error("[Model.Reply] Cannot scan row", e)
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
			    CARDINALITY("likes") as "likes",
			    CARDINALITY("shares") as "shares",
			    rr."user" AS "replyTo", NULL AS "sharedBy",
			    posts."date" AS "act"
			  FROM posts, rr
			  WHERE posts."user" = $1 AND posts."id" = rr."id"
			UNION ALL
			  SELECT
    		    "id", "url", "user", "date",
    		    "vsb", "content", "media",
			    CARDINALITY("likes") as "likes",
			    CARDINALITY("shares") as "shares",
			    NULL AS "replyTo", NULL AS "sharedBy",
			    "date" AS "act"
			  FROM posts
			  WHERE "user" = $1 AND "replying" IS NULL
			UNION ALL
			  SELECT
  			    posts."id", posts."url", posts."user", posts."date",
  			    shares."vsb", posts."content", posts."media",
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
		logger.Error("[Model.Reply] Cannot query", e)
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
			&vsb, &p.Content, p.Media.ToPqArray(),
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
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()
	if p.Replying != "" && !db.IsPostExist(p.Replying) {
		return ErrNotFound
	}

	qs := ` INSERT INTO posts(
  			  "id", "url", "user", "date",
  			  "replying", "vsb", "content",
			  "media"
			)
			VALUES (
			  $1, $2, $3, $4,
			  $5, $6, $7,
			  $8
			);`
	p.Date = p.Date.UTC()
	r, e := conn.Exec(qs,
		p.ID, p.Url, p.User, p.Date,
		p.Replying, p.Vsb.String(), p.Content,
		NewArray(attachments, logger),
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
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` UPDATE posts
			SET
			  "date" = $2, "content" = $3,
			  "media" = $4
			WHERE "id" = $1;`
	p.Date = p.Date.UTC()
	r, e := conn.Exec(qs, p.ID, p.Date, p.Content, NewArray(attachments, logger))
	if e != nil {
		logger.Error("[Model.Reply] Failed to execute", e)
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
		logger.Error("[Model] Failed to open a connection", err)
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
		logger.Error("[Model.Reply] Failed to execute", e)
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
func (db *PostDb) QueryLikes(id string) (list []string, owner string, vsb utils.Vsb, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, "", vsb, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT "user", "vsb", "likes"
			FROM posts
			WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)

	u, v := "", ""
	e := r.Scan(&u, &v, pq.Array(&list))
	vsb, _ = utils.GetVsb(v)
	if e != nil {
		switch e {
		case sql.ErrNoRows:
			return nil, u, vsb, ErrNotFound
		default:
			logger.Error("[Model.Like] Cannot scan row", e)
			return nil, "", vsb, ErrDbInternal
		}
	}
	return list, u, vsb, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
//   - Dunplicate "like"
func (db *PostDb) SetLike(user, id string) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Like] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	qs := ` UPDATE posts
			SET "likes" = ARRAY_APPEND("likes", $1)
			WHERE
			  "id" = $2
  			  AND ARRAY_POSITION("likes", $1) IS NULL;`
	r, e := conn.Exec(qs, user, id)
	if e != nil {
		logger.Error("[Model.Reply] Failed to execute", e)
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
func (db *PostDb) RemoveLike(user, id string) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Like] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` UPDATE posts
			SET "likes" = ARRAY_REMOVE("likes", $1)
			WHERE "id" = $2;`
	r, e := conn.Exec(qs, user, id)
	if e != nil {
		logger.Error("[Model.Reply] Failed to execute", e)
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
func (db *PostDb) QueryShares(id string) (list []string, owner string, vsb utils.Vsb, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Share] Failed to open a connection", err)
		return nil, "", vsb, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT "user", "vsb", "shares"
			FROM posts
			WHERE "id" = $1;`
	r := conn.QueryOne(qs, id)

	u, v := "", ""
	e := r.Scan(&u, &v, pq.Array(&list))
	vsb, _ = utils.GetVsb(v)
	if e != nil {
		switch e {
		case sql.ErrNoRows:
			return nil, u, vsb, ErrNotFound
		default:
			logger.Error("[Model.Like] Cannot scan row", e)
			return nil, "", vsb, ErrDbInternal
		}
	}
	return list, u, vsb, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
//   - Dunplicate "share"
func (db *PostDb) SetShare(user, id string, date time.Time, vsb utils.Vsb) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Share] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	tx, e := conn.BeginTx()
	if e != nil {
		logger.Error("[Model.Share] Cannot start transaction", e)
		return ErrDbInternal
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
		logger.Error("[Model.Reply] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}

	// insert into shares
	qs = `  INSERT INTO shares("user", "id", "date", "vsb")
			VALUES ($1, $2, $3, $4);`
	_, e = tx.Exec(qs, user, id, date, vsb.String())
	if e != nil {
		logger.Error("[Model.Reply] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}

	if e := tx.Commit(); e != nil {
		logger.Error("[Model.Reply] Cannot commit", e)
		return ErrDbInternal
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post", "share"
func (db *PostDb) RemoveShare(user, id string) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Share] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	tx, e := conn.BeginTx()
	if e != nil {
		logger.Error("[Model.Share] Cannot start transaction", e)
		return ErrDbInternal
	}

	// use 2 annoymous func to ensure completely deletion

	err1 := func() error {
		qs := ` UPDATE posts
				SET "shares" = ARRAY_REMOVE("shares", $1)
				WHERE "id" = $2;
		`
		r, e := tx.Exec(qs, user, id)
		if e != nil {
			logger.Error("[Model.Share] Failed to exec", e)
			return ErrDbInternal
		}
		if r == 0 {
			return ErrNotFound
		}
		return nil
	}()

	err2 := func() error {
		qs := ` DELETE FROM shares
				WHERE "user" = $1 and "id" = $2;`
		r, e := tx.Exec(qs, user, id)
		if e != nil {
			logger.Error("[Model.Share] Failed to exec", e)
			return ErrDbInternal
		}
		if r == 0 {
			return ErrNotFound
		}
		return nil
	}()

	if e := tx.Commit(); e != nil {
		logger.Error("[Model.Reply] Cannot commit", e)
		return ErrDbInternal
	}

	if err1 != nil || err2 != nil {
		return ErrNotFound
	} else {
		return nil
	}
}
