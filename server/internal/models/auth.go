package models

import (
	"fmt"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/redis/go-redis/v9"
)

type IAuthDb interface {
	QueryPasswordOfUser(username string) (password string, err error)
	SetUserPassword(username string, password string) error
}

type AuthDb struct {
	lg     logging.Logger
	client *redis.Client
}

var authIns *AuthDb = nil

func AuthInstance(lg logging.Logger, cl *redis.Client) *AuthDb {
	if authIns == nil {
		authIns = &AuthDb{
			lg:     lg,
			client: cl,
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

	passwd, err := db.client.Get(defaultCtx, "pswd:"+username).Result()
	switch err {
	case redis.Nil:
		msg := fmt.Sprintf("[Model.Auth] Cannot find user %s", username)
		logger.Error(msg, err)
		return "", ErrNotFound
	case nil:
	default:
		logger.Error("[Model.Auth] Db error", err)
		return "", ErrDbInternal
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

	if _, err := db.client.Set(defaultCtx, "pswd:"+username, password, 0).Result(); err != nil {
		logger.Error("[Model.Auth] Cannot set password", err)
		return ErrDbInternal
	}
	return nil
}
