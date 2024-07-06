package models

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/db"
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
	mp := newMockingPqPool(postcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}
	t.Cleanup(func() {
		conn, _ := mp.Open()
		defer conn.Close()
		for _, v := range pqstTable {
			conn.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
		}
	})

	inputP := pqstTable[0].input
	postDb.SetPost(&inputP.Post, inputP.Imgs)

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

func checkShare(conn db.PqConn, lg logging.Logger, user, id string) bool {
	qs := `SELECT 1
		   FROM shares
		   WHERE "user" = $1 AND "id" = $2;`
	r := conn.QueryOne(qs, user, id)
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
	mp := newMockingPqPool(postcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}
	t.Cleanup(func() {
		conn, _ := mp.Open()
		defer conn.Close()
		for _, v := range plstTable {
			conn.Exec(`DELETE FROM shares WHERE "user" = $1 AND "id" = $2;`, v.input.actor, v.input.target)
		}
		for _, v := range pqstTable {
			conn.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
		}
	})
	conn, _ := mp.Open()
	defer conn.Close()
	d := time.Now()

	inputP := pqstTable[0].input
	postDb.SetPost(&inputP.Post, inputP.Imgs)

	t.Run("Share 1", func(t *testing.T) {
		input := plstTable[0].input
		err := postDb.SetShare(input.actor, input.target, d, utils.Vsb_PUBLIC)
		test.AssertNoError(t, err, "Error when set(1): %+v")

		want := plstTable[0].want
		got, err := postDb.QueryShares(input.target)
		test.AssertNoError(t, err, "Error when query(1): %+v")
		test.AssertEqual(t, want, got)
		test.AssertEqual(t, true, checkShare(conn, logger, input.actor.String(), input.target))
	})

	t.Run("Share 2", func(t *testing.T) {
		input := plstTable[1].input
		err := postDb.SetShare(input.actor, input.target, d, utils.Vsb_PUBLIC)
		test.AssertNoError(t, err, "Error when set(2): %+v")

		want := plstTable[1].want
		got, err := postDb.QueryShares(input.target)
		test.AssertNoError(t, err, "Error when query(2): %+v")
		test.AssertEqual(t, want, got)
		test.AssertEqual(t, true, checkShare(conn, logger, input.actor.String(), input.target))
	})

	t.Run("Remove share 2", func(t *testing.T) {
		input := plstTable[1].input
		err := postDb.RemoveShare(input.actor, input.target)
		test.AssertNoError(t, err, "Error when remove(1): %+v")

		want := plstTable[0].want
		got, err := postDb.QueryShares(input.target)
		test.AssertNoError(t, err, "Error when query(3): %+v")
		test.AssertEqual(t, want, got)
		test.AssertEqual(t, false, checkShare(conn, logger, input.actor.String(), input.target))
	})
}
