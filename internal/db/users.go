package db

import (
	"errors"
)

type User struct {
	Username string
	Bio      string
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
		Bio:      "i'm u1",
	}
	infoDb["u2"] = &User{
		Username: "u2",
		Bio:      "i'm u2",
	}
	infoDb["u3"] = &User{
		Username: "u3",
		Bio:      "i'm u3",
	}
	infoDb["u4"] = &User{
		Username: "u4",
		Bio:      "i'm u4",
	}

	SetFollow("u1", "u3")
	SetFollow("u2", "u1")
	SetFollow("u2", "u3")
	SetFollow("u3", "u2")
}

func checkUser(username string) bool {
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

func QueryUser(username string) (user User, err error) {
	if !checkUser(username) {
		return user, errors.New("username not found")
	}
	return *infoDb[username], nil
}

func QueryUserFollows(username string) (count uint, err error) {
	if !checkUser(username) {
		return 0, errors.New("user not found")
	}
	count = 0
	for _, f := range followsDb {
		if f.From == username {
			count = count + 1
		}
	}
	return count, nil
}

func QueryUserFollowed(username string) (count uint, err error) {
	if !checkUser(username) {
		return 0, errors.New("username not found")
	}
	count = 0
	for _, f := range followsDb {
		if f.To == username {
			count = count + 1
		}
	}
	return count, nil
}

func QueryUserFollowings(username string) (list []*User, err error) {
	list = make([]*User, 0)
	if !checkUser(username) {
		return list, errors.New("username not found")
	}
	for _, f := range followsDb {
		if f.From == username {
			list = append(list, infoDb[f.To])
		}
	}
	return list, nil
}

func QueryUserFollowers(username string) (list []*User, err error) {
	list = make([]*User, 0)
	if !checkUser(username) {
		return list, errors.New("username not found")
	}
	for _, f := range followsDb {
		if f.To == username {
			list = append(list, infoDb[f.From])
		}
	}
	return list, nil
}

func SetFollow(from string, to string) error {
	if from == to {
		return errors.New("from and to are same")
	}
	if !checkUser(from) {
		return errors.New("from not found")
	}
	if !checkUser(to) {
		return errors.New("to not found")
	}
	if index := checkFollow(from, to); index == -1 {
		followsDb = append(followsDb, &follow{From: from, To: to})
	}
	return nil
}

func UnsetFollow(from string, to string) error {
	if from == to {
		return errors.New("from and to are same")
	}
	if !checkUser(from) {
		return errors.New("from not found")
	}
	if !checkUser(to) {
		return errors.New("to not found")
	}
	if index := checkFollow(from, to); index != -1 {
		followsDb[index] = followsDb[len(followsDb)-1]
		followsDb = followsDb[:len(followsDb)-1]
	}
	return nil
}
