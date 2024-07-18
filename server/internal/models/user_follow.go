package models

import "database/sql"

type IUserFollow interface {
	// at least one should be local user
	IsFollowing(username, target UD) bool
	QueryUserFollowInfo(username string) (follows int64, followed int64, err error)
	// uses: User.Username, User.Nickname, User.Avatar
	QueryUserFollowings(username string) (list []UD, err error)
	// uses: User.Username, User.Nickname, User.Avatar
	QueryUserFollowers(username string) (list []UD, err error)
	// at least one should be local user
	SetFollow(from, to UD) error
	// at least one should be local user
	RemoveFollow(from, to UD) error
}

func (db *UserDb) IsFollowing(username, target UD) bool {
	logger := db.lg
	qs := ` SELECT 1
			FROM follow
			WHERE "from" = $1 AND "to" = $2;`
	r := db.client.QueryRow(qs, username, target)
	var n int
	if e := r.Scan(&n); e != nil {
		switch e {
		case sql.ErrNoRows:
			return false
		default:
			logger.Error("[Model.UserFollow] Cannot query", e)
			return false
		}
	}
	return true
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UserDb) QueryUserFollowInfo(username string) (follows int64, followed int64, err error) {
	logger := db.lg
	if !db.IsUserExist(username) {
		return -1, -1, ErrNotFound
	}

	qs := ` SELECT "followings", "followers"
			FROM follow_info
			WHERE "user" = $1; `
	r := db.client.QueryRow(qs, username)
	if e := r.Scan(&follows, &followed); e != nil {
		switch e {
		case sql.ErrNoRows:
			return 0, 0, nil
		default:
			logger.Error("[Model.UserFollow] Cannot scan row", e)
			return -1, -1, ErrDbInternal
		}
	}
	return follows, followed, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UserDb) QueryUserFollowings(username string) (list []UD, err error) {
	logger := db.lg
	if !db.IsUserExist(username) {
		return list, ErrNotFound
	}

	qs := ` SELECT "to" AS "following"
			FROM follow
			WHERE "from" = $1;`
	r, e := db.client.Query(qs, username)
	if e != nil {
		logger.Error("[Model.UserFollow] Failed to query", e)
		return nil, ErrDbInternal
	}

	list = make([]UD, 0)
	for r.Next() {
		var f UD
		if e := r.Scan(&f); e != nil {
			logger.Error("[Model.UserFollow] Cannot scan row", e)
			continue
		}
		list = append(list, f)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UserDb) QueryUserFollowers(username string) (list []UD, err error) {
	logger := db.lg
	if !db.IsUserExist(username) {
		return nil, ErrNotFound
	}

	qs := ` SELECT "from" AS "follower"
			FROM follow
			WHERE "to" = $1;`
	r, e := db.client.Query(qs, username)
	if e != nil {
		logger.Error("[Model.UserFollow] Failed to query", e)
		return nil, ErrDbInternal
	}

	list = make([]UD, 0)
	for r.Next() {
		var f UD
		if e := r.Scan(&f); e != nil {
			logger.Error("[Model.UserFollow] Cannot scan row", e)
			continue
		}
		list = append(list, f)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "from", "to"
//   - Dunplicate "follow"
func (db *UserDb) SetFollow(from, to UD) error {
	logger := db.lg
	qs := ` INSERT INTO follow
			VALUES ($1, $2);`
	r, err := sqlExec(db.client.Exec(qs, from, to))
	if err != nil {
		logger.Error("[Model.UserFollow] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "from", "to", "follow"
func (db *UserDb) RemoveFollow(from, to UD) error {
	logger := db.lg
	qs := ` DELETE FROM follow
			WHERE "from" = $1 AND "to" = $2;`
	r, err := sqlExec(db.client.Exec(qs, from, to))
	if err != nil {
		logger.Error("[Model.UserFollow] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}
	return nil
}
