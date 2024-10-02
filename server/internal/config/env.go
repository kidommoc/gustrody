package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/kidommoc/gustrody/internal/utils"
)

func loadEnv() {
	config = new(Config)

	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	config.Debug = *debug
	if config.Debug {
		fmt.Println("Start in debug mode.")
	}

	// site domain
	config.Domain = os.Getenv("DOMAIN")
	// check site
	if !utils.DomainReg.MatchString(config.Domain) {
		msg := fmt.Sprintf("Your site domain is illegal! regex:\n    %s", utils.DomainReg.String())
		panic(msg)
	}
	config.Domain = utils.TrimPath(config.Domain)

	// site scheme. http or https
	config.Scheme = os.Getenv("SCHEME")
	switch config.Scheme {
	case "http":
	case "https":
	default:
		config.Scheme = "http"
	}

	// port. default 8000
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port < 0 {
		port = 8000
	}

	// used in encryption. default: penguin
	config.HmacKey = os.Getenv("HMAC_KEY")
	if config.HmacKey == "" {
		config.HmacKey = "penguin"
	}

	// logfile path. default: "./logging.log"
	config.Logfile = os.Getenv("LOGFILE")
	if config.Logfile == "" {
		config.Logfile = "./logging.log"
	}
	utils.EnsureDirs(config.Logfile, false)

	// how to split logfile. default: 0(none)
	logSplit, e := strconv.Atoi(os.Getenv("LOG_SPLIT"))
	if e != nil || logSplit < 0 || logSplit > 1 {
		logSplit = 0
	}
	config.LogSplit = logSplit

	// log level. default: 1(warning)
	logLevel, e := strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if e != nil || logLevel < 0 || logLevel > 3 {
		logLevel = 1
	}
	config.LogLevel = logLevel

	// postgresql user. default: penguin
	config.PqUser = os.Getenv("POSTGRES_USER")
	if config.PqUser == "" {
		config.PqUser = "penguin"
	}

	sp := map[db_name]string{
		db_main:  utils.TrimPath(os.Getenv("POSTGRES_SECRET_PATH")),
		db_redis: utils.TrimPath(os.Getenv("REDIS_SECRET_PATH")),
	}

	if len(secrets) == 0 {
		loadSecrets(sp)
	}

	config.PqSecret = secrets[db_main]
	config.RdSecret = secrets[db_redis]

	// directory of user-uploaded images
	config.ImgDir = utils.TrimPath(os.Getenv("IMAGES_DIR"))
	if config.ImgDir == "" {
		config.ImgDir = "./data/imgs"
	}
	utils.EnsureDirs(config.ImgDir, true)

	// max content length. default: 500
	mcl, err := strconv.Atoi(os.Getenv("MAX_CONTENT_LENGTH"))
	if err != nil || mcl < 1 {
		mcl = 500
	}
	config.MaxContentLength = mcl

	// max image in post. default: 4
	mip, err := strconv.Atoi(os.Getenv("MAX_IMG_IN_POST"))
	if err != nil || mcl < 1 {
		mip = 4
	}
	config.MaxImgInPost = mip

	allowRegistry := os.Getenv("ALLOW_REGISTRY")
	if allowRegistry == "true" {
		config.AllowRegistry = true
	} else {
		config.AllowRegistry = false
	}
}
