package models

import (
	"github.com/kidommoc/gustrody/internal/config"
)

var modelscfg = config.Config{
	Debug:    true,
	PqUser:   "penguin",
	PqSecret: "postgres",
	RdSecret: "redis",
}
