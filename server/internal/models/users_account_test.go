package models

import (
	"testing"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var uatcfg = config.Config{
	PqUser:   "penguin",
	PqSecret: "postgres",
	RdSecret: "redis",
}

var uatTableU = []User{
	{
		Username: "aaa", Nickname: "AAA",
	},
	{
		Username: "aaa", Nickname: "AaA",
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

func TestUserSet(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := db.MainPool(&uatcfg, logger)
	userDb := &UserDb{logger, mp}

	t.Cleanup(func() {
		for _, v := range uatTableU {
			conn, _ := mp.Open()
			conn.Exec("DELETE FROM users WHERE \"username\" = $1;", v.Username)
		}
	})

	input := uatTableU[0]
	input.Keys.Pub, input.Keys.Pri = utils.NewKeyPair()
	err := userDb.SetUser(&input)
	test.AssertNoError(t, err, "Error when set: %+v")

	got, err := userDb.QueryUser(input.Username)
	test.AssertNoError(t, err, "Error when query: %+v")

	t.Logf("\ninput: %+v\ngot: %+v\n", input, got)

	_, _, err = userDb.QueryUserKeys(input.Username)
	test.AssertNoError(t, err, "Error when query keys: %+v")
}

func TestUserUpdate(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := db.MainPool(&uatcfg, logger)
	userDb := &UserDb{logger, mp}

	t.Cleanup(func() {
		for _, v := range uatTableU {
			conn, _ := mp.Open()
			conn.Exec("DELETE FROM users WHERE \"username\" = $1;", v.Username)
		}
	})

	input := uatTableU[0]
	err := userDb.SetUser(&input)
	test.AssertNoError(t, err, "Error when set: %+v")

	input = uatTableU[1]
	err = userDb.UpdateUser(&input)
	test.AssertNoError(t, err, "Error when update: %+v")

	got, err := userDb.QueryUser(input.Username)
	test.AssertNoError(t, err, "Error when query: %+v")

	t.Logf("\ninput: %+v\ngot: %+v\n", input, got)
}

func TestPreferenceUpdate(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := db.MainPool(&uatcfg, logger)
	userDb := &UserDb{logger, mp}

	t.Cleanup(func() {
		for _, v := range uatTableU {
			conn, _ := mp.Open()
			conn.Exec("DELETE FROM users WHERE \"username\" = $1;", v.Username)
		}
	})

	inputU := uatTableU[0]
	err := userDb.SetUser(&inputU)
	test.AssertNoError(t, err, "Error when set user: %+v")

	before, err := userDb.QueryUserPreferences(inputU.Username)
	test.AssertNoError(t, err, "Error when query before: %+v")

	input := uatTablePf[1]
	err = userDb.UpdateUserPreferences(inputU.Username, &input)
	test.AssertNoError(t, err, "Error when update: %+v")

	after, err := userDb.QueryUserPreferences(inputU.Username)
	test.AssertNoError(t, err, "Error when query after: %+v")

	t.Logf("\ninput: %+v\nbefore: %+v\nafter: %+v", input, before, after)
}
