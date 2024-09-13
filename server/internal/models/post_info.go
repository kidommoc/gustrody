package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type IPostQuery interface {
	IsPostExist(id string) bool
	QueryPosts(ids []string) (posts map[string]*Post, err error)
	QueryPostReplyChain(id string) (replyings []ReplyData, replies []ReplyData, err error)
	QueryUserContent(user string, maxDate time.Time, vsb utils.Vsb) (list []Post, err error)
}

type IPostSet interface {
	SetPost(p Post) error
	// uses: id, content, media
	UpdatePost(p Post) error
	RemovePost(id string) error
}

func isPostExist(client ISqlClient, id string) (bool, error) {
	if client == nil {
		return false, ErrFormat
	}
	const qs = `SELECT 1 FROM "posts" WHERE "id" = $1 AND "tombstone" IS FALSE;`
	r, err := client.Query(qs, id)
	if err != nil {
		return false, err
	}
	defer r.Close()
	for r.Next() {
		return true, nil
	}
	return false, nil
}

func (db *PostDb) IsPostExist(id string) bool {
	const loc = "Models.Post.Exist"
	r, err := isPostExist(db.client, id)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query %s", id), err)
		return false
	}
	return r
}

func cachePosts(cache *CacheDb, posts []Post) error {
	if cache == nil {
		return fmt.Errorf("[models.posts.cache] cacheDb is nil")
	}
	for _, v := range posts {
		v.Type = postJsonType
		b, err := json.Marshal(v)
		if err != nil {
			// handle error?
			continue
		}
		err = cache.SetString("post:"+v.ID, string(b), false)
		// handle error?
	}
	return nil
}

func (db *PostDb) QueryPosts(ids []string) (posts map[string]*Post, err error) {
	const loc = "Models.Post.Query"
	if len(ids) == 0 {
		return nil, nil
	}
	posts = make(map[string]*Post)

	// try query cache
	cacheKeys := make([]string, 0, len(ids))
	for _, v := range ids {
		cacheKeys = append(cacheKeys, "post:"+v)
	}
	cm, err := db.cache.QueryString(cacheKeys)
	if err == nil {
		tmp := make([]string, 0, len(ids))
		for _, v := range ids {
			key := "post:" + v
			if cm[key] == "" {
				tmp = append(tmp, v)
				continue
			}
			var p Post
			if err := json.Unmarshal([]byte(cm[v]), &p); err != nil {
				tmp = append(tmp, v)
				continue
			}
			p.Type = ""
			posts[v] = &p
		}
		ids = tmp
	} else {
		db.lg.Warning(fmt.Sprintf("[%s] Failed to query cache.", loc), "error", err, "ids", ids)
	}

	if len(ids) != 0 {
		const qs = `SELECT
					  "id", "tombstone", "fid", "user", "date", "vsb", "content", "media",
					  "replying", "reply_to", "replies", "shares", "likes"
					FROM "post_data"
					WHERE "post_data"."id" = ANY($1);`
		r, err := db.client.Query(qs, pq.Array(ids))
		if err != nil {
			logging.Cannot(db.lg, loc, fmt.Sprintf("query posts: %v", ids), err)
			return nil, ErrDbInternal
		}
		defer r.Close()

		ps := make([]Post, 0, len(ids)-len(posts)) // posts to cache
		for r.Next() {
			var p Post
			var vsb string
			var rply sql.NullString
			var rplt sql.NullString
			if err := r.Scan(
				&p.ID, &p.Tombstone, &p.FederalID, &p.User, &p.Date,
				&vsb, &p.Content, pq.Array(&p.Media),
				&rply, &rplt, &p.Replies, &p.Shares, &p.Likes,
			); err != nil || p.Tombstone {
				db.lg.Warning("err when scan", "err", err)
				continue
			}
			p.Vsb, _ = utils.GetVsb(vsb)
			if rply.Valid && rplt.Valid {
				p.Replying = rply.String
				p.ReplyTo = NewUD(rplt.String)
			}
			posts[p.ID] = &p
			ps = append(ps, p)
		}

		// cache
		go cachePosts(db.cache, ps)
	}

	return posts, nil
}

type ReplyData struct {
	ID string
	To string
}

func (db *PostDb) QueryPostReplyChain(id string) (replyings []ReplyData, replies []ReplyData, err error) {
	const loc = "Models.Post.QueryReplyChain"
	if !db.IsPostExist(id) {
		return nil, nil, ErrNotFound
	}

	const qa = `WITH RECURSIVE "rs" AS (
				    SELECT "id", "tgt" FROM "reply" WHERE "id" = $1
				  UNION ALL
				    SELECT "reply"."id", "reply"."tgt" FROM "reply"
				    JOIN "rs" ON "rs"."tgt" = "reply"."id"
				)
				SELECT "id", "tgt" FROM "rs";`
	r, err := db.client.Query(qa, id)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query replying of %s", id), err)
		return nil, nil, ErrDbInternal
	}
	defer r.Close()

	replyings = make([]ReplyData, 0)
	for r.Next() {
		var d ReplyData
		if e := r.Scan(&d.ID, &d.To); e != nil {
			continue
		}
		replyings = append(replyings, d)
	}

	const qb = `WITH RECURSIVE "rs" AS (
				    SELECT "id", "tgt" FROM "reply" WHERE "tgt" = $1
				  UNION ALL
				    SELECT "reply"."id", "reply"."tgt" FROM "reply"
				    JOIN "rs" ON "rs"."id" = "reply"."tgt"
				)
				SELECT "id", "tgt" FROM "rs";`
	r, err = db.client.Query(qb, id)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query replies of %s", id), err)
		return nil, nil, ErrDbInternal
	}
	defer r.Close()

	replies = make([]ReplyData, 0)
	for r.Next() {
		var d ReplyData
		if e := r.Scan(&d.ID, &d.To); e != nil {
			continue
		}
		replies = append(replies, d)
	}

	return replyings, replies, nil
}

type ugcItem struct {
	ID       string
	Date     time.Time
	SharedBy UD
}

func (db *PostDb) QueryUserContent(user string, maxDate time.Time, vsb utils.Vsb) (list []Post, err error) {
	const loc = "Models.User.QueryUGC"
	logger := db.lg
	cacheKey := "ugc:" + user
	maxDate = maxDate.UTC()
	if vsb == utils.Vsb_DIRECT {
		vsb = utils.Vsb_FOLLOWER
	}

	genQuery := func(where string, limit string) string {
		const qs = `SELECT "id", "date", "shared_by" FROM "user_content"
					WHERE "user" = $1 AND "vsb" <= $2 %s %s;`
		if where != "" {
			where = `AND ` + where
		}
		if limit != "" {
			limit = "LIMIT " + limit
		}
		return fmt.Sprintf(qs, where, limit)
	}

	parseUgc := func(r *sql.Rows, e error) (list []ugcItem, err error) {
		if e != nil {
			return nil, e
		}
		defer r.Close()
		list = make([]ugcItem, 0)
		for r.Next() {
			var p ugcItem
			if e := r.Scan(&p.ID, &p.Date, &p.SharedBy); e != nil {
				continue
			}
			list = append(list, p)
		}
		return list, nil
	}

	toAppend := func(l []ugcItem) [][]PageItem {
		ll := int(math.Ceil(float64(len(l)) / cachePageSize))
		pages := make([][]PageItem, 0, ll)
		for i := 0; i < ll; i++ {
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
				pp = append(pp, PageItem{ID: id, Date: v.Date})
			}
			pages = append(pages, pp)
		}
		return pages
	}

	// cache
	cacheResult := make(chan []PageItem)
	queryResult := make(chan []ugcItem)
	t0 := time.Now()
	go func(key string, maxDate time.Time, res chan []PageItem, query chan []ugcItem) {
		loc := fmt.Sprintf("%s.Cache", loc)

		if err := db.cache.client.Watch(defaultCtx, func(tx *redis.Tx) error {
			var pagesToAppend [][]PageItem
			l, err := db.cache.QueryPageByDate(key, maxDate, false)

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
					qs := genQuery(`"date" < $3`, `$4`)
					ps, err := parseUgc(db.client.Query(qs, user, vsb, last.Date, 3*cachePageSize))
					if err != nil {
						logging.Cannot(db.lg, loc, fmt.Sprintf("query ugc of %s for cache", user), err)
						return nil
					}

					pagesToAppend = toAppend(ps)
					err = db.cache.PushPages(tx, key, pagesToAppend, false)
					if err != nil {
						return err
					}
					ids := make([]string, 0, len(ps))
					for _, v := range ps {
						ids = append(ids, v.ID)
					}
					db.QueryPosts(ids)
					logger.Debug("cache result: no enough pages")
					return nil
				}
				logger.Debug("cache result: end")
				return nil
			}

			var ps []ugcItem
			select {
			case ps = <-query:
				if len(ps) == 0 {
					return nil
				}
			case <-time.After(7 * time.Second):
				return nil
			}
			pagesToAppend = toAppend(ps)

			if len(l) == 1 && l[0].ID == ":last" {
				qs := genQuery(`"date" < $3 AND "date" >= $4`, ``)
				nps, err := parseUgc(db.client.Query(qs, user, vsb, l[0].Date, maxDate))
				if err != nil {
					logging.Cannot(db.lg, loc, fmt.Sprintf("query ugc of %s for cache", user), err)
					return nil
				}
				pagesToAppend = append(toAppend(nps), pagesToAppend...)
			} else {
				if _, err := tx.TxPipelined(defaultCtx, func(p redis.Pipeliner) error {
					return p.LTrim(defaultCtx, key, 1, 0).Err()
				}); err != nil {
					return err
				}
			}
			logger.Debug("not found then query and cache")
			if err = db.cache.PushPages(tx, key, pagesToAppend, false); err != nil {
				return err
			}
			return nil
		}, key); err == redis.TxFailedErr {
			// discard cache
		} else if err != nil {
			logging.FailedTo(db.lg, loc, fmt.Sprintf("handle cache when query ugc of %s", user), err)
		}
	}(cacheKey, maxDate, cacheResult, queryResult)

	select {
	case cache := <-cacheResult:
		t1 := time.Now()
		db.lg.Debug(fmt.Sprintf("1st cache query cost: %s", t1.Sub(t0)))
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
				ids = append(ids, matches[2])
			}

			t2 := time.Now()
			ps, err := db.QueryPosts(ids)
			db.lg.Debug(fmt.Sprintf("2st posts query cost: %s", time.Now().Sub(t2)))
			if err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("failed to query posts when query ugc of %s", user), err)
				return nil, err
			}
			for _, v := range idx {
				if ps[v.ID] != nil {
					if v.Type == "share" {
						ps[v.ID].SharedBy = NewUD(user)
					}
					list = append(list, *ps[v.ID])
				}
			}
			return list, nil
		}
	case <-time.After(3 * time.Second):
	}

	qs := genQuery(`"date" < $3`, `$4`)
	t1 := time.Now()
	l, err := parseUgc(db.client.Query(qs, user, vsb, maxDate, 3*cachePageSize))
	db.lg.Debug(fmt.Sprintf("2st ugc query cost: %s", time.Now().Sub(t1)))
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query ugc of %s", user), err)
		return nil, ErrDbInternal
	}
	queryResult <- l

	ids := make([]string, 0, len(l))
	for _, v := range l {
		ids = append(ids, v.ID)
	}
	t2 := time.Now()
	ps, err := db.QueryPosts(ids)
	db.lg.Debug(fmt.Sprintf("3st posts query cost: %s", time.Now().Sub(t2)))
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query posts of %s's ugc", user), err)
		return nil, err
	}

	for i, v := range l {
		if i >= cachePageSize {
			break
		}
		if ps[v.ID] != nil {
			if v.SharedBy.Username != "" {
				ps[v.ID].SharedBy = v.SharedBy
			}
			list = append(list, *ps[v.ID])
		}
	}
	return list, err
}

func (db *PostDb) SetPost(p Post) error {
	const loc = "Models.Post.Set"
	const qa = `INSERT INTO "posts"("id", "fid", "user", "date", "vsb", "content", "media")
				VALUES ($1, $2, $3, $4, $5, $6, $7);`
	const qb = `INSERT INTO "reply"("id", "tgt", "tgt_user") VALUES($1, $2, $3);`
	const qr = `SELECT "user" FROM "posts" WHERE "id" = $1;`
	p.Date = p.Date.UTC()

	err := func() error {
		for i := 0; i < txMaxRetries; i++ {
			if err := func() error {
				tx, err := db.client.BeginTx(defaultCtx, nil)
				if err != nil {
					return ErrDbInternal
				}
				defer tx.Rollback()

				// check duplicated
				pe, err := isPostExist(tx, p.ID)
				if err != nil {
					logging.Cannot(db.lg, loc, fmt.Sprintf("query post %s for duplicate checking", p.ID), err)
				}
				if pe {
					return ErrDuplicated
				}

				// check replying and get replyTo
				if p.Replying != "" {
					r := tx.QueryRow(qr, p.Replying)
					if err := r.Scan(&p.ReplyTo); err != nil {
						switch err {
						case sql.ErrNoRows:
							return ErrNotFound
						default:
							logging.Cannot(db.lg, loc, fmt.Sprintf("query user of replying post %s when set %s", p.ID, p.Replying), err)
							return ErrDbInternal
						}
					}
				}

				// set post
				if _, err := sqlExec(tx.Exec(qa,
					p.ID, p.FederalID, p.User, p.Date,
					p.Vsb.String(), p.Content, pq.Array(p.Media),
				)); err != nil {
					logging.FailedTo(db.lg, loc, fmt.Sprintf("set post %s", p.ID), err)
					return ErrDbInternal
				}

				// set reply
				if p.Replying != "" {
					if _, err := sqlExec(tx.Exec(qb, p.ID, p.Replying, p.ReplyTo)); err != nil {
						logging.FailedTo(db.lg, loc, fmt.Sprintf("set replying of %s", p.ID), err)
						return ErrDbInternal
					}
				}

				if err := tx.Commit(); err != nil {
					return fmt.Errorf("t")
				}
				return nil
			}(); err == nil || err.Error() != "t" {
				return err
			}
		}
		return ErrMaxRetries
	}()

	// cache
	go func() {
		if err == nil {
			err := cachePosts(db.cache, []Post{{
				ID: p.ID, FederalID: p.FederalID, User: p.User, Date: p.Date,
				Vsb: p.Vsb, Content: p.Content, Media: p.Media,
				Replying: p.Replying, ReplyTo: p.ReplyTo,
				Replies: 0, Shares: 0, Likes: 0,
			}})
			if err != nil {
				db.lg.Warning("[Models.Post.Cache] Failed to cache post.", "err", err, "post_id", p.ID)
			}
			if p.Replying != "" {
				db.cache.UpdateJson("post:"+p.Replying, &PostCache{atomic: true, Replies: 1})
			}
		}
	}()
	return err
}

func (db *PostDb) UpdatePost(p Post) error {
	const loc = "Models.Post.Update"
	const qs = `UPDATE "posts" SET "content" = $2, "media" = $3
				WHERE "id" = $1;`
	r, err := sqlExec(db.client.Exec(qs, p.ID, p.Content, pq.Array(p.Media)))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("update %s", p.ID), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}

	go func(p Post) {
		db.cache.UpdateJson("post:"+p.ID, &PostCache{
			atomic: false, Content: &p.Content, Media: &p.Media,
		})
	}(p)
	return nil
}

func (db *PostDb) RemovePost(id string) error {
	const loc = "Models.Post.Remove"
	if !db.IsPostExist(id) {
		return ErrNotFound
	}
	const qr = `SELECT "tgt" FROM "reply" WHERE "id" = $1;`
	const qa = `DELETE FROM "reply" WHERE "id" = $1 OR "tgt" = $1;`
	const qt = `DELETE FROM "%s" WHERE "tgt" = $1;`
	const qs = `DELETE FROM "posts" WHERE "id" = $1;`
	var replying sql.NullString

	err := func() error {
		for i := 0; i < txMaxRetries; i++ {
			if err := func() error {
				tx, err := db.client.BeginTx(defaultCtx, nil)
				if err != nil {
					return ErrDbInternal
				}
				defer tx.Rollback()

				// query reply
				re := tx.QueryRow(qr, id)
				if err := re.Scan(&replying); err != nil {
					switch err {
					case sql.ErrNoRows:
						replying.Valid = false
					default:
						logging.FailedTo(db.lg, loc, fmt.Sprintf("query replying of %s", id), err)
						return ErrDbInternal
					}
				}

				// reomve reply
				if _, err := sqlExec(tx.Exec(qa, id)); err != nil {
					logging.FailedTo(db.lg, loc, fmt.Sprintf("remove replying of %s", id), err)
					return ErrDbInternal
				}

				// remove shares, likes
				tables := []string{"like", "share"}
				for _, v := range tables {
					q := fmt.Sprintf(qt, v)
					if _, err := sqlExec(tx.Exec(q, id)); err != nil {
						logging.FailedTo(db.lg, loc, fmt.Sprintf("remove %s of %s", v, id), err)
						return ErrDbInternal
					}
				}

				// remove post
				r, err := sqlExec(tx.Exec(qs, id))
				if err != nil {
					logging.FailedTo(db.lg, loc, fmt.Sprintf("remove %s", id), err)
					return ErrDbInternal
				}
				if r == 0 {
					return ErrNotFound
				}

				if err := tx.Commit(); err != nil {
					return fmt.Errorf("t")
				}
				return nil
			}(); err == nil || err.Error() != "t" {
				return err
			}
		}
		return ErrMaxRetries
	}()

	// cache. ensure removed successfully
	if err == nil {
		go func() {
			if err := db.cache.SetString("post:"+id, `{"tombstone":true}`, false); err != nil {
				db.cache.RemoveString("post:" + id)
			}
			if replying.Valid {
				db.cache.UpdateJson("post:"+replying.String, &PostCache{atomic: true, Replies: -1})
			}
		}()
	}
	return err
}
