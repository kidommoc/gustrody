package database

import (
	"slices"

	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type User struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Summary  string `json:"summary"`
}

type follow struct {
	From string
	To   string
}

// database

type IUsersDb interface {
	IsUserExist(username string) bool
	checkFollow(from string, to string) int
	QueryUser(username string) (user User, err utils.Err)
	QueryUserFollows(username string) (count uint, err utils.Err)
	QueryUserFollowed(username string) (count uint, err utils.Err)
	QueryUserFollowings(username string) (list []*User, err utils.Err)
	QueryUserFollowers(username string) (list []*User, err utils.Err)
	SetFollow(from string, to string) utils.Err
	RemoveFollow(from string, to string) utils.Err
}

// should implemented with Postgre
type UsersDb struct {
	infoDb    map[string]*User
	followsDb []*follow
}

var userIns *UsersDb = nil

func UserInstance() *UsersDb {
	if userIns == nil {
		userIns = &UsersDb{
			infoDb:    make(map[string]*User),
			followsDb: make([]*follow, 0, 100),
		}
	}
	return userIns
}

// functions

func initUserDb() {
	db := UserInstance()
	db.infoDb["u1"] = &User{
		Username: "u1",
		Nickname: "User1",
		Summary:  "i'm u1",
	}
	db.infoDb["u2"] = &User{
		Username: "u2",
		Nickname: "User2",
		Summary:  "i'm u2",
	}
	db.infoDb["u3"] = &User{
		Username: "u3",
		Nickname: "User3",
		Summary:  "i'm u3",
	}

	db.SetFollow("u1", "u3")
	db.SetFollow("u2", "u1")
	db.SetFollow("u2", "u3")
	db.SetFollow("u3", "u2")
}

func (db *UsersDb) IsUserExist(username string) bool {
	if db.infoDb[username] != nil {
		return true
	} else {
		return false
	}
}

func (db *UsersDb) checkFollow(from string, to string) int {
	for i, f := range db.followsDb {
		if f.From == from && f.To == to {
			return i
		}
	}
	return -1
}

func (db *UsersDb) QueryUser(username string) (user User, err utils.Err) {
	if !db.IsUserExist(username) {
		return user, utils.NewErr(ErrNotFound, "user")
	}
	return *db.infoDb[username], nil
}

func (db *UsersDb) QueryUserFollows(username string) (count uint, err utils.Err) {
	if !db.IsUserExist(username) {
		return 0, utils.NewErr(ErrNotFound, "user")
	}
	count = 0
	for _, f := range db.followsDb {
		if f.From == username {
			count = count + 1
		}
	}
	return count, nil
}

func (db *UsersDb) QueryUserFollowed(username string) (count uint, err utils.Err) {
	if !db.IsUserExist(username) {
		return 0, utils.NewErr(ErrNotFound, "user")
	}
	count = 0
	for _, f := range db.followsDb {
		if f.To == username {
			count = count + 1
		}
	}
	return count, nil
}

func (db *UsersDb) QueryUserFollowings(username string) (list []*User, err utils.Err) {
	list = make([]*User, 0)
	if !db.IsUserExist(username) {
		return list, utils.NewErr(ErrNotFound, "user")
	}
	for _, f := range db.followsDb {
		if f.From == username {
			list = append(list, db.infoDb[f.To])
		}
	}
	return list, nil
}

func (db *UsersDb) QueryUserFollowers(username string) (list []*User, err utils.Err) {
	list = make([]*User, 0)
	if !db.IsUserExist(username) {
		return list, utils.NewErr(ErrNotFound, "user")
	}
	for _, f := range db.followsDb {
		if f.To == username {
			list = append(list, db.infoDb[f.From])
		}
	}
	return list, nil
}

func (db *UsersDb) SetFollow(from string, to string) utils.Err {
	if !db.IsUserExist(from) {
		return utils.NewErr(ErrNotFound, "from")
	}
	if !db.IsUserExist(to) {
		return utils.NewErr(ErrNotFound, "to")
	}
	if index := db.checkFollow(from, to); index == -1 {
		db.followsDb = append(db.followsDb, &follow{From: from, To: to})
	}
	return nil
}

func (db *UsersDb) RemoveFollow(from string, to string) utils.Err {
	if !db.IsUserExist(from) {
		return utils.NewErr(ErrNotFound, "from")
	}
	if !db.IsUserExist(to) {
		return utils.NewErr(ErrNotFound, "to")
	}
	if index := db.checkFollow(from, to); index != -1 {
		db.followsDb = slices.Delete(db.followsDb, index, index+1)
	}
	return nil
}
