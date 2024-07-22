package models

import (
	"database/sql"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type Post struct {
	ID       string    `json:"id,omitempty"`
	Url      string    `json:"url,omitempty"`
	User     UD        `json:"user,omitempty"`
	Date     time.Time `json:"date,omitempty"`
	Vsb      utils.Vsb `json:"vsb,omitempty"`
	Content  string    `json:"content,omitempty"`
	Media    []Img     `json:"media"`
	Replying string    `json:"replying,omitempty"` // post id
	ReplyTo  string    `json:"replyTo,omitempty"`  // user id, temporary field
	SharedBy string    `json:"sharedBy,omitempty"` // user id, temporary field
	Likes    int64     `json:"likes"`              // count, temporary field
	Shares   int64     `json:"shares"`             // count, temporary field
	ActDate  string    `json:"actDate,omitempty"`  // temporary field, used in sort
	Level    int       `json:"level,omitempty"`    // temporary field, used in replying and replies
}

// db

type PostDb struct {
	lg     logging.Logger
	client *sql.DB
	cache  *CacheDb
}

var postIns *PostDb = nil

func PostInstance(lg logging.Logger, cl *sql.DB, cache *CacheDb) *PostDb {
	if postIns == nil {
		postIns = &PostDb{
			lg:     lg,
			client: cl,
			cache:  cache,
		}
	}
	return postIns
}
