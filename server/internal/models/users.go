package models

import (
	"crypto/rsa"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type KeyPair struct {
	Pub string `json:"pub"`
	Pri string `json:"pri"`
}

func (kp KeyPair) Value() (driver.Value, error) {
	pub := strings.ReplaceAll(kp.Pub, "\n", "#n")
	pri := strings.ReplaceAll(kp.Pri, "\n", "#n")
	return fmt.Sprintf("(\"%s\",\"%s\")", pub, pri), nil
}

func (kp *KeyPair) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Scan key pair: src cannot cast to []byte")
	}
	fields := strings.Split(strings.Trim(string(b), "()"), ",")
	if len(fields) != 2 {
		return fmt.Errorf("Scan key pair: wrong syntax")
	}
	kp.Pub = strings.ReplaceAll(strings.Trim(fields[0], "\""), "#n", "\n")
	kp.Pri = strings.ReplaceAll(strings.Trim(fields[1], "\""), "#n", "\n")
	return nil
}

type Preferences struct {
	PostVsb  string `json:"postVsb"`
	ShareVsb string `json:"shareVsb"`
}

func (p Preferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Preferences) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed.")
	}
	return json.Unmarshal(b, p)
}

type User struct {
	Username    string      `json:"username"`
	Nickname    string      `json:"nickname"`
	Summary     string      `json:"summary"`
	Avatar      string      `json:"avatar"`
	Keys        KeyPair     `json:"keys"`
	Preferences Preferences `json:"preferences"`
}

type follow struct {
	From string
	To   string
}

// db

type IUserAccount interface {
	// uses: User.Username, User.Nickname, User.Keys
	SetUser(user *User) error
	QueryUserKeys(username string) (pub *rsa.PublicKey, pri *rsa.PrivateKey, err error)
	QueryUserPreferences(username string) (pf *Preferences, err error)
	UpdateUserPreferences(username string, pf *Preferences) error
}

type IUserInfo interface {
	IsUserExist(username string) bool
	QueryUser(username string) (user User, err error)
	// uses: User.Username, User.Nickname, User.Summary, User.Avatar
	UpdateUser(user *User) error
}

type IUserFollow interface {
	IsFollowing(username, target string) bool
	QueryUserFollowInfo(username string) (follows int64, followed int64, err error)
	QueryUserFollowings(username string) (list []*User, err error)
	QueryUserFollowers(username string) (list []*User, err error)
	SetFollow(from, to string) error
	RemoveFollow(from, to string) error
}

// should implemented with Postgre
type UserDb struct {
	lg   logging.Logger
	pool *_db.ConnPool[*_db.PqConn]
}

var userIns *UserDb = nil

func UserInstance(lg logging.Logger) *UserDb {
	if userIns == nil {
		userIns = &UserDb{
			lg:   lg,
			pool: _db.MainPool(nil, nil),
		}
	}
	return userIns
}

// account

func (db *UserDb) SetUser(user *User) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` INSERT INTO users(
				"username", "nickname", "summary",
				"createdAt", "avatar", "keys"
			)
			VALUES (
				$1, $2, '',
				$3, '', $4
			);`
	r, err := conn.Exec(qs,
		user.Username, user.Nickname,
		time.Now().UTC(), user.Keys,
	)
	if err != nil {
		logger.Error("[Model.UserAccount] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrDunplicate
	}
	return nil
}

func (db *UserDb) QueryUserKeys(username string) (pub *rsa.PublicKey, pri *rsa.PrivateKey, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, nil, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT "keys"
			FROM users
			WHERE "username" = $1;`
	r := conn.QueryOne(qs, username)
	var kp KeyPair
	if e := r.Scan(&kp); e != nil {
		logger.Error("[Model.UserAccount] Cannot scan row", e)
		return nil, nil, ErrDbInternal
	}
	pub = utils.GetPublicKey(kp.Pub)
	pri = utils.GetPrivateKey(kp.Pri)
	return pub, pri, nil
}

func (db *UserDb) QueryUserPreferences(username string) (pf *Preferences, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, ErrDbInternal
	}
	defer conn.Close()

	qs := ` SELECT "preferences"
			FROM users
			WHERE "username" = $1;`
	r := conn.QueryOne(qs, username)
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
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` UPDATE users
			SET "preferences" = $2
			WHERE "username" = $1;`
	r, err := conn.Exec(qs, username, *pf)
	if err != nil {
		logger.Error("[Model.UserAccount] Failed to execute", err)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}

// info

func (db *UserDb) IsUserExist(username string) bool {
	logger := db.lg
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
			logger.Error("[Model.UserInfo] Cannot query", e)
			return false
		}
	}
	return true
}

// ERRORS
//
//   - DbInternal
//   - NotFound "user"
func (db *UserDb) QueryUser(username string) (user User, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return user, ErrDbInternal
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
			return user, ErrNotFound
		default:
			logger.Error("[Model.UserInfo] Cannot scan row", e)
			return user, ErrDbInternal
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

func (db *UserDb) UpdateUser(user *User) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` UPDATE users
			SET
			  "nickname" = $2, "summary" = $3, "avatar" = $4
			WHERE "username" = $1;`
	r, err := conn.Exec(qs, user.Username,
		user.Nickname, user.Summary, user.Avatar,
	)
	if err != nil {
		logger.Error("[Model.UserInfo] Failed to execute", err)
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}

// follow

func (db *UserDb) IsFollowing(username string, target string) bool {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return false
	}
	defer conn.Close()

	qs := ` SELECT 1
			FROM follow
			WHERE "from" = $1 AND "to" = $2;`
	r := conn.QueryOne(qs, username, target)
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
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return -1, -1, ErrDbInternal
	}
	defer conn.Close()
	if !db.IsUserExist(username) {
		return -1, -1, ErrNotFound
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
func (db *UserDb) QueryUserFollowings(username string) (list []*User, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, ErrDbInternal
	}
	defer conn.Close()
	if !db.IsUserExist(username) {
		return list, ErrNotFound
	}

	qs := ` SELECT "to" AS "following"
			FROM follow
			WHERE "from" = $1;
	`
	r, e := conn.Query(qs, username)
	if e != nil {
		logger.Error("[Model.UserFollow] Failed to query", e)
		return nil, ErrDbInternal
	}

	list = make([]*User, 0)
	for r.Next() {
		var f string
		if e := r.Scan(&f); e != nil {
			logger.Error("[Model.UserFollow] Cannot scan row", e)
			continue
		}
		u, e := db.QueryUser(f)
		if e != nil {
			logger.Error("[Model.UserFollow] Cannot find user", e)
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
func (db *UserDb) QueryUserFollowers(username string) (list []*User, err error) {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return nil, ErrDbInternal
	}
	defer conn.Close()
	if !db.IsUserExist(username) {
		return nil, ErrNotFound
	}

	qs := ` SELECT "from" AS "follower"
			FROM follow
			WHERE "to" = $1;`
	r, e := conn.Query(qs, username)
	if e != nil {
		logger.Error("[Model.UserFollow] Failed to query", e)
		return nil, ErrDbInternal
	}

	list = make([]*User, 0)
	for r.Next() {
		var f string
		if e := r.Scan(&f); e != nil {
			logger.Error("[Model.UserFollow] Cannot scan row", e)
			continue
		}
		u, e := db.QueryUser(f)
		if e != nil {
			logger.Error("[Model.UserFollow] Cannot find user", e)
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
func (db *UserDb) SetFollow(from string, to string) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` INSERT INTO follow
			VALUES ($1, $2);`
	r, e := conn.Exec(qs, from, to)
	if e != nil {
		logger.Error("[Model.UserFollow] Failed to execute", e)
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
func (db *UserDb) RemoveFollow(from string, to string) error {
	logger := db.lg
	conn, err := db.pool.Open()
	if err != nil {
		logger.Error("[Model] Failed to open a connection", err)
		return ErrDbInternal
	}
	defer conn.Close()

	qs := ` DELETE FROM follow
			WHERE "from" = $1 AND "to" = $2;`
	r, e := conn.Exec(qs, from, to)
	if e != nil {
		logger.Error("[Model.UserFollow] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}
