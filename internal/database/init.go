package database

import (
	"github.com/kidommoc/gustrody/internal/logging"
)

func Init() {
	logger := logging.Get()
	initAuthDb()
	logger.Info("[DB]Initailized AuthDb")
	initUserDb()
	logger.Info("[DB]Initailized UserDb")
	initPostDb()
	logger.Info("[DB]Initailized PostDb")
}
