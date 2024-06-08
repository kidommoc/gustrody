package db

import (
	"slices"

	"github.com/kidommoc/gustrody/internal/utils"
)

type User struct {
	Username string
	Nickname string
	Summary  string
}

type follow struct {
	From string
	To   string
}

var infoDb = make(map[string]*User)
var followsDb = make([]*follow, 0, 100)

func initUserDb() {
	infoDb["u1"] = &User{
		Username: "u1",
		Nickname: "User1",
		Summary:  "i'm u1",
	}
	infoDb["u2"] = &User{
		Username: "u2",
		Nickname: "User2",
		Summary:  "i'm u2",
	}
	infoDb["u3"] = &User{
		Username: "u3",
		Nickname: "User3",
		Summary:  "i'm u3",
	}
	infoDb["u4"] = &User{
		Username: "u4",
		Nickname: "User4",
		Summary:  "i'm u4",
	}

	SetFollow("u1", "u3")
	SetFollow("u2", "u1")
	SetFollow("u2", "u3")
	SetFollow("u3", "u2")
}

func IsUserExist(username string) bool {
	if infoDb[username] != nil {
		return true
	} else {
		return false
	}
}

func checkFollow(from string, to string) int {
	for i, f := range followsDb {
		if f.From == from && f.To == to {
			return i
		}
	}
	return -1
}

func QueryUser(username string) (user User, err utils.Err) {
	if !IsUserExist(username) {
		return user, utils.NewErr(ErrNotFound, "user")
	}
	return *infoDb[username], nil
}

func QueryUserFollows(username string) (count uint, err utils.Err) {
	if !IsUserExist(username) {
		return 0, utils.NewErr(ErrNotFound, "user")
	}
	count = 0
	for _, f := range followsDb {
		if f.From == username {
			count = count + 1
		}
	}
	return count, nil
}

func QueryUserFollowed(username string) (count uint, err utils.Err) {
	if !IsUserExist(username) {
		return 0, utils.NewErr(ErrNotFound, "user")
	}
	count = 0
	for _, f := range followsDb {
		if f.To == username {
			count = count + 1
		}
	}
	return count, nil
}

func QueryUserFollowings(username string) (list []*User, err utils.Err) {
	list = make([]*User, 0)
	if !IsUserExist(username) {
		return list, utils.NewErr(ErrNotFound, "user")
	}
	for _, f := range followsDb {
		if f.From == username {
			list = append(list, infoDb[f.To])
		}
	}
	return list, nil
}

func QueryUserFollowers(username string) (list []*User, err utils.Err) {
	list = make([]*User, 0)
	if !IsUserExist(username) {
		return list, utils.NewErr(ErrNotFound, "user")
	}
	for _, f := range followsDb {
		if f.To == username {
			list = append(list, infoDb[f.From])
		}
	}
	return list, nil
}

func SetFollow(from string, to string) utils.Err {
	if !IsUserExist(from) {
		return utils.NewErr(ErrNotFound, "from")
	}
	if !IsUserExist(to) {
		return utils.NewErr(ErrNotFound, "to")
	}
	if index := checkFollow(from, to); index == -1 {
		followsDb = append(followsDb, &follow{From: from, To: to})
	}
	return nil
}

func RemoveFollow(from string, to string) utils.Err {
	if !IsUserExist(from) {
		return utils.NewErr(ErrNotFound, "from")
	}
	if !IsUserExist(to) {
		return utils.NewErr(ErrNotFound, "to")
	}
	if index := checkFollow(from, to); index != -1 {
		followsDb = slices.Delete(followsDb, index, index+1)
	}
	return nil
}
