package models

import (
	"database/sql"
)

type IUserAccount interface {
	// uses: User.Username, User.Nickname, User.Keys
	SetUser(user *User) error
	QueryUserKeys(username string) (pub string, pri string, err error)
	QueryUserPreferences(username string) (pf *Preferences, err error)
	UpdateUserPreferences(username string, pf *Preferences) error
}

func (db *UserDb) SetUser(user *User) error {
	logger := db.lg
	qs := ` INSERT INTO users("username", "nickname", "summary", "avatar", "keys")
			VALUES ($1, $2, '', '', $3);`
	r, err := sqlExec(db.client.Exec(qs,
		user.Username, user.Nickname, user.Keys,
	))
	if err != nil {
		logger.Error("[Model.UserAccount] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}
	return nil
}

func (db *UserDb) QueryUserKeys(username string) (pub string, pri string, err error) {
	logger := db.lg
	qs := ` SELECT "keys"
			FROM users
			WHERE "username" = $1;`
	r := db.client.QueryRow(qs, username)
	var kp KeyPair
	if e := r.Scan(&kp); e != nil {
		logger.Error("[Model.UserAccount] Cannot scan row", e)
		return "", "", ErrDbInternal
	}
	return kp.Pub, kp.Pri, nil
}

func (db *UserDb) QueryUserPreferences(username string) (pf *Preferences, err error) {
	logger := db.lg
	qs := ` SELECT "preferences"
			FROM users
			WHERE "username" = $1;`
	r := db.client.QueryRow(qs, username)
	if err := r.Scan(&pf); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			logger.Error("[Model.UserAccount] Cannot scan row", err)
			return nil, ErrDbInternal
		}
	}
	return pf, nil
}

func (db *UserDb) UpdateUserPreferences(username string, pf *Preferences) error {
	logger := db.lg
	qs := ` UPDATE users
			SET "preferences" = $2
			WHERE "username" = $1;`
	r, err := sqlExec(db.client.Exec(qs, username, *pf))
	if err != nil {
		logger.Error("[Model.UserAccount] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}
	return nil
}
