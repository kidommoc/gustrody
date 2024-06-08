package database

import (
	"github.com/kidommoc/gustrody/internal/utils"
)

type IAuthDb interface {
	QueryPasswordOfUser(username string) (password string, err utils.Err)
	QueryUserOfSession(session string) (username string, err utils.Err)
	SetSession(session string, username string) utils.Err
}

// should implemented with Redis
type AuthDb struct {
	loginDb   map[string]string
	sessionDb map[string]string
}

var authIns *AuthDb = nil

func AuthInstance() *AuthDb {
	if authIns == nil {
		authIns = &AuthDb{
			loginDb:   make(map[string]string),
			sessionDb: make(map[string]string),
		}
	}
	return authIns
}

func initAuthDb() {
	db := AuthInstance()
	db.loginDb["u1"] = "penguin"
	db.loginDb["u2"] = "penguin"
	db.loginDb["u3"] = "penguin"
}

func (db *AuthDb) QueryPasswordOfUser(username string) (password string, err utils.Err) {
	passwd := db.loginDb[username]
	if passwd == "" {
		return "", utils.NewErr(ErrNotFound, "user")
	}
	return passwd, nil
}

func (db *AuthDb) QueryUserOfSession(session string) (username string, err utils.Err) {
	username = db.sessionDb[session]
	if username == "" {
		return "", utils.NewErr(ErrNotFound, "user")
	}
	return username, nil
}

func (db *AuthDb) SetSession(session string, username string) utils.Err {
	db.sessionDb[session] = username
	return nil
}
