package utils

import (
	"io/fs"
	"math/rand"
	"os"
	"strings"
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
