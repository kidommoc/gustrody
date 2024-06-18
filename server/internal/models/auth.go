package models

import (
	"fmt"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

type IAuthDb interface {
	QueryPasswordOfUser(username string) (password string, err utils.Error)
	SetUserPassword(username string, password string) utils.Error
}

type AuthDb struct {
	pool *_db.ConnPool[*_db.RdConn]
}

var authIns *AuthDb = nil

func AuthInstance() *AuthDb {
	if authIns == nil {
		authIns = &AuthDb{
			pool: _db.AuthPool(),
		}
	}
	return authIns
}

// functions

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
//   - Syntax "empty password"
func (db *AuthDb) QueryPasswordOfUser(username string) (password string, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Auth] Failed to open a connection", err)
		return "", newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	passwd, err := conn.Get("pswd:" + username)
	if err != nil {
		switch err.Code() {
		case _db.ErrNotFound:
			msg := fmt.Sprintf("[Model.Auth] Cannot find user %s", username)
			logger.Error(msg, err)
			return "", newErr(ErrNotFound, "user")
		default:
			return "", newErr(ErrDbInternal, err.Error())
		}
	}
	if passwd == "" {
		return "", newErr(ErrSyntax, "empty password")
	}
	return passwd, nil
}

// ERRORS
//
//   - DbInternal
//   - Syntax "empty password"
func (db *AuthDb) SetUserPassword(username string, password string) utils.Error {
	if password == "" {
		return newErr(ErrSyntax, "empty password")
	}

	logger := logging.Get()
	conn, e := db.pool.Open()
	if e != nil {
		logger.Error("[Model.Auth] Failed to open a connection", e)
		return newErr(ErrDbInternal, e.Error())
	}
	defer conn.Close()

	if e := conn.SetString("pswd:"+username, password); e != nil {
		logger.Error("[Model.Auth] Cannot set password", e)
		return newErr(ErrDbInternal, e.Error())
	}
	return nil
}
