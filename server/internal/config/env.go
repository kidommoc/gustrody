package config

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
)

func loadEnv() {
	envmap, err := godotenv.Read()
	if err != nil {
		// handle error
	}
	config = new(Config)

	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	config.Debug = *debug
	if config.Debug {
		fmt.Println("Start in debug mode.")
	}

	site := envmap["SITE"]
	// check site
	config.Site = site

	// port. default 8000
	port, err := strconv.Atoi(envmap["PORT"])
	if err != nil || port < 0 {
		port = 8000
	}
	config.Port = port

	// used in encryption. default: penguin
	hmacKey := envmap["HMAC_KEY"]
	if hmacKey == "" {
		hmacKey = "penguin"
	}
	config.HmacKey = hmacKey

	// logfile path. default: "./logging.log"
	logfile := envmap["LOGFILE"]
	if logfile == "" {
		logfile = "./logging.log"
	}
	config.Logfile = logfile

	// how to split logfile. default: 0(none)
	logSplit, e := strconv.Atoi(envmap["LOG_SPLIT"])
	if e != nil || logSplit < 0 || logSplit > 1 {
		logSplit = 0
	}
	config.LogSplit = logSplit

	// log level. default: 1(warning)
	logLevel, e := strconv.Atoi(envmap["LOG_LEVEL"])
	if e != nil || logLevel < 0 || logLevel > 3 {
		logLevel = 1
	}
	config.LogLevel = logLevel

	// postgresql user. default: penguin
	pqUser := envmap["POSTGRES_USER"]
	if pqUser == "" {
		pqUser = "penguin"
	}
	config.PqUser = pqUser

	secret_paths[db_main] = envmap["POSTGRES_SECRET_PATH"]
	secret_paths[db_redis] = envmap["REDIS_SECRET_PATH"]

	if len(secrets) == 0 {
		loadSecrets()
	}

	config.PqSecret = secrets[db_main]
	config.RdSecret = secrets[db_redis]

	// max content length. default: 500
	mcl, err := strconv.Atoi(envmap["MAX_CONTENT_LENGTH"])
	if err != nil || mcl < 0 {
		mcl = 500
	}
	config.MaxContentLength = mcl
}
