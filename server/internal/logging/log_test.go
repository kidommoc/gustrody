package logging

import (
	"fmt"
	"testing"

	"github.com/kidommoc/gustrody/internal/config"
)

type T struct {
	Field1 string  `json:"field1"`
	Field2 float32 `json:"field2"`
}

func TestShenllLogger(t *testing.T) {
	cfg := config.Config{
		Logfile:  "logging.log",
		LogSplit: 0,
		LogLevel: 3,
	}
	o := &T{
		Field1: "test",
		Field2: 42.42,
	}
	logger, _ := Get(cfg).(*logger)
	logger.level = 3
	logger.Info("This is an info message")
	logger.Info("This is an info with attachments",
		"id", "abcdef",
		"name", "xyzabc",
		"age", 572,
		"obj", o,
	)

	logger.Error("This is an error message", fmt.Errorf("errmsg"))
}

func TestFileLogger(t *testing.T) {
	cfg := config.Config{
		Logfile:  "logging.log",
		LogSplit: 0,
		LogLevel: 0,
	}
	o := &T{
		Field1: "test",
		Field2: 42.42,
	}
	logger, _ := Get(cfg).(*logger)
	logger.level = 0
	logger.Info("This is an info message")
	logger.Info("This is an info with attachments",
		"id", "abcdef",
		"name", "xyzabc",
		"age", 572,
		"obj", o,
	)

	logger.Error("This is an error message", fmt.Errorf("errmsg"))
}
