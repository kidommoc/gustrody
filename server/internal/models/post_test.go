package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
)

func TestPostSetAndQuery(t *testing.T) {
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

	ud := UD{Username: "foobar"}
	u := User{Username: ud, ID: ud.String()}
	err := userDb.SetLocalUser(u, "password")
	test.AssertNoError(t, err, "when set test user: %s")

	d := time.Now().UTC()
	input := Post{
		ID: "123", FederalID: "123", User: ud, Date: d,
		Vsb: utils.Vsb_PUBLIC, Content: "example", Media: []Img{
			{Type: "image/png", Url: "1.png"},
			{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
		},
	}
	rinput := Post{
		ID: "456", FederalID: "456", User: ud, Date: d,
		Vsb: utils.Vsb_PUBLIC, Content: "example reply",
		Replying: input.ID,
	}

	t.Cleanup(func() {
		client.Exec(`DELETE FROM "reply" WHERE "id" = ANY($1);`, pq.Array([]string{input.ID, rinput.ID}))
		client.Exec(`DELETE FROM "posts" WHERE "id" = ANY($1);`, pq.Array([]string{input.ID, rinput.ID}))
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, ud)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, ud)
		redis.GetDel(defaultCtx, "post:"+input.ID)
		redis.GetDel(defaultCtx, "post:"+rinput.ID)
	})

	// SET

	err = postDb.SetPost(input)
	test.AssertNoError(t, err, "when set: %s")
	time.Sleep(500 * time.Millisecond)

	// check cache
	s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
	test.AssertNoError(t, err, "when query cache after set: %s")
	wantCache := input
	var gotCache Post
	err = json.Unmarshal([]byte(s), &gotCache)
	test.AssertNoError(t, err, "when unmarshal cache: %s")
	wantCache.Date = gotCache.Date
	wantCache.Type = postJsonType
	test.AssertEqual(t, wantCache, gotCache)

	// QUERY

	want := input
	err = redis.GetDel(defaultCtx, "post:"+input.ID).Err()
	test.AssertNoError(t, err, "when remove cache: %s")
	time.Sleep(500 * time.Millisecond)

	gots, err := postDb.QueryPosts([]string{input.ID})
	test.AssertNoError(t, err, "when query: %s")
	if gots[input.ID] == nil {
		test.AssertNoError(t, fmt.Errorf("Not Found"), "when query: %s")
	}
	got := *gots[input.ID]
	want.Date = got.Date
	test.AssertEqual(t, want, got)
	time.Sleep(500 * time.Millisecond)

	// check cache
	s, err = redis.Get(defaultCtx, "post:"+input.ID).Result()
	test.AssertNoError(t, err, "when query cache after query: %s")
	wantCache = input
	err = json.Unmarshal([]byte(s), &gotCache)
	test.AssertNoError(t, err, "when unmarshal cache: %s")
	wantCache.Date = gotCache.Date
	wantCache.Type = postJsonType
	test.AssertEqual(t, wantCache, gotCache)

	// REPLY
	err = postDb.SetPost(rinput)
	test.AssertNoError(t, err, "when set reply post: %s")
	time.Sleep(500 * time.Millisecond)

	// check the replied one
	rply, rpls, err := postDb.QueryPostReplyChain(input.ID)
	test.AssertNoError(t, err, "when query reply chain of the replied post: %s")
	wantRpls := []ReplyData{{ID: rinput.ID, To: input.ID}}
	test.AssertEqual(t, 0, len(rply))
	test.AssertEqual(t, wantRpls, rpls)

	// check the replying one
	rply, rpls, err = postDb.QueryPostReplyChain(rinput.ID)
	test.AssertNoError(t, err, "when query reply chain of the replying post: %s")
	wantRply := []ReplyData{{ID: rinput.ID, To: input.ID}}
	test.AssertEqual(t, wantRply, rply)
	test.AssertEqual(t, 0, len(rpls))
}

func TestPostUpdateAndRemove(t *testing.T) {
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

	ud := UD{Username: "foobar"}
	u := User{Username: ud, ID: ud.String()}
	err := userDb.SetLocalUser(u, "password")
	test.AssertNoError(t, err, "when set test user: %s")

	d := time.Now().UTC()
	input := Post{
		ID: "123", FederalID: "123", User: ud, Date: d,
		Vsb: utils.Vsb_PUBLIC, Content: "example", Media: []Img{
			{Type: "image/png", Url: "1.png"},
			{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
		},
	}
	uinput := Post{
		ID: "123", Content: "sample", Media: []Img{
			{Type: "image/png", Url: "1.png", Alt: "alt text"},
			{Type: "image/jpeg", Url: "2.jpeg"},
		},
	}
	uwant := Post{
		ID: "123", FederalID: "123", User: ud,
		Vsb: utils.Vsb_PUBLIC, Content: "sample", Media: []Img{
			{Type: "image/png", Url: "1.png", Alt: "alt text"},
			{Type: "image/jpeg", Url: "2.jpeg"},
		},
	}
	rinput := Post{
		ID: "456", FederalID: "456", User: ud,
		Vsb: utils.Vsb_PUBLIC, Content: "to remove", Replying: input.ID,
	}

	t.Cleanup(func() {
		client.Exec(`DELETE FROM "posts" WHERE "id" = $1;`, input.ID)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, ud)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, ud)
		redis.GetDel(defaultCtx, "post:"+input.ID)
	})

	err = postDb.SetPost(input)
	test.AssertNoError(t, err, "when set: %s")
	err = postDb.SetPost(rinput)
	test.AssertNoError(t, err, "when set: %s")
	time.Sleep(500 * time.Millisecond)

	// UPDATE

	err = postDb.UpdatePost(uinput)
	test.AssertNoError(t, err, "when update: %s")
	time.Sleep(500 * time.Millisecond)

	// check cache
	s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
	test.AssertNoError(t, err, "when query cache: %s")
	var gotCache Post
	err = json.Unmarshal([]byte(s), &gotCache)
	test.AssertNoError(t, err, "when unmarshal cache: %s")
	wantCache := uwant
	wantCache.Date = gotCache.Date
	wantCache.Replies = 1
	wantCache.Type = postJsonType
	test.AssertEqual(t, wantCache, gotCache)
	redis.GetDel(defaultCtx, "post:"+input.ID)

	// check
	gots, err := postDb.QueryPosts([]string{input.ID})
	test.AssertNoError(t, err, "when query updated post: %s")
	if gots[input.ID] == nil {
		test.AssertNoError(t, fmt.Errorf("Not Found"), "when query: %s")
	}
	got := *gots[input.ID]
	want := uwant
	want.Date = got.Date
	want.Replies = 1
	test.AssertEqual(t, want, got)

	// REMOVE

	err = postDb.RemovePost(rinput.ID)
	test.AssertNoError(t, err, "when remove post: %s")
	time.Sleep(500 * time.Millisecond)

	exist := postDb.IsPostExist(rinput.ID)
	test.AssertEqual(t, false, exist)

	err = func() error {
		r, err := client.Query(`SELECT 1 FROM "posts" WHERE "id" = $1;`, rinput.ID)
		if err != nil {
			return err
		}
		defer r.Close()
		for r.Next() {
			return fmt.Errorf("not removed")
		}
		return nil
	}()
	test.AssertNoError(t, err, "when query removed post: %s")

	// check cache
	s, err = redis.Get(defaultCtx, "post:"+rinput.ID).Result()
	test.AssertNoError(t, err, "when check tombstone cache: %s")
	test.AssertEqual(t, `{"tombstone":true}`, s)

	// check reply
	err = func() error {
		r, err := client.Query(`SELECT 1 FROM "reply" WHERE "id" = $1 AND "tgt" = $2;`, rinput.ID, input.ID)
		if err != nil {
			return err
		}
		defer r.Close()
		for r.Next() {
			return fmt.Errorf("not removed")
		}
		return nil
	}()
	test.AssertNoError(t, err, "when query removed reply: %s")
}
