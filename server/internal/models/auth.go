package models

import (
	"fmt"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
)

type IAuthDb interface {
	QueryPasswordOfUser(username string) (password string, err error)
	SetUserPassword(username string, password string) error
}

type AuthDb struct {
	lg   logging.Logger
	pool *_db.ConnPool[*_db.RdConn]
}

var authIns *AuthDb = nil

func AuthInstance(lg logging.Logger) *AuthDb {
	if authIns == nil {
		authIns = &AuthDb{
			lg:   lg,
			pool: _db.AuthPool(nil, nil),
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
func (db *AuthDb) QueryPasswordOfUser(username string) (password string, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.Auth] Failed to open a connection", err)
		return "", ErrDbInternal
	}
	defer conn.Close()

	passwd, err := conn.Get("pswd:" + username)
	if err != nil {
		switch err {
		case _db.ErrNotFound:
			msg := fmt.Sprintf("[Model.Auth] Cannot find user %s", username)
			logger.Error(msg, err)
			return "", ErrNotFound
		default:
			logger.Error("[Model.Auth] Db error", err)
			return "", ErrDbInternal
		}
	}
	if passwd == "" {
		return "", ErrSyntax
	}
	return passwd, nil
}

// ERRORS
//
//   - DbInternal
//   - Syntax "empty password"
func (db *AuthDb) SetUserPassword(username, password string) error {
	if password == "" {
		return ErrSyntax
	}

	logger := db.lg
	conn, e := db.pool.Open()
	if e != nil {
		logger.Error("[Model.Auth] Failed to open a connection", e)
		return ErrDbInternal
	}
	defer conn.Close()

	if e := conn.SetString("pswd:"+username, password); e != nil {
		logger.Error("[Model.Auth] Cannot set password", e)
		return ErrDbInternal
	}
	return nil
}
