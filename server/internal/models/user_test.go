package models

import (
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

func TestUser(t *testing.T) {
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

	u := User{Username: UD{Username: "aaa"}, Nickname: "AAA"}
	pswd := "123456"

	t.Cleanup(func() {
		time.Sleep(500 * time.Millisecond)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, u.Username)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, u.Username)
		redis.GetDel(defaultCtx, "user:"+u.Username.String())
	})

	t.Run("set and query", func(t *testing.T) {
		input := u
		input.PubKey, input.PriKey = utils.NewKeyPair()
		err := userDb.SetLocalUser(input, pswd)
		test.AssertNoError(t, err, "Error when set: %+v")
		time.Sleep(500 * time.Millisecond)

		got, err := userDb.QueryUserByUD(input.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		want := User{
			Username: input.Username, Nickname: input.Nickname,
			Summary: input.Summary, Avatar: input.Avatar,
		}
		test.AssertEqual(t, want, *got)

		gotpswd, err := userDb.QueryPassword(input.Username.String())
		test.AssertNoError(t, err, "Error when query password: %+v")
		test.AssertEqual(t, pswd, gotpswd)

		pub, pri, err := userDb.QueryKeys(input.Username.String())
		test.AssertNoError(t, err, "Error when query keys: %+v")
		test.AssertEqual(t, input.PubKey, pub)
		test.AssertEqual(t, input.PriKey, pri)
	})

	t.Run("update", func(t *testing.T) {
		input := User{
			Username: UD{Username: "aaa"}, Nickname: "AaA", Summary: "abc",
		}
		err := userDb.UpdateUser(&input)
		test.AssertNoError(t, err, "Error when update: %+v")
		time.Sleep(500 * time.Millisecond)

		got, err := userDb.QueryUserByUD(input.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		test.AssertEqual(t, input, *got)
	})
}

func TestUserPreference(t *testing.T) {
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

	u := User{Username: UD{Username: "aaa"}, Nickname: "AAA"}

	t.Cleanup(func() {
		time.Sleep(500 * time.Millisecond)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, u.Username)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, u.Username)
		redis.GetDel(defaultCtx, "user:"+u.Username.String())
	})

	pfDefault := Preferences{
		PostVsb:  utils.Vsb_PUBLIC,
		ShareVsb: utils.Vsb_PUBLIC,
	}
	t.Run("set and query", func(t *testing.T) {
		pswd := "123456"
		err := userDb.SetLocalUser(u, pswd)
		test.AssertNoError(t, err, "Error when set user: %+v")

		gotpswd, err := userDb.QueryPassword(u.Username.String())
		test.AssertNoError(t, err, "Error when query password: %+v")
		test.AssertEqual(t, pswd, gotpswd)

		gotpf, err := userDb.QueryPreferences(u.Username.String())
		test.AssertNoError(t, err, "Error when query preferences: %+v")
		test.AssertEqual(t, pfDefault, *gotpf)
	})

	t.Run("update password", func(t *testing.T) {
		input := "654321"
		err := userDb.UpdatePassword(u.Username.String(), input)
		test.AssertNoError(t, err, "Error when update password: %+v")

		got, err := userDb.QueryPassword(u.Username.String())
		test.AssertNoError(t, err, "Error when query password: %+v")
		test.AssertEqual(t, input, got)
	})

	t.Run("update preferences", func(t *testing.T) {
		input := Preferences{
			PostVsb:  utils.Vsb_PUBLIC,
			ShareVsb: utils.Vsb_FOLLOWER,
		}
		err := userDb.UpdatePreferences(u.Username.Username, &input)
		test.AssertNoError(t, err, "Error when update: %+v")

		after, err := userDb.QueryPreferences(u.Username.Username)
		test.AssertNoError(t, err, "Error when query after: %+v")
		test.AssertEqual(t, input, *after)
	})
}

func TestUserForeign(t *testing.T) {
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

	u := User{
		Username: UD{Username: "bbb", Domain: "out.site"}, ID: "http://out.site/bbb",
		Nickname: "BBB", Summary: "xyz",
	}
	inbox := "http://out.site/bbb/inbox"
	shared := "http://out.site/inbox"

	t.Cleanup(func() {
		client.Exec(`DELETE FROM "foreign_inboxes" WHERE "username" = $1;`, u.Username)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, u.Username)
		redis.GetDel(defaultCtx, "user:"+u.Username.String())
	})

	u.PubKey, _ = utils.NewKeyPair()
	err := userDb.SetForeignUser(u, inbox, shared)
	test.AssertNoError(t, err, "Error when set foreign user: %s")

	gotu, err := userDb.QueryUserByID(u.ID)
	wantu := u
	wantu.PubKey = ""
	test.AssertEqual(t, wantu, *gotu)

	got, err := userDb.QueryInboxes([]UD{u.Username})
	want := []string{shared}
	test.AssertEqual(t, want, got)
}
