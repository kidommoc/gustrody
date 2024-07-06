package models

import "database/sql"

type IUserInfo interface {
	IsUserExist(username string) bool
	QueryUser(username string) (user User, err error)
	// uses: User.Username, User.Nickname, User.Summary, User.Avatar
	UpdateUser(user *User) error
}

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
			  "username", "nickname", "summary", "createdAt"
			FROM users
			WHERE "username" = $1;`
	r := conn.QueryOne(qs, username)
	var nkn sql.NullString
	var smy sql.NullString
	if e := r.Scan(
		&user.Username.Username, &nkn, &smy, &user.Date,
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
	r, err := conn.Exec(qs, user.Username.Username,
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
