package models

import (
	"database/sql"

	"github.com/lib/pq"
)

type IPostLike interface {
	QueryLikes(id string) (list []UD, err error)
	SetLike(user UD, id string) error
	RemoveLike(user UD, id string) error
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
func (db *PostDb) QueryLikes(id string) (list []UD, err error) {
	logger := db.lg
	qs := ` SELECT "likes"
			FROM posts
			WHERE "id" = $1;`
	r := db.client.QueryRow(qs, id)

	e := r.Scan(pq.Array(&list))
	if e != nil {
		switch e {
		case sql.ErrNoRows:
			return nil, ErrNotFound
		default:
			logger.Error("[Model.PostLike] Cannot scan row", e)
			return nil, ErrDbInternal
		}
	}
	return list, nil
}

// ERRORS
//
//   - DbInternal
//   - NotFound "post"
//   - Dunplicate "like"
func (db *PostDb) SetLike(user UD, id string) error {
	logger := db.lg
	if !db.IsPostExist(id) {
		return ErrNotFound
	}

	qs := ` UPDATE posts
			SET "likes" = ARRAY_APPEND("likes", $1)
			WHERE
			  "id" = $2
  			  AND ARRAY_POSITION("likes", $1) IS NULL;`
	r, e := sqlExec(db.client.Exec(qs, user, id))
	if e != nil {
		logger.Error("[Model.PostLike] Failed to execute", e)
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
//   - NotFound "post"
func (db *PostDb) RemoveLike(user UD, id string) error {
	logger := db.lg
	qs := ` UPDATE posts
			SET "likes" = ARRAY_REMOVE("likes", $1)
			WHERE "id" = $2;`
	r, e := sqlExec(db.client.Exec(qs, user, id))
	if e != nil {
		logger.Error("[Model.PostLike] Failed to execute", e)
		return ErrDbInternal
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}
