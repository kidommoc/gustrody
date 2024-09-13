package models

import (
	"database/sql"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

const postJsonType = "::post"

type Post struct {
	ID        string    `json:"id"`
	FederalID string    `json:"fid"`
	User      UD        `json:"user"`
	Date      time.Time `json:"date"`
	Vsb       utils.Vsb `json:"vsb"`
	Content   string    `json:"content"`
	Media     []Img     `json:"media"`
	Replying  string    `json:"replying,omitempty"`  // post id, temporary field
	ReplyTo   UD        `json:"replyTo,omitempty"`   // user id, temporary field
	SharedBy  UD        `json:"sharedBy,omitempty"`  // user id, temporary field
	Replies   int64     `json:"replies"`             // count, temporary field
	Shares    int64     `json:"shares"`              // count, temporary field
	Likes     int64     `json:"likes"`               // count, temporary field
	Tombstone bool      `json:"tombstone,omitempty"` // deletion mark
	Type      string    `json:"__type,omitempty"`    // used for cache update
}

// used for update cache
type PostCache struct {
	atomic    bool
	Tombstone bool    `json:"tombstone,omitempty"` // only meaningful when true
	Content   *string `json:"content,omitempty"`   // content should not be empty string
	Media     *[]Img  `json:"media,omitempty"`
	Replies   int64   `json:"replies,omitempty"`
	Shares    int64   `json:"shares,omitempty"`
	Likes     int64   `json:"likes,omitempty"`
}

// implements cache.IJsonable
func (p *PostCache) Json() (map[string]interface{}, error) {
	m, err := structToMap(p)
	if err != nil {
		return nil, err
	}
	m[jsonTypeKey] = postJsonType
	return m, nil
}

// implements cache.IJsonable
func (p *PostCache) Atomic() bool {
	return p.atomic
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
