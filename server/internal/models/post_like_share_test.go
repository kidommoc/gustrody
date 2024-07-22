package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

type plstInput struct {
	actor  UD
	target string
}

var plstTable = []struct {
	input plstInput
	want  []UD
}{
	{
		input: plstInput{actor: UD{"sugar", ""}, target: "123"},
		want:  []UD{{"sugar", ""}},
	},
	{
		input: plstInput{actor: UD{"tentacle", "cat.sns"}, target: "123"},
		want:  []UD{{"sugar", ""}, {"tentacle", "cat.sns"}},
	},
}

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
	postDb := &PostDb{logger, client, cacheDb}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})

	postDb.SetPost(&pqstTable[0].input)

	t.Run("Like 1", func(t *testing.T) {
		input := plstTable[0].input
		err := postDb.SetLike(input.actor, input.target)
		test.AssertNoError(t, err, "Error when set(1): %+v")

		want := plstTable[0].want
		got, err := postDb.QueryLikes(input.target)
		test.AssertNoError(t, err, "Error when query(1): %+v")
		test.AssertEqual(t, want, got)
	})

	t.Run("Like 2", func(t *testing.T) {
		input := plstTable[1].input
		err := postDb.SetLike(input.actor, input.target)
		test.AssertNoError(t, err, "Error when set(2): %+v")

		want := plstTable[1].want
		got, err := postDb.QueryLikes(input.target)
		test.AssertNoError(t, err, "Error when query(2): %+v")
		test.AssertEqual(t, want, got)
	})

	t.Run("Remove like 2", func(t *testing.T) {
		input := plstTable[1].input
		err := postDb.RemoveLike(input.actor, input.target)
		test.AssertNoError(t, err, "Error when remove(1): %+v")

		want := plstTable[0].want
		got, err := postDb.QueryLikes(input.target)
		test.AssertNoError(t, err, "Error when query(3): %+v")
		test.AssertEqual(t, want, got)
	})
}

func checkShare(client *sql.DB, lg logging.Logger, user, id string) bool {
	qs := `SELECT 1
		   FROM shares
		   WHERE "user" = $1 AND "id" = $2;`
	r := client.QueryRow(qs, user, id)
	var n int
	if e := r.Scan(&n); e != nil {
		switch e {
		case sql.ErrNoRows:
			return false
		default:
			lg.Error("Cannot query", e)
			return false
		}
	}
	return true
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
	postDb := &PostDb{logger, client, cacheDb}
	t.Cleanup(func() {
		for _, v := range plstTable {
			client.Exec(`DELETE FROM shares WHERE "user" = $1 AND "id" = $2;`, v.input.actor, v.input.target)
			redis.GetDel(defaultCtx, "user:"+v.input.actor.String())
		}
		for _, v := range pqstTable {
			client.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
			redis.GetDel(defaultCtx, "post:"+v.input.ID)
		}
	})
	d := time.Now()

	postDb.SetPost(&pqstTable[0].input)

	t.Run("Share 1", func(t *testing.T) {
		input := plstTable[0].input
		err := postDb.SetShare(input.actor, input.target, d, utils.Vsb_PUBLIC)
		test.AssertNoError(t, err, "Error when set(1): %+v")

		want := plstTable[0].want
		got, err := postDb.QueryShares(input.target)
		test.AssertNoError(t, err, "Error when query(1): %+v")
		test.AssertEqual(t, want, got)
		test.AssertEqual(t, true, checkShare(client, logger, input.actor.String(), input.target))
	})

	t.Run("Share 2", func(t *testing.T) {
		input := plstTable[1].input
		err := postDb.SetShare(input.actor, input.target, d, utils.Vsb_PUBLIC)
		test.AssertNoError(t, err, "Error when set(2): %+v")

		want := plstTable[1].want
		got, err := postDb.QueryShares(input.target)
		test.AssertNoError(t, err, "Error when query(2): %+v")
		test.AssertEqual(t, want, got)
		test.AssertEqual(t, true, checkShare(client, logger, input.actor.String(), input.target))
	})

	t.Run("Remove share 2", func(t *testing.T) {
		input := plstTable[1].input
		err := postDb.RemoveShare(input.actor, input.target)
		test.AssertNoError(t, err, "Error when remove(1): %+v")

		want := plstTable[0].want
		got, err := postDb.QueryShares(input.target)
		test.AssertNoError(t, err, "Error when query(3): %+v")
		test.AssertEqual(t, want, got)
		test.AssertEqual(t, false, checkShare(client, logger, input.actor.String(), input.target))
	})
}
