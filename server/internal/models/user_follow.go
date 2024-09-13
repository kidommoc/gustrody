package models

import (
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/logging"
)

type IUserFollow interface {
	// at least one should be local user
	IsFollowing(username, target UD) bool
	QueryFollowInfo(username string) (follows int64, followed int64, err error)
	QueryFollowings(username string) (list []UD, err error)
	QueryFollowers(username string) (list []UD, err error)
	// at least one should be local user
	SetFollow(from, to UD) error
	// at least one should be local user
	RemoveFollow(from, to UD) error
}

func isFollowing(client ISqlClient, username, target UD) (bool, error) {
	if client == nil {
		return false, ErrFormat
	}
	const qs = `SELECT 1 FROM "follow" WHERE "from" = $1 AND "to" = $2;`
	r := client.QueryRow(qs, username, target)
	var n int
	if err := r.Scan(&n); err != nil {
		switch err {
		case sql.ErrNoRows:
		default:
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (db *UserDb) IsFollowing(username, target UD) bool {
	const loc = "Models.User.IsFollowing"
	r, err := isFollowing(db.client, username, target)
	if err != nil {
		logging.Cannot(db.lg, loc, fmt.Sprintf("query %s", username), err)
	}
	return r
}

func (db *UserDb) QueryFollowInfo(username string) (follows int64, followed int64, err error) {
	const loc = "Models.User.QueryFollowInfo"
	const qs = `SELECT "followings", "followers" FROM "follow_data" WHERE "user" = $1;`
	r := db.client.QueryRow(qs, username)
	if e := r.Scan(&follows, &followed); e != nil {
		switch e {
		case sql.ErrNoRows:
			return 0, 0, ErrNotFound
		default:
			logging.Cannot(db.lg, loc, fmt.Sprintf("scan result of follow info of %s", username), err)
			return 0, 0, ErrDbInternal
		}
	}
	return follows, followed, nil
}

func (db *UserDb) QueryFollowings(username string) (list []UD, err error) {
	const loc = "Models.User.QueryFollowings"
	if !db.IsUserExist(username) {
		return list, ErrNotFound
	}

	const qs = `SELECT "to" FROM "follow" WHERE "from" = $1;`
	r, e := db.client.Query(qs, username)
	if e != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("query followings of %s", username), err)
		return nil, ErrDbInternal
	}
	defer r.Close()

	list = make([]UD, 0)
	for r.Next() {
		var f UD
		if e := r.Scan(&f); e != nil {
			continue
		}
		list = append(list, f)
	}
	return list, nil
}

func (db *UserDb) QueryFollowers(username string) (list []UD, err error) {
	const loc = "Models.User.QueryFollowers"
	if !db.IsUserExist(username) {
		return nil, ErrNotFound
	}

	if !db.IsUserExist(username) {
		return list, ErrNotFound
	}

	const qs = `SELECT "from" FROM "follow" WHERE "to" = $1;`
	r, e := db.client.Query(qs, username)
	if e != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("query followers of %s", username), err)
		return nil, ErrDbInternal
	}
	defer r.Close()

	list = make([]UD, 0)
	for r.Next() {
		var f UD
		if e := r.Scan(&f); e != nil {
			continue
		}
		list = append(list, f)
	}
	return list, nil
}

func (db *UserDb) SetFollow(from, to UD) error {
	const loc = "Models.User.SetFollow"
	const qs = `INSERT INTO "follow" VALUES ($1, $2);`

	for i := 0; i < txMaxRetries; i++ {
		if err := func() error {
			tx, err := db.client.BeginTx(defaultCtx, nil)
			if err != nil {
				return ErrDbInternal
			}
			defer tx.Rollback()

			// chech duplicated
			fe, err := isFollowing(tx, from, to)
			if err != nil {
				logging.Cannot(db.lg, loc, fmt.Sprintf("query following from %s to %s", from, to), err)
			}
			if fe {
				return ErrDuplicated
			}

			// set follow
			r, err := sqlExec(tx.Exec(qs, from, to))
			if err != nil {
				logging.FailedTo(db.lg, loc, fmt.Sprintf("set following from %s to %s", from, to), err)
				return ErrDbInternal
			}
			if r == 0 {
				return ErrDuplicated
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

func (db *UserDb) RemoveFollow(from, to UD) error {
	const loc = "Models.User.RemoveFollow"
	const qs = `DELETE FROM "follow" WHERE "from" = $1 AND "to" = $2;`
	r, err := sqlExec(db.client.Exec(qs, from, to))
	if err != nil {
		logging.FailedTo(db.lg, loc, fmt.Sprintf("remove following from %s to %s", from, to), err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDuplicated
	}
	return nil
}
