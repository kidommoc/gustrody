package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

type IPostShare interface {
	QueryShares(id string) (list []UD, err error)
	SetShare(user UD, id string, date time.Time, vsb utils.Vsb) error
	RemoveShare(user UD, id string) error
}

func (db *PostDb) QueryShares(id string) (list []UD, err error) {
	const loc = "Models.SharePost.Query"
	const qs = `SELECT "user" FROM "share" WHERE "tgt" = $1;`
	r, err := db.client.Query(qs, id)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query shares of %s", id), err)
		return nil, ErrDbInternal
	}
	defer r.Close()

	for r.Next() {
		var u UD
		if err := r.Scan(&u); err != nil {
			continue
		}
		list = append(list, u)
	}
	return list, nil
}

func (db *PostDb) SetShare(user UD, id string, date time.Time, vsb utils.Vsb) error {
	const loc = "Models.SharePost.Set"

	const qd = `SELECT 1 FROM "share" WHERE "user" = $1 AND "tgt" = $2;`
	const qs = `INSERT INTO "share"("user", "tgt", "date", "vsb") VALUES($1, $2, $3, $4);`

	err := func() error {
		for i := 0; i < txMaxRetries; i++ {
			if err := func() error {
				tx, err := db.client.BeginTx(defaultCtx, nil)
				if err != nil {
					return ErrDbInternal
				}
				defer tx.Rollback()

				// check post
				pe, err := isPostExist(tx, id)
				if err != nil {
					logging.Cannot(db.lg, loc, fmt.Sprintf("check existence of post %s", id), err)
					return ErrDbInternal
				}
				if !pe {
					return ErrNotFound
				}

				// check user
				ue, err := isUserExist(tx, user.String())
				if err != nil {
					logging.Cannot(db.lg, loc, fmt.Sprintf("check existence of user %s", user.String()), err)
					return ErrDbInternal
				}
				if !ue {
					return ErrNotFound
				}

				// check duplicated
				r := tx.QueryRow(qd, user, id)
				var n int
				if err := r.Scan(&n); err != nil {
					switch err {
					case sql.ErrNoRows:
					default:
						logging.Cannot(db.lg, loc, fmt.Sprintf("check duplicated of %s sharing %s", user.String(), id), err)
						return ErrDbInternal
					}
				} else {
					return ErrDuplicated
				}

				// set share
				if _, err := sqlExec(tx.Exec(qs, user, id, date, vsb)); err != nil {
					logging.FailedTo(db.lg, loc, fmt.Sprintf("set %s sharing %s", user.String(), id), err)
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
	}()

	// cache
	if err == nil {
		go db.cache.UpdateJson("post:"+id, &PostCache{atomic: true, Shares: 1})
	}
	return err
}

func (db *PostDb) RemoveShare(user UD, id string) error {
	const loc = "Models.SharePost.Remove"
	const qs = `DELETE FROM "share" WHERE "user" = $1 AND "tgt" = $2;`
	r, err := sqlExec(db.client.Exec(qs, user, id))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("remove %s sharing %s", user.String(), id), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	go db.cache.UpdateJson("post:"+id, &PostCache{atomic: true, Shares: -1})
	return nil
}
