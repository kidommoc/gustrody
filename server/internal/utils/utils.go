package utils

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var charset = []byte("0123456789abcdef")

func GenerateRamdonHexString(n uint) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(16)]
	}
	return string(b)
}

func TrimPath(path string) string {
	return strings.TrimRight(path, "/ ")
}

func EnsureDirs(path string, isDir bool) {
	if path == "" {
		return
	}
	if !isDir {
		ps := strings.Split(path, "/")
		if len(ps) == 1 {
			return
		}
		path = strings.Join(ps[:len(ps)-1], "/")
	}
	if e := os.MkdirAll(path, fs.ModePerm); e != nil {
		panic(e.Error())
	}
}

const UsernameRegLiteral = `[A-z]{1}[A-z0-9]+`
const LocalUsernameRegLiteral = `[A-z]{1}[A-z0-9]{5,19}`

// support domain or ip addr with/without port
const DomainRegLiteral = `(([A-z0-9\-]+\.)+[A-z]{2,6}|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})(:\d{1,5})?`

var UsernameReg = regexp.MustCompile(UsernameRegLiteral)
var LocalUsernameReg = regexp.MustCompile(LocalUsernameRegLiteral)
var DomainReg = regexp.MustCompile(DomainRegLiteral)
var UdReg = regexp.MustCompile(fmt.Sprintf(`^(%s)@(%s)$`,
	UsernameReg, DomainReg,
))

func NewUUID() string {
	return uuid.New().String()
}

// format: https://domain.site/users/username
func UserIDReg(scheme, domain string) *regexp.Regexp {
	if !DomainReg.MatchString(domain) {
		return nil
	}
	domain = strings.ReplaceAll(domain, `.`, `\.`)
	literal := fmt.Sprintf(`%s://%s/users/(%s)`, scheme, domain, LocalUsernameRegLiteral)
	return regexp.MustCompile(literal)
}

// format: scheme://domain/users/username
func GenerateUserID(username, scheme, domain string) string {
	return fmt.Sprintf("%s://%s/users/%s", scheme, domain, username)
}

const UUIDRegLiteral = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`

func PostIDReg(scheme, domain string) *regexp.Regexp {
	if !DomainReg.MatchString(domain) {
		return nil
	}
	domain = strings.ReplaceAll(domain, `.`, `\.`)
	literal := fmt.Sprintf(`%s://%s/posts/(%s)`, scheme, domain, UUIDRegLiteral)
	return regexp.MustCompile(literal)
}

// format: scheme://domain/posts/id
func GeneratePostID(id, scheme, domain string) string {
	return fmt.Sprintf("%s://%s/posts/%s", scheme, domain, id)
}

// convert time.Time to RFC3339 string. format:
//
//	2006-01-02T15:04:05Z // timezone 0
//	2006-01-02T15:04:05(+|-)07:00 // timezone +/-
func DateString(d time.Time) string {
	return d.UTC().Format(time.RFC3339)
}

// convert RFC3339 string to time.Time. format:
//
//	2006-01-02T15:04:05Z // timezone 0
//	2006-01-02T15:04:05(+|-)07:00 // timezone +/-
func ParseDateString(s string) *time.Time {
	d, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return &d
	} else {
		return nil
	}
}

// convert time.Time to RFC3339 string. format:
//
//	Mon, 02 Jan 2006 15:04:05 MST
func HttpDateString(d time.Time) string {
	return strings.ReplaceAll(d.UTC().Format(time.RFC1123), "UTC", "MST")
}

// convert RFC3339 string to time.Time. format:
//
//	Mon, 02 Jan 2006 15:04:05 MST
func ParseHttpDateString(s string) *time.Time {
	d, err := time.Parse(time.RFC1123, s)
	if err != nil {
		return &d
	} else {
		return nil
	}
}
