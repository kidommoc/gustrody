package models

import (
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/logging"
)

type IPostLike interface {
	QueryLikes(id string) (list []UD, err error)
	SetLike(user UD, id string) error
	RemoveLike(user UD, id string) error
}

func (db *PostDb) QueryLikes(id string) (list []UD, err error) {
	const loc = "Models.LikePost.Query"
	const qs = `SELECT "user" FROM "like" WHERE "tgt" = $1;`
	r, err := db.client.Query(qs, id)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query likes of %s", id), err)
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

func (db *PostDb) SetLike(user UD, id string) error {
	const loc = "Models.LikePost.Set"
	const qd = `SELECT 1 FROM "like" WHERE "user" = $1 AND "tgt" = $2;`
	const qs = `INSERT INTO "like"("user", "tgt") VALUES($1, $2);`

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
					logging.Cannot(db.lg, loc, fmt.Sprintf("check existence of %s", id), err)
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
						logging.Cannot(db.lg, loc, fmt.Sprintf("check duplicated of %s liking %s", user.String(), id), err)
						return ErrDbInternal
					}
				} else {
					return ErrDuplicated
				}

				// set like
				if _, err := sqlExec(tx.Exec(qs, user, id)); err != nil {
					logging.FailedTo(db.lg, loc, fmt.Sprintf("set %s liking %s", user.String(), id), err)
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
		go db.cache.UpdateJson("post:"+id, &PostCache{atomic: true, Likes: 1})
	}
	return err
}

func (db *PostDb) RemoveLike(user UD, id string) error {
	const loc = "Models.LikePost.Remove"
	const qs = `DELETE FROM "like" WHERE "user" = $1 AND "tgt" = $2;`
	r, err := sqlExec(db.client.Exec(qs, user, id))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("remove %s liking %s", user.String(), id), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	go db.cache.UpdateJson("post:"+id, &PostCache{atomic: true, Likes: -1})
	return nil
}
