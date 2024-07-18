package models

import (
	"database/sql"
	"time"

	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/lib/pq"
)

type IPostShare interface {
	QueryShares(id string) (list []UD, err error)
	SetShare(user UD, id string, date time.Time, vsb utils.Vsb) error
	RemoveShare(user UD, id string) error
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryShares(id string) (list []UD, err error) {
	logger := db.lg
	qs := ` SELECT "shares"
			FROM posts
			WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)

	e := r.Scan(pq.Array(&list))
	if e != nil {
		switch e {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			logger.Error("[Model.Like] Cannot scan row", e)
			return nil, ErrDbInternal
		}
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
//   - Dunplicate "share"
func (db *PostDb) SetShare(user UD, id string, date time.Time, vsb utils.Vsb) error {
	logger := db.lg
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	tx, e := db.client.Begin()
	if e != nil {
		logger.Error("[Model.Share] Cannot start transaction", e)
		return ErrDbInternal
	}

	// update posts.shares
	qs := ` UPDATE posts
			SET "shares" = ARRAY_APPEND("shares", $1)
			WHERE "id" = $2 AND ARRAY_POSITION("shares", $1) IS NULL;`
	r, e := sqlExec(tx.Exec(qs, user, id))
	if e != nil {
		logger.Error("[Model.Share] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}

	// insert into shares
	qs = `  INSERT INTO shares("user", "id", "date", "vsb")
			VALUES ($1, $2, $3, $4);`
	r, e = sqlExec(tx.Exec(qs, user, id, date, vsb.String()))
	if e != nil {
		logger.Error("[Model.Share] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}

	if e := tx.Commit(); e != nil {
		logger.Error("[Model.Share] Cannot commit", e)
		return ErrDbInternal
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post", "share"
func (db *PostDb) RemoveShare(user UD, id string) error {
	logger := db.lg
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	tx, e := db.client.Begin()
	if e != nil {
		logger.Error("[Model.Share] Cannot start transaction", e)
		return ErrDbInternal
	}

	// use 2 annoymous func to ensure completely deletion

	err1 := func() error {
		qs := ` UPDATE posts
				SET "shares" = ARRAY_REMOVE("shares", $1)
				WHERE "id" = $2;`
		r, e := sqlExec(tx.Exec(qs, user, id))
		if e != nil {
			logger.Error("[Model.Share] Failed to exec", e)
			return ErrDbInternal
		}
		if r == 0 {
			return ErrNotFound
		}
		return nil
	}()

	err2 := func() error {
		qs := ` DELETE FROM shares
				WHERE "user" = $1 AND "id" = $2;`
		r, e := sqlExec(tx.Exec(qs, user, id))
		if e != nil {
			logger.Error("[Model.Share] Failed to exec", e)
			return ErrDbInternal
		}
		if r == 0 {
			return ErrNotFound
		}
		return nil
	}()

	if e := tx.Commit(); e != nil {
		logger.Error("[Model.Share] Cannot commit", e)
		return ErrDbInternal
	}

	if err1 != nil || err2 != nil {
		return ErrNotFound
	} else {
		return nil
	}
}
