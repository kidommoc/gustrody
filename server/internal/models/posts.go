package models

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type Img struct {
	Type string `json:"type,omitempty"`
	Url  string `json:"url,omitempty"`
	Alt  string `json:"alt,omitempty"`
}

// implement database/sql/driver.Valuer
func (img Img) Value() (driver.Value, error) {
	if img.Alt == "" {
		return fmt.Sprintf(`(%s,%s,)`, img.Type, img.Url), nil
	} else {
		return fmt.Sprintf(`(%s,%s,%s)`, img.Type, img.Url, img.Alt), nil
	}
}

// implement database/sql.Scanner
func (img *Img) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Scan img: src cannot cast to []byte")
	}
	fields := strings.Split(strings.Trim(string(b), "()"), ",")
	if len(fields) < 2 {
		return fmt.Errorf("Scan img: wrong fields number")
	}
	if fields[0] == "" {
		return fmt.Errorf("Scan img: empty type")
	}
	if fields[1] == "" {
		return fmt.Errorf("Scan img: empty url")
	}
	img.Type = fields[0]
	img.Url = fields[1]
	if len(fields) > 2 {
		img.Alt = fields[2]
	}
	return nil
}

type Post struct {
	ID       string           `json:"id,omitempty"`
	Url      string           `json:"url,omitempty"`
	User     UD               `json:"user,omitempty"`
	Date     time.Time        `json:"date,omitempty"`
	Vsb      utils.Vsb        `json:"vsb,omitempty"`
	Content  string           `json:"content,omitempty"`
	Media    Array[Img, *Img] `json:"media,omitempty"`    // magic but sucks
	Replying string           `json:"replying,omitempty"` // post id
	ReplyTo  string           `json:"replyTo,omitempty"`  // user id, temporary field
	SharedBy string           `json:"sharedBy,omitempty"` // user id, temporary field
	Likes    int64            `json:"likes,omitempty"`    // count, temporary field
	Shares   int64            `json:"shares,omitempty"`   // count, temporary field
	ActDate  string           `json:"actDate,omitempty"`  // temporary field, used in sort
	Level    int              `json:"level,omitempty"`    // temporary field, used in replying and replies
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
