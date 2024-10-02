package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	t.Logf("%+v", Get())
}
