package models

import (
	"database/sql"

	"github.com/lib/pq"
)

type IUserForeign interface {
	IsForeignExist(username UD) bool
	GetForeignUserByUD(username UD) (user ForeignUser, err error)
	GetForeignUserByID(id string) (user ForeignUser, err error)

	// insert or update.
	//
	// when update, uses ForeignUser.ID as index.
	SetForeignUser(user *ForeignUser) error

	GetInboxes(usernames []UD) (inboxes []string, err error)
}

type ForeignUser struct {
	Username UD     `json:"user"`
	ID       string `json:"id"`
	Avatar   string `json:"avatar"`    // local avatar url
	AvtUrl   string `json:"avatarUrl"` // remote avatar url
	Inbox    string `json:"inbox"`
	PubKey   string `json:"pub"`
}

func (db *UserDb) IsForeignExist(username UD) bool {
	logger := db.lg
	qs := `SELECT 1 FROM foreign_users WHERE "user" = $1;`
	r := db.client.QueryRow(qs, username)
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

func (db *UserDb) GetForeignUserByUD(username UD) (user ForeignUser, err error) {
	logger := db.lg
	qs := ` SELECT "user", "id", "avatar", "avatarUrl", "pub", "inbox"
			FROM foreign_users
			WHERE "user" = $1;`
	r := db.client.QueryRow(qs, username)
	if err := r.Scan(
		&user.Username, &user.ID, &user.Avatar, &user.AvtUrl,
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

func (db *UserDb) GetForeignUserByID(id string) (user ForeignUser, err error) {
	logger := db.lg
	qs := ` SELECT "user", "id", "avatar", "avatarUrl", "pub", "inbox"
			FROM foreign_users
			WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)
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
	var err error
	if db.IsForeignExist(user.Username) {
		qs := ` UPDATE foreign_users
				SET "avatar" = $2, "avatarUrl" = $3, "pub" = $4, "inbox" = $5
				WHERE "id" = $1;`
		_, err = db.client.Exec(qs, user.ID, user.Avatar, user.AvtUrl, user.PubKey, user.Inbox)
	} else {
		qs := ` INSERT INTO foreign_users("user", "id", "avatar", "avatarUrl", "pub", "inbox")
				VALUES ($1, $2, $3, $4, $5, $6);`
		_, err = db.client.Exec(qs, user.Username, user.ID, user.Avatar, user.AvtUrl, user.PubKey, user.Inbox)
	}
	if err != nil {
		logger.Error("[Model.UserForeign] Failed to execute set", err)
		return ErrDbInternal
	}
	return nil
}

func (db *UserDb) GetInboxes(usernames []UD) (inboxes []string, err error) {
	logger := db.lg
	qs := ` SELECT DISTINCT "inbox"
			FROM foreign_users
			WHERE "user" = ANY($1);`
	r, err := db.client.Query(qs, pq.Array(usernames))
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
