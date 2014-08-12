package slog

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/stretchr/pat/stop"
)

// Level represents the level of logging.
type Level uint8

const (
	// Nothing represents no logging.
	Nothing Level = iota // must always be first value.

	// Err represents error level logging.
	Err
	// Warn represents warning level logging.
	Warn
	// Info represents information level logging.
	Info

	// Everything logs everything.
	Everything // must always be last value
)

// Log represents a single log item.
type Log struct {
	Level  Level
	When   time.Time
	Data   []interface{}
	Source []string
}

// Reporter represents types capable of doing something
// with logs.
type Reporter interface {
	Log(*Log)
}

// ReporterFunc is a function type capable of acting as
// a reporter.
type ReporterFunc func(*Log)

// Log calls the ReporterFunc.
func (f ReporterFunc) Log(l *Log) {
	f(l)
}

// RootLogger represents a the root Logger that has
// more capabilities than a normal Logger.
// Normally, caller code would require the Logger interface only.
type RootLogger interface {
	stop.Stopper
	Logger
	// SetReporter sets the Reporter for this logger and
	// child loggers to use.
	SetReporter(r Reporter)
	// SetReporterFunc sets the specified ReporterFunc as
	// the Reporter.
	SetReporterFunc(f ReporterFunc)
	// New creates a new child logger.
	New(source string) Logger
	// SetLevel sets the level of this and all children loggers.
	SetLevel(level Level)
}

// Logger represents types capable of logging at
// different levels.
type Logger interface {
	// Info gets whether the logger is logging information or not,
	// and also makes such logs.
	Info(a ...interface{}) bool
	// Warn gets whether the logger is logging warnings or not,
	// and also makes such logs.
	Warn(a ...interface{}) bool
	// Err gets whether the logger is logging errors or not,
	// and also makes such logs.
	Err(a ...interface{}) bool
}

type logger struct {
	m        sync.Mutex
	level    Level
	r        Reporter
	c        chan *Log
	src      []string
	stopChan chan stop.Signal
	root     *logger
}

var _ Logger = (*logger)(nil)

// New creates a new RootLogger, which is capable of acting
// like a Logger, used for logging.
// RootLogger is also a stop.Stopper and can have the
// Reporter specified, where children Logger types cannot.
func New(source string, level Level) RootLogger {
	l := &logger{
		level: level,
		src:   []string{source},
	}
	l.root = l // use this one as the root one
	l.Start()
	return l
}

// New makes a new child logger with the specified source.
func (l *logger) New(source string) Logger {
	return &logger{
		level: l.level,
		src:   append(l.src, source),
		root:  l,
	}
}

func (l *logger) SetLevel(level Level) {
	l.root.m.Lock()
	l.root.level = level
	l.root.m.Unlock()
}

func (l *logger) SetReporter(r Reporter) {
	l.root.Stop(stop.NoWait)
	<-l.root.StopChan()
	l.root.r = r
	l.root.Start()
}

func (l *logger) SetReporterFunc(f ReporterFunc) {
	l.SetReporter(f)
}

func (l *logger) Start() {
	l.root.c = make(chan *Log)
	l.root.stopChan = stop.Make()
	go func() {
		for item := range l.root.c {
			l.root.r.Log(item)
		}
	}()
}

func (l *logger) Info(a ...interface{}) bool {
	if l.skip(Info) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	l.root.c <- &Log{When: time.Now(), Data: a, Source: l.src, Level: Info}
	return true
}

func (l *logger) Warn(a ...interface{}) bool {
	if l.skip(Warn) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	l.root.c <- &Log{When: time.Now(), Data: a, Source: l.src, Level: Warn}
	return true
}

func (l *logger) Err(a ...interface{}) bool {
	if l.skip(Err) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	l.root.c <- &Log{When: time.Now(), Data: a, Source: l.src, Level: Err}
	return true
}

func (l *logger) skip(level Level) bool {
	l.root.m.Lock()
	s := l.level < level
	l.root.m.Unlock()
	return s
}

func (l *logger) Stop(time.Duration) {
	close(l.root.c)
	close(l.root.stopChan)
}

func (l *logger) StopChan() <-chan stop.Signal {
	return l.root.stopChan
}

type logReporter struct {
	logger *log.Logger
	fatal  bool
}

// NewLogReporter gets a Reporter that writes to the specified
// log.Logger.
// If fatal is true, errors will call Fatalln on the logger, otherwise
// they will always call Println.
func NewLogReporter(logger *log.Logger, fatal bool) Reporter {
	return &logReporter{logger: logger}
}

func (l *logReporter) Log(log *Log) {
	args := []interface{}{strings.Join(log.Source, "➤") + ":"}
	for _, d := range log.Data {
		args = append(args, d)
	}

	if l.fatal && log.Level == Err {
		l.logger.Fatalln(args...)
	} else {
		l.logger.Println(args...)
	}

}

// Stdout represents a reporter that writes to os.Stdout.
// Errors will also call os.Exit.
var Stdout = NewLogReporter(log.New(os.Stdout, "", log.LstdFlags), true)