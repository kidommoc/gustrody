package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

const redisMaxRetries = 20

type Img struct {
	Type string `json:"type,omitempty"`
	Url  string `json:"url,omitempty"`
	Alt  string `json:"alt,omitempty"`
}

// implement database/sql/driver.Valuer
func (img Img) Value() (driver.Value, error) {
	t := img.Type
	u := img.Url
	a := img.Alt
	if t != "" {
		t = fmt.Sprintf(`\"%s\"`, t)
	}
	if u != "" {
		u = fmt.Sprintf(`\"%s\"`, u)
	}
	if a != "" {
		a = fmt.Sprintf(`\"%s\"`, a)
	}
	return fmt.Sprintf(`(%s,%s,%s)`, t, u, a), nil
}

// implement database/sql.Scanner
func (img *Img) Scan(src interface{}) error {
	if src == nil {
		return fmt.Errorf("nil")
	}
	b, ok := src.([]byte)
	if !ok {
		s, ok := src.(string)
		if !ok {
			return fmt.Errorf("Scan img: src cannot cast to []byte")
		}
		if s == "" {
			return nil
		}
		b = []byte(s)
	}
	fields := strings.Split(strings.Trim(string(b), "()"), ",")
	if len(fields) < 3 {
		return fmt.Errorf("Scan img: wrong fields number")
	}
	img.Type = strings.TrimSuffix(strings.TrimPrefix(fields[0], `"""`), `"""`)
	img.Url = strings.TrimSuffix(strings.TrimPrefix(fields[1], `"""`), `"""`)
	img.Alt = strings.TrimSuffix(strings.TrimPrefix(fields[2], `"""`), `"""`)
	return nil
}
