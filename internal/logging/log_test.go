package logging

import (
	"testing"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/utils"
)

type T struct {
	Field1 string  `json:"field1"`
	Field2 float32 `json:"field2"`
}

type E struct {
	utils.Err
}

const errcode utils.ErrCode = 0

func (e E) CodeString() string {
	return "errcode"
}

func TestShenllLogger(t *testing.T) {
	cfg := config.EnvConfig{
		Logfile:  "logging.log",
		LogSplit: 0,
		LogLevel: 3,
	}
	o := &T{
		Field1: "test",
		Field2: 42.42,
	}
	logger := Get(cfg)
	logger.level = 3
	logger.Info("This is an info message")
	logger.Info("This is an info with attachments",
		"id", "abcdef",
		"name", "xyzabc",
		"age", 572,
		"obj", o,
	)

	e := E{
		Err: utils.NewErr(errcode, "errmsg"),
	}
	logger.Error("This is an error message", e)
}

func TestFileLogger(t *testing.T) {
	cfg := config.EnvConfig{
		Logfile:  "logging.log",
		LogSplit: 0,
		LogLevel: 0,
	}
	o := &T{
		Field1: "test",
		Field2: 42.42,
	}
	logger := Get(cfg)
	logger.level = 0
	logger.Info("This is an info message")
	logger.Info("This is an info with attachments",
		"id", "abcdef",
		"name", "xyzabc",
		"age", 572,
		"obj", o,
	)

	e := E{
		Err: utils.NewErr(errcode, "errmsg"),
	}
	logger.Error("This is an error message", e)
}
