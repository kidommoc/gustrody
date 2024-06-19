package models

import (
	"database/sql"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type User struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Summary  string `json:"summary"`
}

type follow struct {
	From string
	To   string
}

// db

type IUserDb interface {
	IsUserExist(username string) bool
	QueryUser(username string) (user User, err utils.Error)
	QueryUserFollowInfo(username string) (follows int64, followed int64, err utils.Error)
	QueryUserFollowings(username string) (list []*User, err utils.Error)
	QueryUserFollowers(username string) (list []*User, err utils.Error)
	SetFollow(from string, to string) utils.Error
	RemoveFollow(from string, to string) utils.Error
}

// should implemented with Postgre
type UsersDb struct {
	pool *_db.ConnPool[*_db.PqConn]
}

var userIns *UsersDb = nil

func UserInstance() *UsersDb {
	if userIns == nil {
		userIns = &UsersDb{
			pool: _db.MainPool(),
		}
	}
	return userIns
}

// functions

func (db *UsersDb) IsUserExist(username string) bool {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return false
	}
	defer conn.Close()

	qs := ` SELECT 1
			FROM users
			WHERE "username" = $1;`
	r := conn.QueryOne(qs, username)
	var n int
	if e := r.Scan(&n); e != nil { // can't understand why i MUST scan to check whether result is empty. silly design
		switch e {
		case sql.ErrNoRows:
			return false
		default:
			logger.Error("[Model.Users] Cannot query", newErr(ErrDbInternal, e.Error()))
			return false
		}
	}
	return true
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UsersDb) QueryUser(username string) (user User, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return user, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()

	qs := ` SELECT
			  "username", "nickname", "summary"
			FROM users
			WHERE "username" = $1;`
	r := conn.QueryOne(qs, username)
	user = User{}
	var nkn sql.NullString
	var smy sql.NullString
	if e := r.Scan(
		&user.Username, &nkn, &smy,
	); e != nil {
		switch e {
		case sql.ErrNoRows:
			return user, newErr(ErrNotFound, "user")
		default:
			err = newErr(ErrDbInternal, e.Error())
			logger.Error("[Model.Users] Cannot scan row", err)
			return user, err
		}
	}
	if nkn.Valid {
		user.Nickname = nkn.String
	}
	if smy.Valid {
		user.Summary = smy.String
	}
	return user, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UsersDb) QueryUserFollowInfo(username string) (follows int64, followed int64, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return -1, -1, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsUserExist(username) {
		return -1, -1, newErr(ErrNotFound, "user")
	}

	qs := ` SELECT "followings", "followers"
			FROM follow_info
			WHERE "user" = $1; `
	r := conn.QueryOne(qs, username)
	if e := r.Scan(&follows, &followed); e != nil {
		switch e {
		case sql.ErrNoRows:
			return 0, 0, nil
		default:
			err = newErr(ErrDbInternal, e.Error())
			logger.Error("[Model.Users] Cannot scan row", err)
			return -1, -1, err
		}
	}
	return follows, followed, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UsersDb) QueryUserFollowings(username string) (list []*User, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsUserExist(username) {
		return list, newErr(ErrNotFound, "user")
	}

	qs := ` SELECT "to" AS "following"
			FROM follow
			WHERE "from" = $1;
	`
	r, e := conn.Query(qs, username)
	if e != nil {
		return nil, newErr(ErrDbInternal, e.Error())
	}

	list = make([]*User, 0)
	for r.Next() {
		var f string
		if e := r.Scan(&f); e != nil {
			logger.Error("[Model] Cannot scan row", newErr(ErrDbInternal, e.Error()))
			continue
		}
		u, e := db.QueryUser(f)
		if e != nil {
			logger.Error("[Model] Cannot find user", newErr(ErrNotFound, e.Error()))
			continue
		}
		list = append(list, &u)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UsersDb) QueryUserFollowers(username string) (list []*User, err utils.Error) {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsUserExist(username) {
		return nil, newErr(ErrNotFound, "user")
	}

	qs := ` SELECT "from" AS "follower"
			FROM follow
			WHERE "to" = $1;`
	r, e := conn.Query(qs, username)
	if e != nil {
		return nil, newErr(ErrDbInternal, e.Error())
	}

	list = make([]*User, 0)
	for r.Next() {
		var f string
		if e := r.Scan(&f); e != nil {
			logger.Error("[Model] Cannot scan row", newErr(ErrDbInternal, e.Error()))
			continue
		}
		u, e := db.QueryUser(f)
		if e != nil {
			logger.Error("[Model] Cannot find user", newErr(ErrNotFound, e.Error()))
			continue
		}
		list = append(list, &u)
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "from", "to"
//   - Dunplicate "follow"
func (db *UsersDb) SetFollow(from string, to string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsUserExist(from) {
		return newErr(ErrNotFound, "from")
	}
	if !db.IsUserExist(to) {
		return newErr(ErrNotFound, "to")
	}

	qs := ` INSERT INTO follow
			VALUES ($1, $2);`
	r, e := conn.Exec(qs, from, to)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrDunplicate, "follow")
	}
	return nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "from", "to", "follow"
func (db *UsersDb) RemoveFollow(from string, to string) utils.Error {
	logger := logging.Get()
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return newErr(ErrDbInternal, err.Error())
	}
	defer conn.Close()
	if !db.IsUserExist(from) {
		return newErr(ErrNotFound, "from")
	}
	if !db.IsUserExist(to) {
		return newErr(ErrNotFound, "to")
	}

	qs := ` DELETE FROM follow
			WHERE "from" = $1 AND "to" = $2;`
	r, e := conn.Exec(qs, from, to)
	if e != nil {
		return newErr(ErrDbInternal, e.Error())
	}
	if r == 0 {
		return newErr(ErrNotFound, "follow")
	}
	return nil
}
