package models

import (
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var postcfg = config.Config{
	PqUser:   "penguin",
	PqSecret: "postgres",
	RdSecret: "redis",
}

type pqstInput struct {
	Post
	Imgs []Img
}

var pqstTable = []struct {
	input pqstInput
	want  Post
}{
	{
		input: pqstInput{Post: Post{
			ID: "123", User: UD{"foo", "bar.sns"}, Replying: "",
			Vsb: utils.Vsb_PUBLIC, Content: "example",
		}, Imgs: []Img{
			{Type: "image/png", Url: "1.png"},
			{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
		}},
		want: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "bar",
			Media: Array[Img, *Img]{data: []Img{
				{Type: "image/png", Url: "1.png"},
				{Type: "image/jpeg", Url: "2.jpeg", Alt: "alt text"},
			}}},
	},
	{
		input: pqstInput{Post: Post{ID: "123", Content: "sample"}, Imgs: []Img{
			{Type: "image/png", Url: "1.png", Alt: "alt text"},
			{Type: "image/jpeg", Url: "2.jpeg"},
		}},
		want: Post{ID: "123", User: UD{"foo", "bar.sns"},
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "sample",
			Media: Array[Img, *Img]{data: []Img{
				{Type: "image/png", Url: "1.png", Alt: "alt text"},
				{Type: "image/jpeg", Url: "2.jpeg"},
			}}},
	},
}

func TestPostSetAndQuery(t *testing.T) {
	d := time.Now().UTC()

	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			conn, _ := mp.Open()
			defer conn.Close()
			conn.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := pqstTable[0].input
		want := pqstTable[0].want
		want.Date = d
		err := postDb.SetPost(&input.Post, input.Imgs)
		test.AssertNoError(t, err, "Error when set: %+v")
	})

	t.Run("Query", func(t *testing.T) {
		got, err := postDb.QueryPostByID("123")
		test.AssertNoError(t, err, "Error when query: %+v")
		t.Logf("got: %+v", got)
		t.Logf("want: %+v", pqstTable[0].want)
	})
}

func TestPostUpdate(t *testing.T) {
	d := time.Now().UTC()
	d1 := d.Add(time.Hour)

	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			conn, _ := mp.Open()
			defer conn.Close()
			conn.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := pqstTable[0].input
		err := postDb.SetPost(&input.Post, input.Imgs)
		test.AssertNoError(t, err, "Error when set: %+v")
	})

	t.Run("Update", func(t *testing.T) {
		input := pqstTable[1].input
		want := pqstTable[1].want
		want.Date = d1
		err := postDb.UpdatePost(&input.Post, input.Imgs)
		test.AssertNoError(t, err, "Error when update: %+v")

		got, err := postDb.QueryPostByID("123")
		test.AssertNoError(t, err, "Error when query: %+v")

		t.Logf("got: %+v", got)
		t.Logf("want: %+v", want)
	})
}

func TestPostRemove(t *testing.T) {
	d := time.Now().UTC()

	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}
	t.Cleanup(func() {
		for _, v := range pqstTable {
			conn, _ := mp.Open()
			defer conn.Close()
			conn.Exec(`DELETE FROM posts WHERE "id" = $1;`, v.input.ID)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := pqstTable[0].input
		want := pqstTable[0].want
		want.Date = d
		err := postDb.SetPost(&input.Post, input.Imgs)
		test.AssertNoError(t, err, "Error when set: %+v")
		test.AssertEqual(t, true, postDb.IsPostExist(input.ID))
	})

	t.Run("Remove", func(t *testing.T) {
		input := pqstTable[0].input
		err := postDb.RemovePost(input.ID)
		test.AssertNoError(t, err, "Error when remove: %+v")
		test.AssertEqual(t, false, postDb.IsPostExist(input.ID))
	})
}
