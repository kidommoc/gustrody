package models

import (
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var uatTableU = []User{
	{
		Username: UD{Username: "aaa"}, Nickname: "AAA",
	},
	{
		Username: UD{Username: "aaa"}, Nickname: "AaA",
		Summary: "abcdefg",
	},
}

var uatTablePf = []Preferences{
	{
		PostVsb:  utils.Vsb_PUBLIC.String(),
		ShareVsb: utils.Vsb_PUBLIC.String(),
	},
	{
		PostVsb:  utils.Vsb_PUBLIC.String(),
		ShareVsb: utils.Vsb_FOLLOWER.String(),
	},
}

func TestUserSetAndQuery(t *testing.T) {
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
	t.Cleanup(func() {
		for _, v := range uatTableU {
			client.Exec(`DELETE FROM users WHERE "username" = $1;`, v.Username)
			redis.GetDel(defaultCtx, "user:"+v.Username.String())
		}
	})

	input := uatTableU[0]
	t.Run("Set", func(t *testing.T) {
		input.Keys.Pub, input.Keys.Pri = utils.NewKeyPair()
		err := userDb.SetUser(&input)
		test.AssertNoError(t, err, "Error when set: %+v")
		time.Sleep(500 * time.Millisecond)
	})

	t.Run("Query", func(t *testing.T) {
		got, err := userDb.QueryUser(input.Username.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		want := User{
			Username: input.Username, Nickname: input.Nickname,
			Summary: input.Summary, Avatar: input.Avatar,
		}
		test.AssertEqual(t, want, got)

		pub, pri, err := userDb.QueryUserKeys(input.Username.Username)
		test.AssertNoError(t, err, "Error when query keys: %+v")
		test.AssertEqual(t, input.Keys.Pub, pub)
		test.AssertEqual(t, input.Keys.Pri, pri)
	})
}

func TestUserUpdate(t *testing.T) {
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
	t.Cleanup(func() {
		for _, v := range uatTableU {
			client.Exec(`DELETE FROM users WHERE "username" = $1;`, v.Username)
			redis.GetDel(defaultCtx, "user:"+v.Username.String())
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := uatTableU[0]
		err := userDb.SetUser(&input)
		test.AssertNoError(t, err, "Error when set: %+v")
	})

	t.Run("Update", func(t *testing.T) {
		input := uatTableU[1]
		err := userDb.UpdateUser(&input)
		test.AssertNoError(t, err, "Error when update: %+v")
		time.Sleep(500 * time.Millisecond)

		got, err := userDb.QueryUser(input.Username.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		want := User{
			Username: input.Username, Nickname: input.Nickname,
			Summary: input.Summary, Avatar: input.Avatar,
		}
		test.AssertEqual(t, want, got)

		input = uatTableU[0]
		err = userDb.UpdateUser(&input)
		test.AssertNoError(t, err, "Error when update: %+v")
		time.Sleep(500 * time.Millisecond)

		got, err = userDb.QueryUser(input.Username.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		want = User{
			Username: input.Username, Nickname: input.Nickname,
			Summary: input.Summary, Avatar: input.Avatar,
		}
		test.AssertEqual(t, want, got)
	})
}

func TestUserPreferenceUpdate(t *testing.T) {
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
	t.Cleanup(func() {
		for _, v := range uatTableU {
			client.Exec(`DELETE FROM users WHERE "username" = $1;`, v.Username)
			redis.GetDel(defaultCtx, "user:"+v.Username.String())
		}
	})

	inputU := uatTableU[0]
	input := uatTablePf[1]
	t.Run("set and query", func(t *testing.T) {
		err := userDb.SetUser(&inputU)
		test.AssertNoError(t, err, "Error when set user: %+v")
		before, err := userDb.QueryUserPreferences(inputU.Username.Username)
		want := uatTablePf[0]
		test.AssertEqual(t, want, *before)
		test.AssertNoError(t, err, "Error when query before: %+v")
	})

	t.Run("update", func(t *testing.T) {
		err := userDb.UpdateUserPreferences(inputU.Username.Username, &input)
		test.AssertNoError(t, err, "Error when update: %+v")
		after, err := userDb.QueryUserPreferences(inputU.Username.Username)
		test.AssertNoError(t, err, "Error when query after: %+v")
		test.AssertEqual(t, input, *after)
	})
}
