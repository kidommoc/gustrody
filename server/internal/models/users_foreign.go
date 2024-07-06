package models

import (
	"database/sql"

	"github.com/lib/pq"
)

type IUserForeign interface {
	IsForeignExist(username UD) bool
	GetForeignUser(username UD) (user ForeignUser, err error)
	SetForeignUser(user *ForeignUser) error
	GetInboxes(usernames []UD) (inboxes []string, err error)
}

type ForeignUser struct {
	Username UD     `json:"user"`
	ID       string `json:"id"`
	Avatar   string `json:"avatar"`
	AvtUrl   string `json:"avatarUrl"`
	Inbox    string `json:"inbox"`
	PubKey   string `json:"pub"`
}

func (db *UserDb) IsForeignExist(username UD) bool {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.UserForeign] Failed to open a connection", err)
		return false
	}
	defer conn.Close()

	qs := `SELECT 1 FROM foreign_users WHERE "user" = $1;`
	r := conn.QueryOne(qs, username)
	var n int
	if err := r.Scan(&n); err != nil {
		switch err {
		case sql.ErrNoRows:
		default:
			logger.Error("[Model.UserForeign] Cannot query", err)
		}
		return false
	}
	return true
}

func (db *UserDb) GetForeignUser(username UD) (user ForeignUser, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.UserForeign] Failed to open a connection", err)
		return user, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT "user", "id", "avatar", "avatarUrl", "pub", "inbox"
			FROM foreign_users
			WHERE "user" = $1;`
	r := conn.QueryOne(qs, username)
	if err := r.Scan(
		&user.Username, &user.ID,
		&user.Avatar, &user.AvtUrl,
		&user.PubKey, &user.Inbox,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return user, ErrNotFound
		default:
			logger.Error("[Model.UserForeign] Cannot scan row of foreign user.", err)
		}
	}
	return user, nil
}

func (db *UserDb) SetForeignUser(user *ForeignUser) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.UserForeign] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	if db.IsForeignExist(user.Username) {
		qs := ` UPDATE foreign_users
				SET "avatar" = $2, "avatarUrl" = $3
				WHERE "user" = $1; `
		_, err = conn.Exec(qs, user.Username, user.Avatar, user.AvtUrl)
	} else {
		qs := ` INSERT INTO foreign_users("user", "id", "avatar", "avatarUrl", "pub", "inbox")
				VALUES ($1, $2, $3, $4, $5, $6);`
		_, err = conn.Exec(qs, user.Username, user.ID, user.Avatar, user.AvtUrl, user.PubKey, user.Inbox)
	}
	if err != nil {
		logger.Error("[Model.UserForeign] Failed to execute set", err)
		return ErrDbInternal
	}
	return nil
}

func (db *UserDb) GetInboxes(usernames []UD) (inboxes []string, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model.UserForeign] Failed to open a connection", err)
		return nil, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT DISTINCT "inbox"
			FROM foreign_users
			WHERE "user" = ANY($1);`
	r, err := conn.Query(qs, pq.Array(usernames))
	if err != nil {
		logger.Error("[Model.UserForeign] Cannot query inboxes.", err)
	}
	for r.Next() {
		var inbox string
		if err := r.Scan(&inbox); err != nil {
			logger.Error("[Model.UserForeign] Cannot scan row of inbox.", err)
			continue
		}
		inboxes = append(inboxes, inbox)
	}
	return
}
