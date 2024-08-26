package models

import (
	"database/sql"
	"encoding/json"
)

type IUserInfo interface {
	IsUserExist(username string) bool
	QueryUser(username string) (user User, err error)
	// uses: User.Username, User.Nickname, User.Summary, User.Avatar
	UpdateUser(user *User) error
}

func (db *UserDb) IsUserExist(username string) bool {
	logger := db.lg

	qs := `SELECT 1 FROM users WHERE "username" = $1;`
	r := db.client.QueryRow(qs, username)
	var n int
	if err := r.Scan(&n); err != nil { // can't understand why i MUST scan to check whether result is empty. silly design
		switch err {
		case sql.ErrNoRows:
		default:
			logger.Error("[Model.UserInfo] Cannot query", err)
		}
		return false
	}
	return true
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UserDb) QueryUser(username string) (user User, err error) {
	logger := db.lg
	if username == "" {
		return user, ErrFormat
	}

	// try query cache
	cacheKey := "user:" + username
	ss, err := db.cache.QueryString([]string{cacheKey})
	if err == nil {
		if ss[cacheKey] != "" {
			// cache hit
			err = json.Unmarshal([]byte(ss[cacheKey]), &user)
			if err == nil {
				// use cache
				return user, nil
			} else {
				// cache corrupted. clear cache
				db.cache.RemoveString("user:" + username)
			}
		}
	} else {
		logger.Warning("[Models.QueryUser] Failed to query cache.", "error", err)
	}

	qs := ` SELECT "username", "nickname", "summary", "avatar"
			FROM users
			WHERE "username" = $1;`
	r := db.client.QueryRow(qs, username)

	user = User{}
	if e := r.Scan(
		&user.Username.Username, &user.Nickname, &user.Summary, &user.Avatar,
	); e != nil {
		switch e {
		case sql.ErrNoRows:
			return user, ErrNotFound
		default:
			logger.Error("[Model.UserInfo] Cannot scan row", e)
			return user, ErrDbInternal
		}
	}

	// cache
	s, _ := json.Marshal(user)
	go db.cache.SetString(cacheKey, string(s), true)

	return user, nil
}

func (db *UserDb) UpdateUser(user *User) error {
	logger := db.lg
	qs := ` UPDATE users
			SET "nickname" = $2, "summary" = $3, "avatar" = $4
			WHERE "username" = $1;`
	r, err := sqlExec(db.client.Exec(qs, user.Username.Username,
		user.Nickname, user.Summary, user.Avatar,
	))
	if err != nil {
		logger.Error("[Model.UserInfo] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}

	// cache
	summary := user.Summary
	if summary == "" {
		summary = stringClearFlag
	}
	go db.cache.UpdateJson("user:"+user.Username.String(), User{
		Nickname: user.Nickname, Summary: summary, Avatar: user.Avatar,
	})
	return nil
}
