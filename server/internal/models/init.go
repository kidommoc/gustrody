package models

import (
	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/logging"
)

func Init() {
	logger := logging.Get()
	cfg := config.Get()
	auth := AuthInstance(logger)
	logger.Info("[Models] Initailized AuthDb")
	if cfg.Debug {
		registerTestUsers(auth)
		logger.Debug("[Models] registered test user accounts")
	}
	UserInstance(logger)
	logger.Info("[Models] Initailized UserDb")
	PostInstance(logger)
	logger.Info("[Models] Initailized PostDb")
}

func registerTestUsers(db IAuthDb) {
	logger := logging.Get()
	e := db.SetUserPassword("u1", "penguin")
	if e != nil {
		logger.Error("when init u1", e)
	}
	e = db.SetUserPassword("u2", "penguin")
	if e != nil {
		logger.Error("when init u1", e)
	}
	e = db.SetUserPassword("u3", "penguin")
	if e != nil {
		logger.Error("when init u1", e)
	}
}
