package models

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
	"github.com/kidommoc/gustrody/internal/utils"
	"github.com/redis/go-redis/v9"
)

var defaultCtx = context.Background()

func Init() {
	logger := logging.Get()
	cfg := config.Get()

	authClient := initRedis(cfg, logger, redisOpt{
		Addr:    "localhost:6739", // will change after apply docker compose
		Db:      0,
		MaxConn: 3,
	})
	auth := AuthInstance(logger, authClient)
	logger.Info("[Models] Initailized AuthDb")
	if cfg.Debug {
		registerTestUsers(auth)
		logger.Debug("[Models] registered test user accounts")
	}

	cacheClient := initRedis(cfg, logger, redisOpt{
		Addr:    "localhost:6738",
		Db:      0,
		MaxConn: 10,
	})
	cache := CacheInstance(logger, cacheClient)
	logger.Info("[Models] Initailized AuthDb")

	mainClient := initMainDb(cfg, logger, pqOpt{
		Addr:    "localhost:5432", // will change after apply docker compose
		MaxConn: 10,
	})
	UserInstance(logger, mainClient, cache)
	logger.Info("[Models] Initailized UserDb")
	PostInstance(logger, mainClient, cache)
	logger.Info("[Models] Initailized PostDb")
}

func registerTestUsers(db IAuthDb) {
	logger := logging.Get()
	pwd := string(utils.SHA256Hash("penguin"))
	e := db.SetUserPassword("u1", pwd)
	if e != nil {
		logger.Error("when init u1", e)
	}
	e = db.SetUserPassword("u2", pwd)
	if e != nil {
		logger.Error("when init u1", e)
	}
	e = db.SetUserPassword("u3", pwd)
	if e != nil {
		logger.Error("when init u1", e)
	}
}

type pqOpt struct {
	Addr    string
	MaxConn int
}

func initMainDb(cfg config.Config, logger logging.Logger, opt pqOpt) *sql.DB {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/austrody?sslmode=disable", // db name: austrody
		cfg.PqUser, cfg.PqSecret,
		opt.Addr,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("[Db.Postgres] Cannot create Postgres client", nil)
		panic("Cannot create Postgres client")
	}
	db.SetMaxOpenConns(2 * opt.MaxConn)
	return db
}

type redisOpt struct {
	Addr    string
	Db      int
	MaxConn int
}

func initRedis(cfg config.Config, logger logging.Logger, opt redisOpt) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     opt.Addr,
		Password: cfg.RdSecret,
		DB:       opt.Db,
		PoolSize: 2 * opt.MaxConn,
	})
	if client == nil {
		logger.Error("[Db.Redis] Cannot create Redis client", nil)
		panic("Cannot create Redis client")
	}
	return client
}

func sqlExec(result sql.Result, e error) (affected int64, err error) {
	if e != nil {
		return 0, e
	}
	return result.RowsAffected()
}
