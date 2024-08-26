package models

import (
	"encoding/json"
	"regexp"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
	_redis "github.com/redis/go-redis/v9"
)

var pqstTable = []struct {
	input Post
	want  Post
}{
	{
		input: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "example",
			Media: []Img{
				{Type: "image/png", Url: "1.png"},
				{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
			}, Replies: []string{}},
	},
	{
		input: Post{ID: "123", Content: "sample", Media: []Img{
			{Type: "image/png", Url: "1.png", Alt: "alt text"},
			{Type: "image/jpeg", Url: "2.jpeg"},
		}},
		want: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "sample",
			Media: []Img{
				{Type: "image/png", Url: "1.png", Alt: "alt text"},
				{Type: "image/jpeg", Url: "2.jpeg"},
			}, Replies: []string{}},
	},
}

func TestPostSetAndQuery(t *testing.T) {
	d := time.Now().UTC()

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
	postDb := &PostDb{logger, client, cacheDb, regexp.MustCompile(`(post|share):([0-9a-z]+)`)}

	t.Cleanup(func() {
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := pqstTable[0].input
		input.Date = d
		err := postDb.SetPost(&input)
		test.AssertNoError(t, err, "when set: %s")
		time.Sleep(500 * time.Millisecond)

		s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertNoError(t, err, "when query cache: %s")
		wantCache := pqstTable[0].input
		var gotCache Post
		err = json.Unmarshal([]byte(s), &gotCache)
		test.AssertNoError(t, err, "when unmarshal cache: %s")
		wantCache.Date = gotCache.Date
		if gotCache.Replies == nil {
			gotCache.Replies = []string{}
		}
		test.AssertEqual(t, wantCache, gotCache)
	})

	t.Run("Query", func(t *testing.T) {
		input := pqstTable[0].input
		want := pqstTable[0].input
		redis.GetDel(defaultCtx, "post:"+input.ID)

		got, err := postDb.QueryPost(input.ID)
		test.AssertNoError(t, err, "when query: %s")
		want.Date = got.Date
		time.Sleep(500 * time.Millisecond)

		s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertNoError(t, err, "when query cache: %s")
		wantCache := pqstTable[0].input
		var gotCache Post
		err = json.Unmarshal([]byte(s), &gotCache)
		test.AssertNoError(t, err, "when unmarshal cache: %s")
		wantCache.Date = gotCache.Date
		if gotCache.Replies == nil {
			gotCache.Replies = []string{}
		}
		test.AssertEqual(t, wantCache, gotCache)
	})
}

func TestPostUpdate(t *testing.T) {
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
	postDb := &PostDb{logger, client, cacheDb, regexp.MustCompile(`(post|share):([0-9a-z]+)`)}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})
	d := time.Now().UTC()
	d1 := d.Add(time.Hour)

	t.Run("Set and Update", func(t *testing.T) {
		input := pqstTable[0].input
		input.Date = d
		err := postDb.SetPost(&input)
		test.AssertNoError(t, err, "Error when set: %s")

		input = pqstTable[1].input
		input.Date = d1
		want := pqstTable[1].want
		want.Date = d1
		err = postDb.UpdatePost(&input)
		test.AssertNoError(t, err, "Error when update: %s")
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("Query", func(t *testing.T) {
		input := pqstTable[1].input
		want := pqstTable[1].want

		s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertNoError(t, err, "when query cache: %s")
		var gotCache Post
		err = json.Unmarshal([]byte(s), &gotCache)
		test.AssertNoError(t, err, "when unmarshal cache: %s")
		if gotCache.Replies == nil {
			gotCache.Replies = []string{}
		}
		wantCache := pqstTable[1].want
		wantCache.Date = gotCache.Date
		test.AssertEqual(t, wantCache, gotCache)
		redis.GetDel(defaultCtx, "post:"+input.ID)

		got, err := postDb.QueryPost(input.ID)
		test.AssertNoError(t, err, "Error when query: %s")
		want.Date = got.Date
		test.AssertEqual(t, want, got)
	})

}

func TestPostRemove(t *testing.T) {
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
	postDb := &PostDb{logger, client, cacheDb, regexp.MustCompile(`(post|share):([0-9a-z]+)`)}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})
	d := time.Now().UTC()

	input := pqstTable[0].input
	want := pqstTable[0].want
	want.Date = d
	err := postDb.SetPost(&input)
	test.AssertNoError(t, err, "Error when set: %s")
	test.AssertEqual(t, true, postDb.IsPostExist(input.ID))

	input = pqstTable[0].input
	err = postDb.RemovePost(input.ID)
	test.AssertNoError(t, err, "Error when remove: %s")
	test.AssertEqual(t, false, postDb.IsPostExist(input.ID))
	time.Sleep(500 * time.Millisecond)
	_, err = redis.Get(defaultCtx, "post:"+input.ID).Result()
	test.AssertEqual(t, err, _redis.Nil)
}

func TestPostUserContent(t *testing.T) {
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
	postDb := &PostDb{logger, client, cacheDb, regexp.MustCompile(`(post|share):([0-9a-z]+)`)}
	d := time.Now().Add(-5 * time.Minute).UTC()

	ptable := []Post{
		{ID: "testid1", User: UD{"testuser1", ""}, Replies: []string{"testid3", "testid2"}, Vsb: utils.Vsb_PUBLIC, Content: "u1-1"},
		{ID: "testid2", User: UD{"testuser2", ""}, Replying: "testid1", ReplyTo: UD{"u1", ""}, Vsb: utils.Vsb_PUBLIC, Content: "u2-1"},
		{ID: "testid3", User: UD{"testuser1", ""}, Replying: "testid1", ReplyTo: UD{"u1", ""}, Vsb: utils.Vsb_PUBLIC, Content: "u1-2"},
	}

	for i := 0; i < len(ptable); i++ {
		ptable[i].Date = d
		ptable[i].ActDate = d
		d = d.Add(time.Minute)
		err := postDb.SetPost(&ptable[i])
		test.AssertNoError(t, err, "when set post#"+ptable[i].ID+": %s")
	}

	err := postDb.SetShare(UD{"testuser1", ""}, "testid2", d, utils.Vsb_PUBLIC)
	test.AssertNoError(t, err, "when set shares#testid2: %s")

	time.Sleep(2 * time.Second)

	t.Cleanup(func() {
		client.Exec(`DELETE FROM shares WHERE "id" = 'testid2';`)
		for _, v := range ptable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.ID)
			redis.GetDel(defaultCtx, "post:"+v.ID)
		}
		redis.LTrim(defaultCtx, "ugc:testuser1", 1, 0)
	})

	want := []Post{
		ptable[0], ptable[2], {
			ID: ptable[1].ID, User: ptable[1].User, SharedBy: UD{"testuser1", ""},
			Vsb: ptable[1].Vsb, Content: ptable[1].Content,
			Date: ptable[1].Date, ActDate: d, Shares: 1,
		},
	}

	got, err := postDb.QueryUserContent("testuser1", time.Now())
	test.AssertNoError(t, err, "when query by user: %s")
	slices.Reverse(want)
	for i := 0; i < len(got); i++ {
		want[i].Date = got[i].Date
		want[i].ActDate = got[i].ActDate
	}
	test.AssertEqual(t, want, got)
}

func TestQueryUserContent(t *testing.T) {
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
	postDb := &PostDb{logger, client, cacheDb, regexp.MustCompile(`(post|share):([0-9a-z]+)`)}

	d := time.Now().UTC()
	t.Log(d)
	d1 := d.Add((-10*cachePageSize - 1) * time.Minute)
	userid := "testuser"
	userud := NewUD(userid)
	ptable := []Post{}
	count := 0
	for i := 0; i < 10; i++ {
		for j := 0; j < cachePageSize; j++ {
			count++
			cs := strconv.Itoa(count)
			p := Post{
				ID: "testid" + cs, User: userud, Vsb: utils.Vsb_PUBLIC,
				Content: cs, Date: d1.Add(time.Duration(count) * time.Minute),
			}
			ptable = append(ptable, p)
			postDb.SetPost(&p)
		}
	}
	slices.Reverse(ptable)

	t.Cleanup(func() {
		// time.Sleep(5 * time.Second)
		for _, v := range ptable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.ID)
			redis.GetDel(defaultCtx, "post:"+v.ID)
		}
		redis.LTrim(defaultCtx, "ugc:"+userid, 1, 0)
	})

	t.Run("query latest user content", func(t *testing.T) {
		got, err := postDb.QueryUserContent(userid, d)
		test.AssertNoError(t, err, "when query latest user content: %s")
		want := ptable[:cachePageSize]
		for i := 0; i < len(want); i++ {
			want[i].Date = got[i].Date
			want[i].ActDate = got[i].ActDate
		}
		test.AssertEqual(t, want, got)
		time.Sleep(5 * time.Second)
		l, _ := redis.LRange(defaultCtx, "ugc:"+userid, 0, 0).Result()
		var idx []string
		json.Unmarshal([]byte(l[0]), &idx)
		test.AssertEqual(t, 3, len(idx))
	})

	t.Run("query no enough pages", func(t *testing.T) {
		got, err := postDb.QueryUserContent(userid, d.Add((-2*cachePageSize)*time.Minute))
		test.AssertNoError(t, err, "when query no enough pages: %s")
		want := ptable[2*cachePageSize : 3*cachePageSize]
		for i := 0; i < len(want); i++ {
			want[i].Date = got[i].Date
			want[i].ActDate = got[i].ActDate
		}
		test.AssertEqual(t, want, got)
		time.Sleep(5 * time.Second)
		l, _ := redis.LRange(defaultCtx, "ugc:"+userid, 0, 0).Result()
		var idx []string
		json.Unmarshal([]byte(l[0]), &idx)
		test.AssertEqual(t, 6, len(idx))
	})

	t.Run("query not found pages", func(t *testing.T) {
		got, err := postDb.QueryUserContent(userid, d.Add((-7*cachePageSize)*time.Minute))
		test.AssertNoError(t, err, "when query no enough pages: %s")
		want := ptable[7*cachePageSize : 8*cachePageSize]
		for i := 0; i < len(want); i++ {
			want[i].Date = got[i].Date
			want[i].ActDate = got[i].ActDate
		}
		test.AssertEqual(t, want, got)
		time.Sleep(5 * time.Second)
		l, _ := redis.LRange(defaultCtx, "ugc:"+userid, 0, 0).Result()
		var idx []string
		json.Unmarshal([]byte(l[0]), &idx)
		test.AssertEqual(t, 10, len(idx))
	})
}
