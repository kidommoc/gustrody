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

var UsernameReg = regexp.MustCompile(`[A-z]{1}[A-z0-9]+`)
var DomainReg = regexp.MustCompile(`([A-z0-9\-]+\.)+[A-z]{2,6}`)
var UdReg = regexp.MustCompile(fmt.Sprintf(`^(%s)@(%s)$`,
	UsernameReg, DomainReg,
))

func NewUUID() string {
	return uuid.New().String()
}

// format: https://domain.site/users/username
func GenerateUserID(username, site string) string {
	return fmt.Sprintf("https://%s/users/%s", site, username)
}

// format: https://domain.site/posts/id
func GeneratePostID(id, site string) string {
	return fmt.Sprintf("https://%s/posts/%s", site, id)
}

// convert time.Time to RFC3339 string
//
// format: 2006-01-02T15:04:05Z70:00
func DateString(d time.Time) string {
	return d.Format(time.RFC3339)
}
