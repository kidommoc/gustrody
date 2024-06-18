package config

import (
	"fmt"
	"os"
)

type db_name string

const (
	db_main  db_name = "pg"
	db_redis db_name = "rd"
)

var secret_paths map[db_name]string = make(map[db_name]string)
var secrets map[db_name]string = make(map[db_name]string)

// default postgresql secret: postgres
// default redis secret: redis
var default_secrets map[db_name]string = map[db_name]string{
	db_main:  "postgres",
	db_redis: "redis",
}

func loadSecrets() {
	for k, v := range secret_paths {
		secret, e := os.ReadFile(v)
		if e == nil || len(secret) == 0 {
			secrets[k] = string(secret)
		} else {
			secrets[k] = default_secrets[k]
		}
		fmt.Printf("Load secret from %s\n", v)
	}
}
