package db

import (
	"github.com/kidommoc/gustrody/internal/utils"
)

// should implemented with Redis

var loginDb = make(map[string]string)
var sessionDb = make(map[string]string)

func initAuthDb() {
	loginDb["u1"] = "penguin"
	loginDb["u2"] = "penguin"
	loginDb["u3"] = "penguin"
	loginDb["u4"] = "penguin"
}

func QueryPasswordOfUser(username string) (password string, err utils.Err) {
	passwd := loginDb[username]
	if passwd == "" {
		return "", utils.NewErr(ErrNotFound, "user")
	}
	return passwd, nil
}

func QueryUserOfSession(session string) (username string, err utils.Err) {
	username = sessionDb[session]
	if username == "" {
		return "", utils.NewErr(ErrNotFound, "user")
	}
	return username, nil
}

func SetSession(session string, username string) {
	sessionDb[session] = username
}
