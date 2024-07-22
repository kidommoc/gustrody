package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
	_redis "github.com/redis/go-redis/v9"
)

var pqstTable = []struct {
	input Post
	want  Post
	cache Post
}{
	{
		input: Post{
			ID: "123", User: UD{"foo", "bar.sns"}, Replying: "",
			Vsb: utils.Vsb_PUBLIC, Content: "example", Media: []Img{
				{Type: "image/png", Url: "1.png"},
				{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
			}},
		want: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "example",
			Media: []Img{
				{Type: "image/png", Url: "1.png"},
				{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
			}},
		cache: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "example",
			Media: []Img{
				{Type: "image/png", Url: "1.png"},
				{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
			}, Likes: 0, Shares: 0},
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
			}},
		cache: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "sample",
			Media: []Img{
				{Type: "image/png", Url: "1.png", Alt: "alt text"},
				{Type: "image/jpeg", Url: "2.jpeg"},
			}, Likes: 0, Shares: 0},
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
	postDb := &PostDb{logger, client, cacheDb}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := pqstTable[0].input
		want := pqstTable[0].want
		want.Date = d
		err := postDb.SetPost(&input)
		test.AssertNoError(t, err, "Error when set: %s")
		time.Sleep(500 * time.Millisecond)

		s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertNoError(t, err, "when query cache: %s")
		wantCache := pqstTable[0].cache
		var gotCache Post
		err = json.Unmarshal([]byte(s), &gotCache)
		test.AssertNoError(t, err, "when unmarshal cache: %s")
		test.AssertEqual(t, wantCache, gotCache)
	})

	t.Run("Query", func(t *testing.T) {
		input := pqstTable[0].input
		want := pqstTable[0].want

		redis.GetDel(defaultCtx, "post:"+input.ID)

		got, err := postDb.QueryPost(input.ID)
		test.AssertNoError(t, err, "Error when query: %s")
		got.Date = d
		want.Date = d
		test.AssertEqual(t, want, got)
		time.Sleep(500 * time.Millisecond)

		s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertNoError(t, err, "when query cache: %s")
		wantCache := pqstTable[0].cache
		var gotCache Post
		err = json.Unmarshal([]byte(s), &gotCache)
		test.AssertNoError(t, err, "when unmarshal cache: %s")
		test.AssertEqual(t, wantCache, gotCache)
	})
}

func TestPostUpdate(t *testing.T) {
	d := time.Now().UTC()
	d1 := d.Add(time.Hour)

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
	postDb := &PostDb{logger, client, cacheDb}
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
		test.AssertNoError(t, err, "Error when set: %s")
	})

	t.Run("Update", func(t *testing.T) {
		input := pqstTable[1].input
		input.Date = d1
		want := pqstTable[1].want
		want.Date = d1
		err := postDb.UpdatePost(&input)
		test.AssertNoError(t, err, "Error when update: %s")
		time.Sleep(500 * time.Millisecond)

		s, err := redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertNoError(t, err, "when query cache: %s")
		wantCache := pqstTable[1].cache
		wantCache.Date = d1
		var gotCache Post
		err = json.Unmarshal([]byte(s), &gotCache)
		test.AssertNoError(t, err, "when unmarshal cache: %s")
		test.AssertEqual(t, wantCache, gotCache)
		redis.GetDel(defaultCtx, "post:"+input.ID)

		got, err := postDb.QueryPost(input.ID)
		test.AssertNoError(t, err, "Error when query: %s")
		got.Date = d1
		test.AssertEqual(t, want, got)
	})
}

func TestPostRemove(t *testing.T) {
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
	postDb := &PostDb{logger, client, cacheDb}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := pqstTable[0].input
		want := pqstTable[0].want
		want.Date = d
		err := postDb.SetPost(&input)
		test.AssertNoError(t, err, "Error when set: %s")
		test.AssertEqual(t, true, postDb.IsPostExist(input.ID))
	})

	t.Run("Remove", func(t *testing.T) {
		input := pqstTable[0].input
		err := postDb.RemovePost(input.ID)
		test.AssertNoError(t, err, "Error when remove: %s")
		test.AssertEqual(t, false, postDb.IsPostExist(input.ID))
		time.Sleep(500 * time.Millisecond)
		_, err = redis.Get(defaultCtx, "post:"+input.ID).Result()
		test.AssertEqual(t, err, _redis.Nil)
	})
}
