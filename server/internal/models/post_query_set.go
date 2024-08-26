package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type IPostQuery interface {
	IsPostExist(id string) bool
	QueryPost(id string) (post Post, err error)
	QueryPostReplies(id string) (replyings []Post, replies []Post, err error)
	QueryUserContent(user string, maxDate time.Time) (list []Post, err error)
}

type IPostSet interface {
	SetPost(p *Post) error
	UpdatePost(p *Post) error
	RemovePost(id string) error
}

func (db *PostDb) IsPostExist(id string) bool {
	logger := db.lg
	const qs = `SELECT 1 FROM posts WHERE "id" = $1;`
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
	cacheKey := "post:" + id
	ss, err := db.cache.QueryString([]string{cacheKey})
	if err == nil {
		if ss[cacheKey] != "" {
			// cache hit
			err = json.Unmarshal([]byte(ss[cacheKey]), &post)
			if err == nil {
				// use cache
				return post, nil
			} else {
				// cache corrupted. clear cache
				db.cache.RemoveString("post:" + id)
			}
		}
	} else {
		logger.Warning("[Models.QueryPost] Failed to query cache.", "error", err)
	}

	const qs = `SELECT
				  "id", "url", "user", "date",
				  "replying", "vsb", "content", "media",
				  CARDINALITY("likes") as "likes",
				  CARDINALITY("shares") as "shares",
				  ARRAY(
				    SELECT "id" FROM posts
				    WHERE "replying" = $1
				  ) AS "replies"
				FROM posts
				WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)

	post = Post{}
	var vsb string
	var rpy sql.NullString
	if e := r.Scan(
		&post.ID, &post.Url, &post.User, &post.Date,
		&rpy, &vsb, &post.Content, pq.Array(&post.Media),
		&post.Likes, &post.Shares, pq.Array(&post.Replies),
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
		qs := `SELECT "user" FROM posts WHERE "id" = $1;`
		r := db.client.QueryRow(qs, rpy.String)
		if err := r.Scan(&post.ReplyTo); err == nil {
			post.Replying = rpy.String
		}
	}

	// cache
	s, _ := json.Marshal(post)
	go db.cache.SetString(cacheKey, string(s), false)

	return post, nil
}

func (db *PostDb) CachePosts(posts []Post) error {
	_, err := db.cache.client.TxPipelined(defaultCtx, func(p redis.Pipeliner) error {
		for _, v := range posts {
			b, _ := json.Marshal(Post{
				ID: v.ID, Url: v.Url, User: v.User, Date: v.Date,
				Replying: v.Replying, Replies: v.Replies,
				Vsb: v.Vsb, Content: v.Content, Media: v.Media,
				Likes: v.Likes, Shares: v.Shares,
			})
			p.Set(defaultCtx, "post:"+v.ID, string(b), cacheExpires)
		}
		return nil
	})
	return err
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryPostReplies(id string) (replyings []Post, replies []Post, err error) {
	logger := db.lg
	if !db.IsPostExist(id) {
		return nil, nil, ErrNotFound
	}

	const qa = `WITH RECURSIVE rt AS (
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
	r, e := db.client.Query(qa, id)
	if e != nil {
		logger.Error("[Model.Posts] Cannot query", e)
		return nil, nil, ErrDbInternal
	}

	replyings = make([]Post, 0)
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
		replyings = append(replyings, p)
	}

	const qb = `WITH RECURSIVE rs AS (
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
	r, e = db.client.Query(qb, id)
	if e != nil {
		logger.Error("[Model.Posts] Cannot query", e)
		return nil, nil, ErrDbInternal
	}
	replies = make([]Post, 0)
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
		replies = append(replies, p)
	}

	return replyings, replies, nil
}

func (db *PostDb) parseQureiedUserContent(r *sql.Rows, e error) (list []Post, err error) {
	logger := db.lg
	if e != nil {
		logger.Error("[Model.Posts] Cannot query", e)
		return nil, ErrDbInternal
	}
	list = make([]Post, 0)
	for r.Next() {
		p := Post{}
		var vsb string
		var rpy sql.NullString
		if e := r.Scan(
			&p.ID, &p.Url, &p.User, &p.Date, &rpy,
			&vsb, &p.Content, pq.Array(&p.Media),
			&p.Likes, &p.Shares,
			&p.ReplyTo, &p.SharedBy, &p.ActDate, pq.Array(&p.Replies),
		); e != nil {
			logger.Error("[Model.Posts] Cannot scan row", e)
			continue
		}
		p.Vsb, _ = utils.GetVsb(vsb)
		if rpy.Valid {
			p.Replying = rpy.String
		}
		list = append(list, p)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
func (db *PostDb) QueryUserContent(user string, maxDate time.Time) (list []Post, err error) {
	logger := db.lg
	cacheKey := "ugc:" + user
	maxDate = maxDate.UTC()
	const qs = `WITH pp AS (
				    WITH rr AS (
				      SELECT p1."id", p2."user" as "user"
				      FROM posts AS p1, posts AS p2
				      WHERE p1."user" = $1 AND p2."id" = p1."replying"
				    )
				    SELECT
				      posts."id", posts."url", posts."user", posts."date",
				      posts."replying", posts."vsb", posts."content", posts."media",
				      CARDINALITY(posts."likes") as "likes", CARDINALITY(posts."shares") as "shares",
				      rr."user" AS "replyTo", NULL AS "sharedBy", posts."date" AS "act"
				    FROM posts, rr
				    WHERE posts."user" = $1 AND posts."id" = rr."id"
				  UNION ALL
				    SELECT
				      "id", "url", "user", "date", "replying", "vsb", "content", "media",
				      CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares",
				      NULL AS "replyTo", NULL AS "sharedBy", "date" AS "act"
				    FROM posts
				    WHERE "user" = $1 AND "replying" IS NULL
				  UNION ALL
				    SELECT
				      posts."id", posts."url", posts."user", posts."date",
				      NULL AS "replying", shares."vsb", posts."content", posts."media",
				      CARDINALITY(posts."likes") as "likes", CARDINALITY(posts."shares") as "shares",
				      NULL AS "replyTo", shares."user" as "sharedBy", shares."date" AS "act"
				    FROM posts, shares
				    WHERE shares."user" = $1 AND posts."id" = shares."id"
				), pr AS (
				  SELECT pp."id", ARRAY_AGG(posts."id") AS "replies"
				  FROM posts, pp
				  WHERE posts."replying" = pp."id"
				  GROUP BY pp."id"
				)
				SELECT pp.*, pr."replies"
				FROM pp LEFT JOIN pr ON pp."id" = pr."id"
				%s;` // condition and order and limit

	// cache
	cacheResult := make(chan []PageItem)
	queryResult := make(chan []Post)
	go func(key string, maxDate time.Time, res chan []PageItem, query chan []Post) {
		toAppend := func(l []Post) [][]PageItem {
			ll := int(math.Ceil(float64(len(l)) / cachePageSize))
			pages := make([][]PageItem, 0, ll)
			for i := 0; i < ll; i += 1 {
				upper := (i + 1) * cachePageSize
				if upper > len(l) {
					upper = len(l)
				}
				pl := l[i*cachePageSize : upper]
				pp := make([]PageItem, 0, len(pl))
				for _, v := range pl {
					id := v.ID
					if v.SharedBy.Username != "" {
						id = "share:" + id
					} else {
						id = "post:" + id
					}
					pp = append(pp, PageItem{ID: id, Date: v.ActDate})
				}
				pages = append(pages, pp)
			}
			return pages
		}

		var pagesToAppend [][]PageItem
		if err := db.cache.client.Watch(defaultCtx, func(tx *redis.Tx) error {
			l, err := db.cache.QueryPageByDate(cacheKey, maxDate, false)

			if err == nil || err == ErrNoEnoughPages {
				last := l[len(l)-1]
				if last.ID == ":last" || last.ID == ":end" {
					res <- l[:len(l)-1]
				} else {
					res <- l
				}
			} else {
				logger.Debug("cache result: not found")
				res <- nil
			}

			if err == nil {
				logger.Debug("cache result: ok")
				return nil
			}
			if err == ErrNoEnoughPages {
				last := l[len(l)-1]
				if last.ID == ":last" {
					qs := fmt.Sprintf(qs, `WHERE pp."act" < $2 ORDER BY "act" DESC LIMIT $3`)
					ps, err := db.parseQureiedUserContent(db.client.Query(qs, user, last.Date, 3*cachePageSize))
					if err != nil {
						// handle error
						logger.Error("[Models.QueryUserContent.Cache] When query user content", err)
						return nil
					}

					pagesToAppend = toAppend(ps)
					err = db.cache.PushPages(tx, key, pagesToAppend, false)
					if err != nil {
						return err
					}
					logger.Debug("cache result: no enough pages")
					return db.CachePosts(ps)
				}
				logger.Debug("cache result: end")
				return nil
			}

			var ps []Post
			select {
			case ps = <-query:
				if len(ps) == 0 {
					return nil
				}
			case <-time.After(10 * time.Second):
				return nil
			}
			pagesToAppend = toAppend(ps)

			if len(l) == 1 && l[0].ID == ":last" {
				qs := fmt.Sprintf(qs, `WHERE pp."act" < $2 AND pp."act" >= $3 ORDER BY "act" DESC`)
				nps, err := db.parseQureiedUserContent(db.client.Query(qs, user, l[0].Date, maxDate))
				if err != nil {
					// handle error
					logger.Error("[Models.QueryUserContent.Cache] When query user content", err)
					return nil
				}
				ps = append(nps, ps...)
				pagesToAppend = append(toAppend(nps), pagesToAppend...)
			} else {
				if _, err := tx.TxPipelined(defaultCtx, func(p redis.Pipeliner) error {
					return p.LTrim(defaultCtx, key, 1, 0).Err()
				}); err != nil {
					return err
				}
			}
			if err = db.cache.PushPages(tx, key, pagesToAppend, false); err != nil {
				logger.Debug("not found then query and cache")
				return err
			}
			return db.CachePosts(ps)
		}, key); err == redis.TxFailedErr {
			// discard cache
		} else if err != nil {
			// handle error
			logger.Error("[Models.QueryUserContent.Cache] When query user content", err)
		}
	}(cacheKey, maxDate, cacheResult, queryResult)

	select {
	case cache := <-cacheResult:
		if len(cache) != 0 {
			// cache hit
			idx := make([]struct {
				Type string
				ID   string
				Date time.Time
			}, 0, len(cache))
			ids := make([]string, 0, len(cache))
			tmp := make([][]string, 0)
			reg := regexp.MustCompile(`(post|share):([0-9a-z]+)`)
			for _, v := range cache {
				// don't know why this won't work
				// matches := db.pageIDReg.FindStringSubmatch(v.ID)
				matches := reg.FindStringSubmatch(v.ID)
				tmp = append(tmp, matches)
				if len(matches) < 3 {
					continue
				}
				idx = append(idx, struct {
					Type string
					ID   string
					Date time.Time
				}{matches[1], matches[2], v.Date})
				ids = append(ids, "post:"+matches[2])
			}

			// try to get posts from cache
			cm, err := db.cache.QueryString(ids)
			m := make(map[string]*Post)
			if err == nil {
				tmp := make([]string, 0, len(cache)-len(cm))
				for i, v := range ids {
					if cm[v] == "" {
						matches := reg.FindStringSubmatch(v)
						tmp = append(tmp, matches[2])
						continue
					}
					var p Post
					if err := json.Unmarshal([]byte(cm[v]), &p); err != nil {
						tmp = append(tmp, v)
						continue
					}
					m[idx[i].ID] = &p
				}
				ids = tmp
			}

			if len(ids) != 0 {
				// get uncached posts
				nqs := `WITH pp AS (
					    SELECT
					      p1."id", p1."url", p1."user", p1."date", p1."vsb", p1."content", p1."media",
					      p2."user" AS "replyTo", p1."replying",
					      CARDINALITY(p1."likes") as "likes", CARDINALITY(p1."shares") as "shares"
					    FROM posts AS p1, posts AS p2
					    WHERE p1."id" IN $1 AND p2."id" = p1."replying"
					  UNION ALL
					    SELECT
					      "id", "url", "user", "date", "vsb", "content", "media", NULL AS "replyTo", "replying",
					      CARDINALITY("likes") as "likes", CARDINALITY("shares") as "shares"
					    FROM posts
					    WHERE "id" IN $1 AND "replying" IS NULL
					), pr AS (
					  SELECT pp."id", ARRAY_AGG(posts."id") AS "replies"
					  FROM posts, pp
					  WHERE posts."replying" = pp."id"
					  GROUP BY pp."id"
					)
					SELECT pp.*, pr."replies"
					FROM pp LEFT JOIN pr ON pp."id" = pr."id"; `
				l, err := db.parseQureiedUserContent(db.client.Query(nqs, pq.Array(ids)))
				if err != nil {
					return nil, err
				}
				go db.CachePosts(l) // cache them
				for k := range l {
					m[l[k].ID] = &l[k]
				}
			}
			list := make([]Post, 0, len(cache))
			for _, v := range idx {
				p := m[v.ID]
				if v.Type == "share" {
					p.Replying = ""
					p.ReplyTo = UD{}
					p.SharedBy = NewUD(user)
				}
				p.ActDate = v.Date
				list = append(list, *p)
			}
			return list, nil
		}
	case <-time.After(5 * time.Second):
	}

	nqs := fmt.Sprintf(qs, `WHERE pp."act" < $2 ORDER BY "act" DESC LIMIT $3`)
	list, err = db.parseQureiedUserContent(db.client.Query(nqs, user, maxDate, 3*cachePageSize))
	queryResult <- list
	if len(list) >= cachePageSize {
		list = list[:cachePageSize]
	}
	return list, err
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

	const qs = `INSERT INTO posts("id", "url", "user", "date", "replying", "vsb", "content", "media")
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
	if p.Replying != "" {
		qs := `SELECT "user" FROM posts WHERE "id" = $1;`
		r := db.client.QueryRow(qs, p.Replying)
		if err := r.Scan(&p.ReplyTo); err != nil {
			if err == sql.ErrNoRows {
				p.Replying = ""
			} else {
				// handle error
				return nil
			}
		}
	}
	s, _ := json.Marshal(Post{
		ID: p.ID, Url: p.Url, User: p.User,
		Date: p.Date, Vsb: p.Vsb,
		Content: p.Content, Media: p.Media,
		Replying: p.Replying, ReplyTo: p.ReplyTo,
		Likes: 0, Shares: 0,
	})
	go db.cache.SetString("post:"+p.ID, string(s), false)
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound
func (db *PostDb) UpdatePost(p *Post) error {
	logger := db.lg
	const qs = `UPDATE posts
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
	go db.cache.UpdateJson("post:"+p.ID, Post{
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

	const qs = `DELETE FROM posts
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
	go db.cache.RemoveString("post:" + id)
	return nil
}
