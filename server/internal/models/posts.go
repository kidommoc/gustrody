package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"

	_db "github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type Img struct {
	Type string `json:"type"`
	Url  string `json:"url"`
	Alt  string `json:"alt,omitempty"`
}

// implement database/sql/driver.Valuer
func (img Img) Value() (driver.Value, error) {
	if img.Alt == "" {
		return fmt.Sprintf(`"(%s,%s,)"`, img.Type, img.Url), nil
	} else {
		return fmt.Sprintf(`"(%s,%s,%s)"`, img.Type, img.Url, img.Alt), nil
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
	ID       string           `json:"id"`
	Url      string           `json:"url"`
	User     UD               `json:"user"`
	Date     time.Time        `json:"date"`
	Vsb      utils.Vsb        `json:"vsb"`
	Content  string           `json:"content"`
	Media    Array[Img, *Img] `json:"media"`    // magic but sucks
	Replying string           `json:"replying"` // post id
	ReplyTo  string           `json:"replyTo"`  // user id, temporary field
	SharedBy string           `json:"sharedBy"` // user id, temporary field
	Likes    int64            `json:"likes"`    // count, temporary field
	Shares   int64            `json:"shares"`   // count, temporary field
	ActDate  string           `json:"actDate"`  // temporary field, used in sort
	Level    int              `json:"level"`    // temporary field, used in replying and replies
}

// db

type PostDb struct {
	lg   logging.Logger
	pool _db.ConnPool[_db.PqConn]
}

var postIns *PostDb = nil

func PostInstance(lg logging.Logger) *PostDb {
	if postIns == nil {
		postIns = &PostDb{
			lg:   lg,
			pool: _db.MainPool(nil),
		}
	}
	return postIns
}
