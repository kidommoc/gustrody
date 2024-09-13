package models

import (
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/logging"
)

type IUserAccount interface {
	SetLocalUser(user User, pswd string) error
	SetForeignUser(user User, inbox, sharedInbox string) error
	UpdateForeignInboxes(username, inbox, sharedInbox string) error
	QueryKeys(username string) (pub string, pri string, err error)

	QueryPreferences(username string) (pf *Preferences, err error)
	UpdatePreferences(username string, pf *Preferences) error
	QueryPassword(username string) (password string, err error)
	UpdatePassword(username string, password string) error
}

func (db *UserDb) SetLocalUser(user User, pswd string) error {
	const loc = "Models.User.SetLocal"
	const qu = `INSERT INTO "users"("username", "foreign", "id", "nickname", "pub_key", "pri_key")
				VALUES($1, FALSE, $2, $3, $4, $5);`
	const qp = `INSERT INTO "user_preferences"("username", "password") VALUES($1, $2);`

	for i := 0; i < txMaxRetries; i++ {
		if err := func() error {
			tx, err := db.client.BeginTx(defaultCtx, nil)
			if err != nil {
				return ErrDbInternal
			}
			defer tx.Rollback()

			// check duplicated
			ue, err := isUserExist(tx, user.Username.String())
			if err != nil {
				logging.Cannot(db.lg, loc, fmt.Sprintf("check existence of %s", user.Username), err)
			}
			if ue {
				return ErrDuplicated
			}

			// set user
			if _, err := sqlExec(tx.Exec(qu,
				user.Username, user.ID, user.Nickname, user.PubKey, user.PriKey,
			)); err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("set local user %s", user.Username), err)
				return ErrDbInternal
			}

			// set preferences
			if _, err := sqlExec(tx.Exec(qp, user.Username, pswd)); err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("set local user's preferences %s", user.Username), err)
				return ErrDbInternal
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("t")
			}
			return nil
		}(); err == nil || err.Error() != "t" {
			return err
		}
	}
	return ErrMaxRetries
}

func (db *UserDb) SetForeignUser(user User, inbox, sharedInbox string) error {
	const loc = "Models.User.SetForeign"
	const qu = `INSERT INTO "users"(
				  "username", "foreign", "id", "nickname",
				  "summary", "avatar", "lock", "pub_key"
				)
				VALUES($1, TRUE, $2, $3, $4, $5, $6, $7);`
	const qi = `INSERT INTO "foreign_inboxes"("username", "inbox", "shared") VALUES($1, $2, $3);`

	for i := 0; i < txMaxRetries; i++ {
		if err := func() error {
			tx, err := db.client.BeginTx(defaultCtx, nil)
			if err != nil {
				return ErrDbInternal
			}
			defer tx.Rollback()

			// check duplicated
			ue, err := isUserExist(tx, user.Username.String())
			if err != nil {
				logging.Cannot(db.lg, loc, fmt.Sprintf("check existence of %s", user.Username), err)
			}
			if ue {
				return ErrDuplicated
			}

			// set user
			if _, err := sqlExec(tx.Exec(qu,
				user.Username, user.ID, user.Nickname,
				user.Summary, user.Avatar, user.Lock, user.PubKey,
			)); err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("set foreign user %s", user.Username), err)
				return ErrDbInternal
			}

			// set inboxes
			var s *string = nil
			if sharedInbox != "" {
				s = &sharedInbox
			}
			if _, err := sqlExec(tx.Exec(qi, user.Username, inbox, s)); err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("set foreign user's inbox %s", user.Username), err)
				return ErrDbInternal
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("t")
			}
			return nil
		}(); err == nil || err.Error() != "t" {
			return err
		}
	}
	return ErrMaxRetries
}

func (db *UserDb) QueryKeys(username string) (pub string, pri string, err error) {
	const loc = "Models.User.QueryKeys"
	const qs = `SELECT "pub_key", "pri_key" FROM "users" WHERE "username" = $1;`

	r := db.client.QueryRow(qs, username)
	var pk sql.NullString
	if err := r.Scan(&pub, &pk); err != nil {
		switch err {
		case sql.ErrNoRows:
			return "", "", ErrNotFound
		default:
			logging.Cannot(db.lg, loc, fmt.Sprintf("scan result when query keys of %s", username), err)
			return "", "", ErrDbInternal
		}
	}
	if pk.Valid {
		pri = pk.String
	}
	return pub, pri, nil
}

func (db *UserDb) QueryPreferences(username string) (pf *Preferences, err error) {
	const loc = "Models.User.QueryPreferences"
	const qs = `SELECT "preferences" FROM "user_preferences" WHERE "username" = $1;`
	r := db.client.QueryRow(qs, username)
	if err := r.Scan(&pf); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			logging.Cannot(db.lg, loc, fmt.Sprintf("scan result when query preferences of %s", username), err)
			return nil, ErrDbInternal
		}
	}
	return pf, nil
}

func (db *UserDb) UpdatePreferences(username string, pf *Preferences) error {
	const loc = "Models.User.UpdatePreferences"
	const qu = `UPDATE "users" SET "lock" = $2 WHERE "username" = $1;`
	const qp = `UPDATE "user_preferences" SET "preferences" = $2 WHERE "username" = $1;`

	for i := 0; i < txMaxRetries; i++ {
		if err := func() error {
			tx, err := db.client.BeginTx(defaultCtx, nil)
			if err != nil {
				return ErrDbInternal
			}
			defer tx.Rollback()

			// set lock
			r, err := sqlExec(tx.Exec(qu, username, pf.Lock))
			if err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("update lock of %s", username), err)
				return ErrDbInternal
			}
			if r == 0 {
				return ErrNotFound
			}

			// set preferences
			r, err = sqlExec(tx.Exec(qp, username, *pf))
			if err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("update preferences of %s", username), err)
				return ErrDbInternal
			}
			if r == 0 {
				return ErrNotFound
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("t")
			}
			return nil
		}(); err == nil || err.Error() != "t" {
			return err
		}
	}
	return ErrMaxRetries
}

func (db *UserDb) UpdateForeignInboxes(username, inbox, sharedInbox string) error {
	const loc = "Models.User.UpdateForeignInboxes"
	const qs = `UPDATE "foreign_inboxes" SET "inbox" = $2, "shared" = $3 WHERE "username" = $1;`

	var s *string = nil
	if sharedInbox != "" {
		s = &sharedInbox
	}
	r, err := sqlExec(db.client.Exec(qs, username, inbox, s))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("update inboxes of foreign user %s", username), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}

func (db *UserDb) QueryPassword(username string) (password string, err error) {
	loc := "Model.User.QueryPassword"
	const qs = `SELECT "password" FROM "user_preferences" WHERE "username" = $1;`

	r := db.client.QueryRow(qs, username)
	if err := r.Scan(&password); err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query password of %s", username), err)
		return "", ErrDbInternal
	}
	return password, nil
}

func (db *UserDb) UpdatePassword(username string, password string) error {
	loc := "Model.User.UpdatePassword"
	const qs = `UPDATE "user_preferences" SET "password" = $2 WHERE "username" = $1;`

	r, err := sqlExec(db.client.Exec(qs, username, password))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("set password of %s", username), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}
