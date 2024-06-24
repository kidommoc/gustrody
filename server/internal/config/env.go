package config

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/kidommoc/gustrody/internal/utils"
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
	config.Site = utils.TrimPath(site)

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
	utils.EnsureDirs(logfile, false)

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

	secret_paths[db_main] = utils.TrimPath(envmap["POSTGRES_SECRET_PATH"])
	secret_paths[db_redis] = utils.TrimPath(envmap["REDIS_SECRET_PATH"])

	if len(secrets) == 0 {
		loadSecrets()
	}

	config.PqSecret = secrets[db_main]
	config.RdSecret = secrets[db_redis]

	// directory of user-uploaded images
	imgDir := utils.TrimPath(envmap["IMAGES_DIR"])
	if imgDir == "" {
		imgDir = "./data/imgs"
	}
	config.ImgDir = imgDir
	utils.EnsureDirs(imgDir, true)

	// max content length. default: 500
	mcl, err := strconv.Atoi(envmap["MAX_CONTENT_LENGTH"])
	if err != nil || mcl < 1 {
		mcl = 500
	}
	config.MaxContentLength = mcl

	// max image in post. default: 4
	mip, err := strconv.Atoi(envmap["MAX_IMG_IN_POST"])
	if err != nil || mcl < 1 {
		mip = 4
	}
	config.MaxImgInPost = mip
}
