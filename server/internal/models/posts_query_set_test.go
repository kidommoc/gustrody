package models

import (
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var pqstcfg = config.Config{
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
		pqstInput{Post: Post{
			ID: "123", User: "foo", Replying: "",
			Content: "bar",
		}, Imgs: []Img{
			{Url: "1.png"},
			{Url: "2.jpeg", Alt: "alt text"},
		}},
		Post{ID: "123", Url: "/123", User: "foo",
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "bar",
			Media: Array[Img, *Img]{data: []Img{
				{Url: "1.png"},
				{Url: "2.jpeg", Alt: "alt text"},
			}}},
	},
	{
		pqstInput{Post: Post{ID: "123", Content: "foobar"}, Imgs: []Img{
			{Url: "1.png", Alt: "alt text"},
			{Url: "2.jpeg"},
		}},
		Post{ID: "123", Url: "/123", User: "foo",
			Replying: "", Vsb: utils.Vsb_PUBLIC, Content: "foobar",
			Media: Array[Img, *Img]{data: []Img{
				{Url: "1.png", Alt: "alt text"},
				{Url: "2.jpeg"},
			}}},
	},
}

func TestPostSetAndQuery(t *testing.T) {
	d := time.Now().UTC()

	logger := test.NewMockingLogger(t)
	mp := db.MainPool(&pqstcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}

	t.Cleanup(func() {
		for _, v := range pqstTable {
			conn, _ := mp.Open()
			conn.Exec("DELETE FROM posts WHERE \"id\" = $1;", v.input.ID)
		}
	})

	input := pqstTable[0].input
	want := pqstTable[0].want
	want.Date = d
	err := postDb.SetPost(&input.Post, input.Imgs)
	test.AssertNoError(t, err, "Error when set: %+v")

	got, err := postDb.QueryPostByID("123")
	test.AssertNoError(t, err, "Error when query: %+v")

	t.Logf("got: %+v", got)
	t.Logf("want: %+v", pqstTable[0].want)
}

func TestPostUpdate(t *testing.T) {
	d := time.Now().UTC()
	d1 := d.Add(time.Hour)

	logger := test.NewMockingLogger(t)
	mp := db.MainPool(&pqstcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}

	t.Cleanup(func() {
		for _, v := range pqstTable {
			conn, _ := mp.Open()
			conn.Exec("DELETE FROM posts WHERE \"id\" = $1;", v.input.ID)
		}
	})

	input := pqstTable[0].input
	err := postDb.SetPost(&input.Post, input.Imgs)
	test.AssertNoError(t, err, "Error when set: %+v")

	input = pqstTable[1].input
	want := pqstTable[1].want
	want.Date = d1
	err = postDb.UpdatePost(&input.Post, input.Imgs)
	test.AssertNoError(t, err, "Error when update: %+v")

	got, err := postDb.QueryPostByID("123")
	test.AssertNoError(t, err, "Error when query: %+v")

	t.Logf("got: %+v", got)
	t.Logf("want: %+v", want)
}

func TestPostRemove(t *testing.T) {
	d := time.Now().UTC()

	logger := test.NewMockingLogger(t)
	mp := db.MainPool(&pqstcfg, logger)
	postDb := &PostDb{lg: logger, pool: mp}

	t.Cleanup(func() {
		for _, v := range pqstTable {
			conn, _ := mp.Open()
			conn.Exec("DELETE FROM posts WHERE \"id\" = $1;", v.input.ID)
		}
	})

	input := pqstTable[0].input
	want := pqstTable[0].want
	want.Date = d
	err := postDb.SetPost(&input.Post, input.Imgs)
	test.AssertNoError(t, err, "Error when set: %+v")

	if !postDb.IsPostExist(input.ID) {
		t.Errorf("Wrong result of IsPostExist: false")
	}

	err = postDb.RemovePost(input.ID)
	test.AssertNoError(t, err, "Error when remove: %+v")

	if postDb.IsPostExist(input.ID) {
		t.Errorf("Wrong result of IsPostExist: true")
	}
}
