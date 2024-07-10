package models

import (
	"testing"

	"github.com/kidommoc/gustrody/internal/test"
)

var uftTable = []ForeignUser{
	{Username: UD{"aaa", "exam.ple"}, ID: "idOfAaa", Avatar: "51b8b50a9a.png", AvtUrl: "51b8b50a9a.png", Inbox: "exam.ple/inbox", Followers: "idOfAaa/followers"},
	{Username: UD{"bbb", "exam.ple"}, ID: "idOfBbb", Avatar: "f55cf38f49.png", AvtUrl: "f55cf38f49.png", Inbox: "exam.ple/inbox", Followers: "idOfBbb/followers"},
	{Username: UD{"ccc", "site.sns"}, ID: "idOfCcc", Avatar: "a2af90f964.png", AvtUrl: "a2af90f964.png", Inbox: "site.sns/inbox", Followers: "idOfCcc/followers"},
}

var uftInput = ForeignUser{
	Username: UD{"aaa", "exam.ple"}, ID: "idOfAaa", Avatar: "ee483a3654.png", AvtUrl: "ee483a3654.png", Inbox: "exam.ple/inbox", Followers: "exam.ple/Aaa/followers",
}

var uftTableI = []struct {
	input []UD
	want  []string
}{
	{[]UD{
		{"aaa", "exam.ple"}, {"bbb", "exam.ple"},
	}, []string{"exam.ple/inbox"}},
	{[]UD{
		{"aaa", "exam.ple"}, {"ccc", "site.sns"},
	}, []string{"exam.ple/inbox", "site.sns/inbox"}},
	{[]UD{
		{"aaa", "exam.ple"}, {"bbb", "exam.ple"}, {"ccc", "site.sns"},
	}, []string{"exam.ple/inbox", "site.sns/inbox"}},
}

func TestForeignUserSetAndQuery(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	userDb := &UserDb{logger, mp}
	t.Cleanup(func() {
		conn, _ := mp.Open()
		defer conn.Close()
		for _, v := range uftTable {
			conn.Exec(`DELETE FROM foreign_users WHERE "user" = $1;`, v.Username)
		}
	})

	input := uftTable[0]
	t.Run("Set", func(t *testing.T) {
		err := userDb.SetForeignUser(&input)
		test.AssertNoError(t, err, "Error when set: %+v")
	})

	t.Run("Query UD", func(t *testing.T) {
		got, err := userDb.GetForeignUserByUD(input.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		test.AssertEqual(t, input, got)
	})

	t.Run("Query ID", func(t *testing.T) {
		got, err := userDb.GetForeignUserByID(input.ID)
		test.AssertNoError(t, err, "Error when query: %+v")
		test.AssertEqual(t, input, got)
	})
}

func TestForeignUserUpdate(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	userDb := &UserDb{logger, mp}
	t.Cleanup(func() {
		conn, _ := mp.Open()
		defer conn.Close()
		for _, v := range uftTable {
			conn.Exec(`DELETE FROM foreign_users WHERE "user" = $1;`, v.Username)
		}
	})

	t.Run("Set", func(t *testing.T) {
		input := uftTable[0]
		err := userDb.SetForeignUser(&input)
		test.AssertNoError(t, err, "Error when set: %+v")
	})

	t.Run("Update", func(t *testing.T) {
		input := uftInput
		err := userDb.SetForeignUser(&input)
		test.AssertNoError(t, err, "Error when update: %+v")

		got, err := userDb.GetForeignUserByUD(input.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		test.AssertEqual(t, input, got)
	})
}

func TestForeignInboxGet(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	userDb := &UserDb{logger, mp}
	t.Cleanup(func() {
		conn, _ := mp.Open()
		defer conn.Close()
		for _, v := range uftTable {
			conn.Exec(`DELETE FROM foreign_users WHERE "user" = $1;`, v.Username)
		}
	})

	t.Run("Set", func(t *testing.T) {
		for _, input := range uftTable {
			err := userDb.SetForeignUser(&input)
			test.AssertNoError(t, err, "Error when set: %+v")
		}
	})

	t.Run("Get Inbox", func(t *testing.T) {
		for _, v := range uftTableI {
			got, err := userDb.GetInboxes(v.input)
			test.AssertNoError(t, err, "Error when get inbox: %+v")
			test.AssertEqual(t, v.want, got)
		}
	})
}
