package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
)

// models

type User struct {
	Username    UD          `json:"username"`
	Nickname    string      `json:"nickname"`
	Summary     string      `json:"summary"`
	Avatar      Img         `json:"avatar"`
	Keys        KeyPair     `json:"keys,omitempty"`
	Preferences Preferences `json:"preferences,omitempty"`
}

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

type KeyPair struct {
	Pub string `json:"pub,omitempty"`
	Pri string `json:"pri,omitempty"`
}

// implement database/sql/driver.Valuer
func (kp KeyPair) Value() (driver.Value, error) {
	pub := strings.ReplaceAll(kp.Pub, "\n", "#n")
	pri := strings.ReplaceAll(kp.Pri, "\n", "#n")
	return fmt.Sprintf(`("%s","%s")`, pub, pri), nil
}

// implement database/sql.Scanner
func (kp *KeyPair) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("Scan key pair: src cannot cast to []byte")
	}
	fields := strings.Split(strings.Trim(string(b), "()"), ",")
	if len(fields) != 2 {
		return fmt.Errorf("Scan key pair: wrong syntax")
	}
	kp.Pub = strings.ReplaceAll(strings.Trim(fields[0], `"`), "#n", "\n")
	kp.Pri = strings.ReplaceAll(strings.Trim(fields[1], `"`), "#n", "\n")
	return nil
}

type Preferences struct {
	PostVsb  string `json:"postVsb,omitempty"`
	ShareVsb string `json:"shareVsb,omitempty"`
}

// implement database/sql/driver.Valuer
func (p Preferences) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// implement database/sql.Scanner
func (p *Preferences) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed.")
	}
	return json.Unmarshal(b, p)
}

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
