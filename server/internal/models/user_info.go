package models

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/lib/pq"
)

type IUserInfo interface {
	IsUserExist(username string) bool
	QueryUserByUD(username UD) (user User, err error)
	QueryUserByID(id string) (user User, err error)
	// uses: User.Username, User.Nickname, User.Summary, User.Avatar
	UpdateUser(user *User) error
	QueryInboxes(usernames []UD) (inboxes []string, err error)
}

func isUserExist(client ISqlClient, username string) (bool, error) {
	if client == nil {
		return false, ErrFormat
	}
	const qs = `SELECT 1 FROM "users" WHERE "username" = $1;`
	r, err := client.Query(qs, username)
	if err != nil {
		return false, err
	}
	defer r.Close()
	for r.Next() {
		return true, nil
	}
	return false, nil
}

func (db *UserDb) IsUserExist(username string) bool {
	const loc = "Models.User.Exist"
	r, err := isUserExist(db.client, username)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query %s", username), err)
		return false
	}
	return r
}

func cacheUser(cache *CacheDb, u User) error {
	if cache == nil {
		return fmt.Errorf("[models.user.cache] cacheDb is nil")
	}
	u.Type = userJsonType
	s, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return cache.SetString("user:"+u.Username.String(), string(s), true)
}

func (db *UserDb) QueryUserByUD(username UD) (*User, error) {
	const loc = "Models.User.QueryByUD"
	if username.Username == "" {
		return nil, ErrFormat
	}

	var user User
	// try query cache
	cacheKey := "user:" + username.String()
	ss, err := db.cache.QueryString([]string{cacheKey})
	if err == nil {
		if ss[cacheKey] != "" {
			// cache hit
			err = json.Unmarshal([]byte(ss[cacheKey]), &user)
			if err == nil {
				// use cache
				user.Type = ""
				return &user, nil
			} else {
				// cache corrupted. clear cache
				db.cache.RemoveString("user:" + username.String())
			}
		}
	} else {
		db.lg.Warning(fmt.Sprintf("[%s] Failed to query cache.", loc), "error", err)
	}

	const qs = `SELECT "username", "id", "nickname", "summary", "avatar", "lock"
				FROM users WHERE "username" = $1;`
	r := db.client.QueryRow(qs, username)

	if err := r.Scan(
		&user.Username, &user.ID, &user.Nickname,
		&user.Summary, &user.Avatar, &user.Lock,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			logging.Cannot(db.lg, loc, fmt.Sprintf("scan result of ud-query %s", username), err)
			return nil, ErrDbInternal
		}
	}

	// cache
	go cacheUser(db.cache, user)

	return &user, nil
}

func (db *UserDb) QueryUserByID(id string) (*User, error) {
	const loc = "Models.User.QueryByUD"
	if id == "" {
		return nil, ErrFormat
	}

	const qs = `SELECT "username", "id", "nickname", "summary", "avatar", "lock"
				FROM users WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)

	var user User
	if err := r.Scan(
		&user.Username, &user.ID, &user.Nickname,
		&user.Summary, &user.Avatar, &user.Lock,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			logging.Cannot(db.lg, loc, fmt.Sprintf("scan result of id-query %s", id), err)
			return nil, ErrDbInternal
		}
	}

	// cache
	go cacheUser(db.cache, user)

	return &user, nil
}

func (db *UserDb) UpdateUser(user *User) error {
	const loc = "Models.User.UpdateUser"
	const qs = `UPDATE "users" SET "nickname" = $2, "summary" = $3, "avatar" = $4 WHERE "username" = $1;`

	r, err := sqlExec(db.client.Exec(qs, user.Username, user.Nickname, user.Summary, user.Avatar))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("update user %s", user.Username), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}

	// cache
	go func(u User) {
		db.cache.UpdateJson("user:"+user.Username.String(), &UserCache{
			Nickname: u.Nickname, Summary: u.Summary, Avatar: u.Avatar, Lock: u.Lock,
		})
	}(*user)
	return nil
}

func (db *UserDb) QueryInboxes(usernames []UD) (inboxes []string, err error) {
	const loc = "Models.User.QueryInboxes"
	const qs = `SELECT "inbox", "shared" FROM "foreign_inboxes" WHERE "username" = ANY($1);`
	r, err := db.client.Query(qs, pq.Array(usernames))
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query inboxes of %v", usernames), err)
		return nil, ErrDbInternal
	}
	defer r.Close()

	m := make(map[string]bool)
	for r.Next() {
		var inbox string
		var shared sql.NullString
		if err := r.Scan(&inbox, &shared); err != nil {
			continue
		}
		if shared.Valid {
			inbox = shared.String
		}
		if !m[inbox] {
			m[inbox] = true
			inboxes = append(inboxes, inbox)
		}
	}
	return inboxes, nil
}
