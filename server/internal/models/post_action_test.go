package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

func TestLike(t *testing.T) {
	logger := test.NewMockingLogger(t)
	client := initMainDb(modelscfg, logger, pqOpt{
		Addr: "localhost:5432", MaxConn: 5,
	})
	redis := initRedis(modelscfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cacheDb := &CacheDb{logger, redis}
	userDb := &UserDb{logger, client, cacheDb}
	postDb := &PostDb{logger, client, cacheDb}

	ud := NewUD("testuser")
	u := User{Username: ud, ID: ud.String()}
	p := Post{
		ID: "123", FederalID: "123", User: ud, Date: time.Now().UTC(),
		Vsb: utils.Vsb_PUBLIC, Content: "example",
	}
	err := userDb.SetLocalUser(u, "123456")
	test.AssertNoError(t, err, "when set test user: %s")
	err = postDb.SetPost(p)
	test.AssertNoError(t, err, "when set test post: %s")
	time.Sleep(500 * time.Millisecond)

	t.Cleanup(func() {
		client.Exec(`DELETE FROM "like" WHERE "user" = $1 OR "tgt" = $2;`, ud, p.ID)
		client.Exec(`DELETE FROM "posts" WHERE "id" = $1;`, p.ID)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, ud)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, ud)
		redis.GetDel(defaultCtx, "post:"+p.ID)
	})

	// LIKE
	err = postDb.SetLike(ud, p.ID)
	test.AssertNoError(t, err, "when set like: %s")

	exist := func() bool {
		r, err := client.Query(`SELECT 1 FROM "like" WHERE "user" = $1 OR "tgt" = $2;`, ud, p.ID)
		test.AssertNoError(t, err, "when query like: %s")
		defer r.Close()
		for r.Next() {
			return true
		}
		return false
	}()
	test.AssertEqual(t, true, exist)
	time.Sleep(500 * time.Millisecond)

	// check cache
	gots, err := cacheDb.QueryString([]string{"post:" + p.ID})
	test.AssertNoError(t, err, "when query cache after set like: %s")
	gotCache := gots["post:"+p.ID]
	var got Post
	err = json.Unmarshal([]byte(gotCache), &got)
	test.AssertNoError(t, err, "when unmarshal cache after set like: %s")
	test.AssertEqual(t, int64(1), got.Likes)

	// REMOVE LIKE

	err = postDb.RemoveLike(ud, p.ID)
	test.AssertNoError(t, err, "when remove like: %s")

	exist = func() bool {
		r, err := client.Query(`SELECT 1 FROM "like" WHERE "user" = $1 OR "tgt" = $2;`, ud, p.ID)
		test.AssertNoError(t, err, "when query like: %s")
		defer r.Close()
		for r.Next() {
			return true
		}
		return false
	}()
	test.AssertEqual(t, false, exist)
	time.Sleep(500 * time.Millisecond)

	// check cache
	gots, err = cacheDb.QueryString([]string{"post:" + p.ID})
	test.AssertNoError(t, err, "when query cache after remove like: %s")
	gotCache = gots["post:"+p.ID]
	err = json.Unmarshal([]byte(gotCache), &got)
	test.AssertNoError(t, err, "when unmarshal cache after remove like: %s")
	test.AssertEqual(t, int64(0), got.Likes)
}

func TestShare(t *testing.T) {
	logger := test.NewMockingLogger(t)
	client := initMainDb(modelscfg, logger, pqOpt{
		Addr: "localhost:5432", MaxConn: 5,
	})
	redis := initRedis(modelscfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cacheDb := &CacheDb{logger, redis}
	userDb := &UserDb{logger, client, cacheDb}
	postDb := &PostDb{logger, client, cacheDb}

	ud := NewUD("testuser")
	u := User{Username: ud, ID: ud.String()}
	p := Post{
		ID: "123", FederalID: "123", User: ud, Date: time.Now().UTC().Add(-1 * time.Hour),
		Vsb: utils.Vsb_PUBLIC, Content: "example",
	}

	t.Cleanup(func() {
		client.Exec(`DELETE FROM "share" WHERE "user" = $1 OR "tgt" = $2;`, ud, p.ID)
		client.Exec(`DELETE FROM "posts" WHERE "id" = $1;`, p.ID)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, ud)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, ud)
		redis.GetDel(defaultCtx, "post:"+p.ID)
	})

	err := userDb.SetLocalUser(u, "123456")
	test.AssertNoError(t, err, "when set test user: %s")
	err = postDb.SetPost(p)
	test.AssertNoError(t, err, "when set test post: %s")
	time.Sleep(500 * time.Millisecond)

	// SHARE
	err = postDb.SetShare(ud, p.ID, time.Now().UTC(), utils.Vsb_PUBLIC)
	test.AssertNoError(t, err, "when set share: %s")

	exist := func() bool {
		r, err := client.Query(`SELECT 1 FROM "share" WHERE "user" = $1 OR "tgt" = $2;`, ud, p.ID)
		test.AssertNoError(t, err, "when query share: %s")
		defer r.Close()
		for r.Next() {
			return true
		}
		return false
	}()
	test.AssertEqual(t, true, exist)
	time.Sleep(500 * time.Millisecond)

	// check cache
	gots, err := cacheDb.QueryString([]string{"post:" + p.ID})
	test.AssertNoError(t, err, "when query cache after set share: %s")
	gotCache := gots["post:"+p.ID]
	var got Post
	err = json.Unmarshal([]byte(gotCache), &got)
	test.AssertNoError(t, err, "when unmarshal cache after set share: %s")
	test.AssertEqual(t, int64(1), got.Shares)

	// REMOVE SHARE

	err = postDb.RemoveShare(ud, p.ID)
	test.AssertNoError(t, err, "when remove share: %s")

	exist = func() bool {
		r, err := client.Query(`SELECT 1 FROM "share" WHERE "user" = $1 OR "tgt" = $2;`, ud, p.ID)
		test.AssertNoError(t, err, "when query share: %s")
		defer r.Close()
		for r.Next() {
			return true
		}
		return false
	}()
	test.AssertEqual(t, false, exist)
	time.Sleep(500 * time.Millisecond)

	// check cache
	gots, err = cacheDb.QueryString([]string{"post:" + p.ID})
	test.AssertNoError(t, err, "when query cache after remove share: %s")
	gotCache = gots["post:"+p.ID]
	err = json.Unmarshal([]byte(gotCache), &got)
	test.AssertNoError(t, err, "when unmarshal cache after remove share: %s")
	test.AssertEqual(t, int64(0), got.Shares)
}
