package users

import (
	"testing"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/test"
	"github.com/kidommoc/gustrody/internal/utils"
)

var atcfg = config.Config{
	Site: "account.test.sns",
}

var us = []struct {
	Username string
	Password string
	Nickname string
	Summary  string
}{
	{"a", "aaa", "A", "abcdef"},
}

func TestAccountRegister(t *testing.T) {
	logger := test.NewMockingLogger(t)
	udb := newUdb(t)
	mAccDb := newMockingAccountDb(udb)
	mAthDb := newMockingAuthDb()
	dbs := UserDbs{
		Account: mAccDb,
		Auth:    mAthDb,
	}
	service := NewService(dbs, atcfg, logger)

	err := service.Register(us[0].Username, us[0].Nickname, us[0].Password)
	test.AssertNoError(t, err)

	wantUser := models.User{Username: us[0].Username, Nickname: us[0].Nickname}
	gotUser := udb.data[us[0].Username]
	t.Logf("\nPublic key:\n%s\nPrivate key:\n%s",
		gotUser.Keys.Pub, gotUser.Keys.Pri,
	)
	gotUser.Keys = models.KeyPair{}
	test.AssertEqual(t, wantUser, *gotUser)

	gotPwd := mAthDb.data[us[0].Username]
	wantPwd := string(utils.SHA256Hash(us[0].Password))
	test.AssertEqual(t, wantPwd, gotPwd)
}

func TestAccountUpdatePassword(t *testing.T) {
	logger := test.NewMockingLogger(t)
	udb := newUdb(t)
	mInfDb := newMockingInfoDb(udb)
	mAthDb := newMockingAuthDb()
	dbs := UserDbs{
		Info: mInfDb,
		Auth: mAthDb,
	}
	service := NewService(dbs, atcfg, logger)
	username := us[0].Username
	pwd := us[0].Password

	pwd = string(utils.SHA256Hash(pwd))
	mAthDb.SetUserPassword(username, pwd)

	want := "abc"
	err := service.UpdatePassword(username, want)
	test.AssertNoError(t, err)

	got := mAthDb.data[username]
	want = string(utils.SHA256Hash(want))
	test.AssertEqual(t, want, got)
}

/*
func TestAccountUpdateProfile(t *testing.T) {
	logger := test.NewMockingLogger(t)
	udb := newUdb(t)
	mAccDb := newMockingAccountDb(udb)
	mInfDb := newMockingInfoDb(udb)
	dbs := UserDbs{
		Account: mAccDb,
		Info:    mInfDb,
	}
	service := NewService(dbs, atcfg, logger)
	username := us[0].Username
}
*/

func TestAccountPreferences(t *testing.T) {
	logger := test.NewMockingLogger(t)

	udb := newUdb(t)
	mAccDb := newMockingAccountDb(udb)
	dbs := UserDbs{
		Account: mAccDb,
	}
	service := NewService(dbs, atcfg, logger)
	username := us[0].Username

	ps := utils.Vsb_PUBLIC.String()
	fs := utils.Vsb_FOLLOWER.String()
	pfrs := []struct {
		m models.Preferences
		g Preferences
		i PreferenceBody
		w Preferences
	}{
		{
			m: models.Preferences{PostVsb: utils.Vsb_PUBLIC.String(), ShareVsb: utils.Vsb_PUBLIC.String()},
			g: Preferences{PostVsb: utils.Vsb_PUBLIC, ShareVsb: utils.Vsb_PUBLIC},
			i: PreferenceBody{PostVsb: nil, ShareVsb: nil},
			w: Preferences{PostVsb: utils.Vsb_PUBLIC, ShareVsb: utils.Vsb_PUBLIC},
		},
		{
			m: models.Preferences{PostVsb: utils.Vsb_PUBLIC.String(), ShareVsb: utils.Vsb_FOLLOWER.String()},
			g: Preferences{PostVsb: utils.Vsb_PUBLIC, ShareVsb: utils.Vsb_FOLLOWER},
			i: PreferenceBody{PostVsb: &fs, ShareVsb: nil},
			w: Preferences{PostVsb: utils.Vsb_FOLLOWER, ShareVsb: utils.Vsb_FOLLOWER},
		},
		{
			m: models.Preferences{PostVsb: utils.Vsb_FOLLOWER.String(), ShareVsb: utils.Vsb_PUBLIC.String()},
			g: Preferences{PostVsb: utils.Vsb_FOLLOWER, ShareVsb: utils.Vsb_PUBLIC},
			i: PreferenceBody{PostVsb: &ps, ShareVsb: &fs},
			w: Preferences{PostVsb: utils.Vsb_PUBLIC, ShareVsb: utils.Vsb_FOLLOWER},
		},
	}

	for _, v := range pfrs {
		udb.data[username] = &models.User{Preferences: v.m}

		got, err := service.GetPreferences(username)
		test.AssertNoError(t, err)
		test.AssertEqual(t, v.g, got)

		err = service.UpdatePreferences(username, &v.i)
		test.AssertNoError(t, err)
		got, err = service.GetPreferences(username)
		test.AssertNoError(t, err)
		test.AssertEqual(t, v.w, got)
	}
}
