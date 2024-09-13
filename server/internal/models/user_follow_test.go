package models

import (
	"slices"
	"strings"
	"testing"

	"github.com/kidommoc/gustrody/internal/test"
)

func TestFollowSet(t *testing.T) {
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

	u1 := UD{Username: "aaa"}
	u2 := UD{Username: "bbb"}

	t.Cleanup(func() {
		client.Exec(`DELETE FROM "follow" WHERE "from" = $1;`, u1)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, u1)
		client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, u2)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, u1)
		client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, u2)
	})

	err := userDb.SetLocalUser(User{Username: u1, ID: u1.String()}, "password")
	test.AssertNoError(t, err, "Error when set u1: %s")
	err = userDb.SetLocalUser(User{Username: u2, ID: u2.String()}, "password")
	test.AssertNoError(t, err, "Error when set u2: %s")

	err = userDb.SetFollow(u1, u2)
	test.AssertNoError(t, err, "Error when set follow from u1 to u2: %s")
	got := userDb.IsFollowing(u1, u2)
	test.AssertEqual(t, true, got)
}

func TestFollowQuery(t *testing.T) {
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

	utable := []UD{
		NewUD("u1"), NewUD("u2"), NewUD("u3"), NewUD("u4"), NewUD("u5"),
	}

	t.Cleanup(func() {
		for _, v := range utable {
			client.Exec(`DELETE FROM "follow" WHERE "from" = $1 OR "to" = $1;`, v.Username)
		}
		for _, v := range utable {
			client.Exec(`DELETE FROM "user_preferences" WHERE "username" = $1;`, v.Username)
			client.Exec(`DELETE FROM "users" WHERE "username" = $1;`, v.Username)
		}
	})

	for _, v := range utable {
		err := userDb.SetLocalUser(User{Username: v, ID: v.String()}, "password")
		test.AssertNoError(t, err, "Error when set "+v.String()+": %s")
	}
	for i := range utable {
		for j := i + 1; j < len(utable); j++ {
			err := userDb.SetFollow(utable[j], utable[i])
			test.AssertNoError(t, err, "Error when set follow from "+utable[j].String()+" to "+utable[i].String()+": %s")
		}
	}

	cmpUD := func(a UD, b UD) int {
		return strings.Compare(a.String(), b.String())
	}

	// check followings and followers
	for i, v := range utable {
		wantFollows := int64(i)
		wantFollowed := int64(len(utable) - i - 1)
		gotFollows, gotFollowed, err := userDb.QueryFollowInfo(v.String())
		test.AssertNoError(t, err, "Error when query follow info of "+v.String()+": %s")
		test.AssertEqual(t, wantFollows, gotFollows)
		test.AssertEqual(t, wantFollowed, gotFollowed)

		wantFollowings := utable[:i]
		gotFollowings, err := userDb.QueryFollowings(v.String())
		test.AssertNoError(t, err, "Error when query followings of "+v.String()+": %s")
		slices.SortFunc(wantFollowings, cmpUD)
		slices.SortFunc(gotFollowings, cmpUD)
		test.AssertEqual(t, wantFollowings, gotFollowings)

		wantFollowers := []UD{}
		if i < len(utable)-1 {
			wantFollowers = utable[i+1:]
		}
		gotFollowers, err := userDb.QueryFollowers(v.String())
		test.AssertNoError(t, err, "Error when query followers of "+v.String()+": %s")
		slices.SortFunc(wantFollowers, cmpUD)
		slices.SortFunc(gotFollowers, cmpUD)
		test.AssertEqual(t, wantFollowers, gotFollowers)
	}
}
