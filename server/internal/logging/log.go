package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kidommoc/gustrody/internal/config"
)

type logLevel uint
type logSplit uint

const (
	level_info logLevel = iota
	level_warning
	level_error
	level_all
	split_none logSplit = iota
	split_date
)

type color int

const (
	reset              = "\x1b[0m"
	color_red    color = 31
	color_green  color = 32
	color_yellow color = 33
	color_blue   color = 34
	color_light  color = 37
	color_gray   color = 90
	color_white  color = 97
)

var levels = []logLevel{level_info, level_warning, level_error, level_all}
var splits = []logSplit{split_none, split_date}
var timeFormat = "[15:04:05.000]"

// shell handler

type shellHandler struct {
	slog.Handler
	w io.Writer
}

func newShellHandler(opts *slog.HandlerOptions) *shellHandler {
	h := &shellHandler{
		Handler: slog.NewTextHandler(os.Stdout, opts),
		w:       os.Stdout,
	}
	return h
}

func colorize(s string, c color) string {
	// not work in windows git bash.
	// return fmt.Sprintf("\x1b[%dm%s%s", int(c), s, reset)
	return s
}

func (h *shellHandler) Handle(ctx context.Context, r slog.Record) error {
	time := colorize(r.Time.Format(timeFormat), color_gray)
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = colorize(level, color_green)
	case slog.LevelInfo:
		level = colorize(level, color_blue)
	case slog.LevelWarn:
		level = colorize(level, color_yellow)
	case slog.LevelError:
		level = colorize(level, color_red)
	}
	msg := colorize(r.Message, color_white)

	buf := fmt.Sprint(time, " ", level, " ", msg)
	if r.NumAttrs() == 0 {
		buf = fmt.Sprintln(buf)
	} else {
		attach := make(map[string]interface{}, r.NumAttrs())
		r.Attrs(func(attr slog.Attr) bool {
			attach[attr.Key] = attr.Value.Any()
			return true
		})
		b, _ := json.MarshalIndent(attach, "", "  ")
		buf = fmt.Sprintln(buf, colorize(string(b), color_light))
	}
	fmt.Fprint(h.w, buf)
	return nil
}

// logger

type Logger interface {
	Debug(msg string, attach ...any)
	Info(msg string, attach ...any)
	Warning(msg string, attach ...any)
	Error(msg string, err error)
}

type logger struct {
	path        string
	date        string
	level       logLevel
	split       logSplit
	opts        *slog.HandlerOptions
	shellLogger *slog.Logger
	fileLogger  *slog.Logger
	logfile     *os.File
}

var instance *logger = nil

func (l *logger) update() {
	if l.split == split_none && l.fileLogger != nil { // don't split, update is not required
		return
	}
	date := time.Now().Format(time.DateOnly)
	if date == l.date && l.fileLogger != nil { // same date, update is not required
		return
	}

	// close old logfile
	if l.logfile != nil {
		l.logfile.Close()
	}

	// open new logfile
	l.date = date
	path := l.path
	fp := strings.Contains(path, "%s")
	switch l.split {
	case split_date:
		if fp {
			path = fmt.Sprintf(path, "_"+date)
		} else {
			path = path + "_" + date
		}
	case split_none:
		if fp {
			path = fmt.Sprintf(path, "")
		}
	}
	var e error
	l.logfile, e = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fs.ModePerm)
	if e != nil {
		msg := fmt.Sprintf("Cannot open logfile \"%s\":\n%s", path, e.Error())
		panic(msg)
	}

	// update fileLogger
	l.fileLogger = slog.New(slog.NewJSONHandler(
		l.logfile,
		l.opts,
	))
}

func Get(c ...config.Config) Logger {
	if instance != nil {
		return instance
	}

	var cfg config.Config
	if len(c) == 0 {
		cfg = config.Get()
	} else {
		cfg = c[0]
	}
	l := logger{
		path:  cfg.Logfile,
		level: levels[cfg.LogLevel],
		split: splits[cfg.LogSplit],
	}

	// set debug
	l.opts = &slog.HandlerOptions{
		AddSource: cfg.Debug,
	}
	if cfg.Debug {
		l.opts.Level = slog.LevelDebug
	} else {
		l.opts.Level = slog.LevelInfo
	}

	l.shellLogger = slog.New(newShellHandler(l.opts))
	l.update() // set file logger

	instance = &l
	return instance
}

func handleAttrs(attrs ...any) []any {
	if len(attrs) > 1 && len(attrs)%2 != 0 {
		attrs = attrs[:len(attrs)-1]
	}
	return attrs
}

func (l *logger) Debug(msg string, attach ...any) {
	// shell
	l.shellLogger.Debug(msg, handleAttrs(attach...)...)
}

func (l *logger) Info(msg string, attach ...any) {
	if l.level <= level_info { // file
		l.update()
		l.fileLogger.Info(msg, handleAttrs(attach...)...)
	} else { // shell
		l.shellLogger.Info(msg, handleAttrs(attach...)...)
	}
}

func (l *logger) Warning(msg string, attach ...any) {
	if l.level <= level_warning {
		l.update()
		l.fileLogger.Warn(msg, handleAttrs(attach...)...)
	} else {
		l.shellLogger.Warn(msg, handleAttrs(attach...)...)
	}
}

func (l *logger) Error(msg string, err error) {
	if l.level <= level_error {
		l.update()
		if err == nil {
			l.fileLogger.Error(msg)
		} else {
			l.fileLogger.Error(msg,
				slog.String("msg", err.Error()),
			)
		}
	} else {
		if err == nil {
			l.shellLogger.Error(msg)
		} else {
			l.shellLogger.Error(msg,
				slog.String("msg", err.Error()),
			)
		}
	}
}
