package models

import (
	"testing"
	"time"

	"github.com/kidommoc/gustrody/internal/config"
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
	mp := newMockingPqPool(postcfg, logger)
	userDb := &UserDb{logger, mp}
	t.Cleanup(func() {
		for _, v := range uatTableU {
			conn, _ := mp.Open()
			defer conn.Close()
			conn.Exec(`DELETE FROM users WHERE "username" = $1;`, v.Username)
		}
	})

	input := uatTableU[0]
	t.Run("Set", func(t *testing.T) {
		input.Date = time.Now()
		input.Keys.Pub, input.Keys.Pri = utils.NewKeyPair()
		err := userDb.SetUser(&input)
		test.AssertNoError(t, err, "Error when set: %+v")
	})

	t.Run("Query", func(t *testing.T) {
		got, err := userDb.QueryUser(input.Username.Username)
		test.AssertNoError(t, err, "Error when query: %+v")

		t.Logf("\ninput: %+v\ngot: %+v\n", input, got)

		_, _, err = userDb.QueryUserKeys(input.Username.Username)
		test.AssertNoError(t, err, "Error when query keys: %+v")
	})
}

func TestUserUpdate(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	userDb := &UserDb{logger, mp}
	t.Cleanup(func() {
		for _, v := range uatTableU {
			conn, _ := mp.Open()
			defer conn.Close()
			conn.Exec(`DELETE FROM users WHERE "username" = $1;`, v.Username)
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

		got, err := userDb.QueryUser(input.Username.Username)
		test.AssertNoError(t, err, "Error when query: %+v")
		t.Logf("\ninput: %+v\ngot: %+v\n", input, got)
	})
}

func TestPreferenceUpdate(t *testing.T) {
	logger := test.NewMockingLogger(t)
	mp := newMockingPqPool(postcfg, logger)
	userDb := &UserDb{logger, mp}
	t.Cleanup(func() {
		for _, v := range uatTableU {
			conn, _ := mp.Open()
			defer conn.Close()
			conn.Exec(`DELETE FROM users WHERE "username" = $1;`, v.Username)
		}
	})

	inputU := uatTableU[0]
	input := uatTablePf[1]
	var before *Preferences
	t.Run("Set", func(t *testing.T) {
		err := userDb.SetUser(&inputU)
		test.AssertNoError(t, err, "Error when set user: %+v")
		before, err = userDb.QueryUserPreferences(inputU.Username.Username)
		test.AssertNoError(t, err, "Error when query before: %+v")
	})

	t.Run("Update", func(t *testing.T) {
		err := userDb.UpdateUserPreferences(inputU.Username.Username, &input)
		test.AssertNoError(t, err, "Error when update: %+v")
	})

	t.Run("Query", func(t *testing.T) {
		after, err := userDb.QueryUserPreferences(inputU.Username.Username)
		test.AssertNoError(t, err, "Error when query after: %+v")
		t.Logf("\ninput: %+v\nbefore: %+v\nafter: %+v", input, before, after)
	})
}
