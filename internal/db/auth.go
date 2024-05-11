package db

import "errors"

// should implemented with Redis

var loginDb = make(map[string]string)
var sessionDb = make(map[string]string)

func initAuthDb() {
	loginDb["u1"] = "penguin"
	loginDb["u2"] = "penguin"
	loginDb["u3"] = "penguin"
	loginDb["u4"] = "penguin"
}

func QueryPasswordOfUser(username string) (password string, err error) {
	passwd := loginDb[username]
	if passwd == "" {
		return "", errors.New("user not found")
	}
	return passwd, nil
}

func QueryUserOfSession(session string) (username string, err error) {
	username = sessionDb[session]
	if username == "" {
		return "", errors.New("user not found")
	}
	return username, nil
}

func SetSession(session string, username string) {
	sessionDb[session] = username
}
