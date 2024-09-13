package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

type UD struct {
	Username string `json:"username,omitempty"`
	Domain   string `json:"domain,omitempty"`
}

func NewUD(user string) UD {
	groups := utils.UdReg.FindStringSubmatch(user)
	if len(groups) >= 3 {
		return UD{groups[1], groups[2]}
	} else {
		if utils.UsernameReg.MatchString(user) {
			return UD{Username: user}
		} else {
			return UD{}
		}
	}
}

func (ud UD) String() string {
	if ud.Domain == "" {
		return ud.Username
	} else {
		return fmt.Sprintf("%s@%s", ud.Username, ud.Domain)
	}
}

// implement database/sql/driver.Valuer
func (ud UD) Value() (driver.Value, error) {
	return ud.String(), nil
}

// implement database/sql.Scanner
func (ud *UD) Scan(src interface{}) error {
	if src == nil {
		*ud = NewUD("")
		return nil
	}
	var s string
	b, ok := src.([]byte)
	if !ok {
		s, ok = src.(string)
		if !ok {
			return fmt.Errorf("Scan username-domain failed: src cannot cast to string or []byte")
		}
	} else {
		s = string(b)
	}
	*ud = NewUD(s)
	if ud.Username == "" {
		return fmt.Errorf(`Scan username-domain failed: invalid syntax of "%s"`, string(b))
	}
	return nil
}

type Preferences struct {
	Lock     bool      `json:"lock,omitempty"`
	PostVsb  utils.Vsb `json:"postVsb"`
	ShareVsb utils.Vsb `json:"shareVsb"`
}

// implement database/sql/driver.Valuer
func (p Preferences) Value() (driver.Value, error) {
	pp := p
	pp.Lock = false
	return json.Marshal(pp)
}

// implement database/sql.Scanner
func (p *Preferences) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed.")
	}
	return json.Unmarshal(b, p)
}

// models

const userJsonType = "::user"

type User struct {
	Username UD     `json:"username"`
	Foreign  bool   `json:"foreign"`
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Summary  string `json:"summary"`
	Avatar   *Img   `json:"avatar"`
	Lock     bool   `json:"lock"`
	PubKey   string `json:"pub_key,omitempty"`
	PriKey   string `json:"pri_key,omitempty"`
	Type     string `json:"__type,omitempty"` // used for cache update
}

// used for update cache
type UserCache struct {
	Nickname string `json:"nickname"`
	Summary  string `json:"summary"`
	Avatar   *Img   `json:"avatar"`
	Lock     bool   `json:"lock"`
}

// implements cache.IJsonable
func (u *UserCache) Json() (map[string]interface{}, error) {
	m, err := structToMap(u)
	if err != nil {
		return nil, err
	}
	m[jsonTypeKey] = userJsonType
	return m, nil
}

// implements cache.IJsonable
func (p *UserCache) Atomic() bool {
	return false
}

/*
type UserPf struct {
	Username    UD          `json:"username"`
	Password    string      `json:"password"`
	Preferences Preferences `json:"preferences"`
}

type ForeignInbox struct {
	Username    UD     `json:"username"`
	Inbox       string `json:"inbox"`
	SharedInbox string `json:"shared"`
}
*/

// db

type UserDb struct {
	lg     logging.Logger
	client *sql.DB
	cache  *CacheDb
}

var userIns *UserDb = nil

func UserInstance(lg logging.Logger, cl *sql.DB, cache *CacheDb) *UserDb {
	if userIns == nil {
		userIns = &UserDb{
			lg:     lg,
			client: cl,
			cache:  cache,
		}
	}
	return userIns
}
