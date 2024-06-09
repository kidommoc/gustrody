package config

import (
	"flag"
	"strconv"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	Debug   bool   `json:"debug"`
	Site    string `json:"site"`
	Port    int    `json:"port"`
	HmacKey string `json:"hmacKey"`

	// logging
	Logfile  string `json:"logfile"`
	LogSplit int    `json:"logSplit"`
	LogLevel int    `json:"logLevel"`

	// database
	PqUser   string `json:"pqUser"`
	PqSecret string `json:"pqSecret"`
	RdSecret string `json:"rdSecret"`

	// perference
	MaxContentLength int `json:"maxCotentLength"`
}

var config *EnvConfig

func loadEnv() {
	envmap, err := godotenv.Read()
	if err != nil {
		// handle error
	}
	config = new(EnvConfig)

	config.Debug = *flag.Bool("debug", false, "debug mode")

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

	// postgresql secret. default: postgres
	pqSecret := envmap["POSTGRES_SECRET"]
	if pqSecret == "" {
		pqSecret = "postgres"
	}
	config.PqSecret = pqSecret

	// redis secret. default: redis
	rdSecret := envmap["REDIS_SECRET"]
	if rdSecret == "" {
		rdSecret = "redis"
	}
	config.RdSecret = rdSecret

	// max content length. default: 500
	mcl, err := strconv.Atoi(envmap["MAX_CONTENT_LENGTH"])
	if err != nil || mcl < 0 {
		mcl = 500
	}
	config.MaxContentLength = mcl
}

func Get() EnvConfig {
	if config == nil {
		loadEnv()
	}
	return *config
}

// USE IT CAREFULLY!
func Set(cfg EnvConfig) {
	config = &cfg
}
