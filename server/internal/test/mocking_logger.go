package test

import (
	"fmt"
	"testing"
)

type MockingLogger struct {
	t *testing.T
}

func NewMockingLogger(t *testing.T) *MockingLogger {
	return &MockingLogger{t: t}
}

func mapping(args ...any) map[string]string {
	m := make(map[string]string)
	key := ""
	for i, v := range args {
		if i%2 == 1 {
			key = fmt.Sprint(v)
		} else {
			value := fmt.Sprint(v)
			m[key] = value
			key = ""
		}
	}
	if key != "" {
		m["!END"] = key
	}
	return m
}

func (l *MockingLogger) Debug(msg string, attach ...any) {
	l.t.Log(msg, mapping(attach...))
}

func (l *MockingLogger) Info(msg string, attach ...any) {
	l.t.Log(msg, mapping(attach...))
}

func (l *MockingLogger) Warning(msg string, attach ...any) {
	l.t.Log(msg, mapping(attach...))
}

func (l *MockingLogger) Error(msg string, err error) {
	l.t.Log(msg, err)
}
