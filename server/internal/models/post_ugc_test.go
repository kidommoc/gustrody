package models

import (
	"encoding/json"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
)

func TestUGC(t *testing.T) {
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

	ud1 := NewUD("testuser1")
	ud2 := NewUD("testuser2")
	u1 := User{Username: ud1, ID: ud1.String()}
	u2 := User{Username: ud2, ID: ud2.String()}
	ptable := []Post{
		{ID: "testid1", FederalID: "testid1", User: ud1, Vsb: utils.Vsb_PUBLIC, Content: "u1-1", Replies: 2},
		{ID: "testid2", FederalID: "testid2", User: ud2, Vsb: utils.Vsb_PUBLIC, Content: "u2-1", Replying: "testid1", ReplyTo: ud1},
		{ID: "testid3", FederalID: "testid3", User: ud1, Vsb: utils.Vsb_FOLLOWER, Content: "u1-2", Replying: "testid1", ReplyTo: ud1},
	}

	t.Cleanup(func() {
		for _, v := range ptable {
			client.Exec(`DELETE FROM "reply" WHERE "id" = $1 OR "tgt" = $1;`, v.ID)
			client.Exec(`DELETE FROM "share" WHERE "tgt" = $1;`, v.ID)
		}
		for _, v := range ptable {
			client.Exec(`DELETE FROM "posts" WHERE "id" = $1;`, v.ID)
			redis.GetDel(defaultCtx, "post:"+v.ID)
		}
		redis.LTrim(defaultCtx, "ugc:testuser1", 1, 0)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = ANY($1);`, pq.Array([]UD{ud1, ud2}))
		client.Exec(`DELETE FROM "users" WHERE "username" = ANY($1);`, pq.Array([]UD{ud1, ud2}))
	})

	err := userDb.SetLocalUser(u1, "123456")
	test.AssertNoError(t, err, "when set test user1: %s")
	err = userDb.SetLocalUser(u2, "123456")
	test.AssertNoError(t, err, "when set test user1: %s")

	d := time.Now().Add(-5 * time.Minute).UTC()
	for i := 0; i < len(ptable); i++ {
		ptable[i].Date = d
		d = d.Add(time.Minute)
		err := postDb.SetPost(ptable[i])
		test.AssertNoError(t, err, "when set post#"+ptable[i].ID+": %s")
	}
	err = postDb.SetShare(ud1, "testid2", d, utils.Vsb_PUBLIC)
	test.AssertNoError(t, err, "when set shares#testid2: %s")
	time.Sleep(2 * time.Second)

	sp := ptable[1]
	sp.SharedBy = ud1
	sp.Shares = 1
	want := []Post{sp, ptable[0]}

	got, err := postDb.QueryUserContent("testuser1", time.Now(), utils.Vsb_PUBLIC)
	test.AssertNoError(t, err, "when query by user: %s")
	for i := 0; i < len(got); i++ {
		want[i].Date = got[i].Date
	}
	test.AssertEqual(t, want, got)
}

func TestQueryUGC(t *testing.T) {
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

	userid := "testuser"
	userud := NewUD(userid)
	u := User{Username: userud, ID: userid}
	err := userDb.SetLocalUser(u, "123456")
	test.AssertNoError(t, err, "when set test user: %s")

	ptable := []Post{}
	d := time.Now().UTC()
	d1 := d.Add((-10*cachePageSize - 1) * time.Minute)
	count := 0
	for i := 0; i < 10; i++ {
		for j := 0; j < cachePageSize; j++ {
			count++
			cs := strconv.Itoa(count)
			p := Post{
				ID: "testid" + cs, FederalID: "testid" + cs, User: userud,
				Vsb: utils.Vsb_PUBLIC, Content: cs,
				Date: d1.Add(time.Duration(count) * time.Minute),
			}
			ptable = append(ptable, p)
			err := postDb.SetPost(p)
			test.AssertNoError(t, err, "when set test post#"+cs+": %s")
		}
	}
	slices.Reverse(ptable)
	vsb := utils.Vsb_PUBLIC

	t.Cleanup(func() {
		// time.Sleep(5 * time.Second)
		for _, v := range ptable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.ID)
			redis.GetDel(defaultCtx, "post:"+v.ID)
		}
		err := redis.LTrim(defaultCtx, "ugc:"+userid, 1, 0).Err()
		if err != nil {
			t.Log(err)
		}
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, userud)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, userud)
	})

	t.Run("query latest user content", func(t *testing.T) {
		startAt := time.Now()
		got, err := postDb.QueryUserContent(userid, d, vsb)
		cost := time.Now().Sub(startAt)
		t.Logf("time cost: %s", cost)
		test.AssertNoError(t, err, "when query latest user content: %s")
		want := ptable[:cachePageSize]
		for i := 0; i < len(want); i++ {
			want[i].Date = got[i].Date
		}
		test.AssertEqual(t, want, got)
		time.Sleep(2 * time.Second)
		l, _ := redis.LRange(defaultCtx, "ugc:"+userid, 0, 0).Result()
		var idx []string
		json.Unmarshal([]byte(l[0]), &idx)
		test.AssertEqual(t, 3, len(idx))
	})

	// query same content (cached)
	t.Run("benchmark", func(t *testing.T) {
		startAt := time.Now()
		postDb.QueryUserContent(userid, d, vsb)
		cost := time.Now().Sub(startAt)
		t.Logf("time cost: %s", cost)
	})

	t.Run("query no enough pages", func(t *testing.T) {
		startAt := time.Now()
		got, err := postDb.QueryUserContent(userid, d.Add((-2*cachePageSize)*time.Minute), vsb)
		cost := time.Now().Sub(startAt)
		t.Logf("time cost: %s", cost)
		test.AssertNoError(t, err, "when query no enough pages: %s")
		want := ptable[2*cachePageSize : 3*cachePageSize]
		for i := 0; i < len(want); i++ {
			want[i].Date = got[i].Date
		}
		test.AssertEqual(t, want, got)
		time.Sleep(2 * time.Second)
		l, _ := redis.LRange(defaultCtx, "ugc:"+userid, 0, 0).Result()
		var idx []string
		json.Unmarshal([]byte(l[0]), &idx)
		test.AssertEqual(t, 6, len(idx))
	})

	t.Run("query not found pages", func(t *testing.T) {
		startAt := time.Now()
		got, err := postDb.QueryUserContent(userid, d.Add((-7*cachePageSize)*time.Minute), vsb)
		cost := time.Now().Sub(startAt)
		t.Logf("time cost: %s", cost)
		test.AssertNoError(t, err, "when query no enough pages: %s")
		want := ptable[7*cachePageSize : 8*cachePageSize]
		for i := 0; i < len(want); i++ {
			want[i].Date = got[i].Date
		}
		test.AssertEqual(t, want, got)
		time.Sleep(2 * time.Second)
		l, _ := redis.LRange(defaultCtx, "ugc:"+userid, 0, 0).Result()
		var idx []string
		json.Unmarshal([]byte(l[0]), &idx)
		test.AssertEqual(t, 10, len(idx))
	})
}
