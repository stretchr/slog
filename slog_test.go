package slog_test

import (
	"bytes"
	"errors"
	golog "log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/log"
	"github.com/stretchr/pat/stop"
	"github.com/stretchr/testify/require"
)

type TestReporter struct {
	logs    []*log.Log
	logFunc func(l *log.Log)
}

func (r *TestReporter) Log(l *log.Log) {
	r.logFunc(l)
}

func NewTestReporter() *TestReporter {
	r := &TestReporter{}
	r.logFunc = func(l *log.Log) {
		r.logs = append(r.logs, l)
	}
	return r
}

func TestLog(t *testing.T) {

	l := log.New("parent", log.Err)
	defer func() {
		l.Stop(stop.NoWait)
		<-l.StopChan()
	}()

	r := NewTestReporter()
	l.SetReporter(r)

	require.False(t, l.Warn())
	require.False(t, l.Info())
	require.True(t, l.Err())
	require.True(t, l.Err("Something went", "wrong"))
	require.False(t, l.Warn("this should be ignored"))
	require.False(t, l.Info("this should be ignored"))

	require.Equal(t, 1, len(r.logs))

	require.Equal(t, "parent", r.logs[0].Source[0])
	require.Equal(t, "Something went", r.logs[0].Data[0])
	require.Equal(t, "wrong", r.logs[0].Data[1])
	require.Equal(t, log.Err, r.logs[0].Level)
	require.NotNil(t, r.logs[0].When)

}

func TestLogChildren(t *testing.T) {

	var wg sync.WaitGroup

	parent := log.New("parent", log.Info)
	defer func() {
		parent.Stop(stop.NoWait)
		<-parent.StopChan()
	}()

	child := parent.New("child")

	require.NotNil(t, child)
	r := NewTestReporter()
	f := r.logFunc
	r.logFunc = func(l *log.Log) {
		f(l)
		wg.Done()
	}
	parent.SetReporter(r)

	wg.Add(2)
	require.True(t, parent.Info("Something went", "wrong"))
	require.True(t, child.Info("something went wrong in the child too"))

	wg.Wait()

	require.Equal(t, 2, len(r.logs))

	require.Equal(t, "parent", r.logs[0].Source[0])
	require.Equal(t, "Something went", r.logs[0].Data[0])
	require.Equal(t, "wrong", r.logs[0].Data[1])
	require.Equal(t, log.Info, r.logs[0].Level)
	require.NotNil(t, r.logs[0].When)

	require.Equal(t, "parent", r.logs[1].Source[0])
	require.Equal(t, "child", r.logs[1].Source[1])
	require.Equal(t, "something went wrong in the child too", r.logs[1].Data[0])
	require.Equal(t, log.Info, r.logs[1].Level)
	require.NotNil(t, r.logs[1].When)

}

func TestLevels(t *testing.T) {

	logger := log.New("parent", log.Info)
	defer func() {
		logger.Stop(stop.NoWait)
		<-logger.StopChan()
	}()
	r := NewTestReporter()
	logger.SetReporter(r)

	logger.SetLevel(log.Nothing)
	require.False(t, logger.Info())
	require.False(t, logger.Warn())
	require.False(t, logger.Err())

	logger.SetLevel(log.Err)
	require.False(t, logger.Info())
	require.False(t, logger.Warn())
	require.True(t, logger.Err())

	logger.SetLevel(log.Warn)
	require.False(t, logger.Info())
	require.True(t, logger.Warn())
	require.True(t, logger.Err())

	logger.SetLevel(log.Info)
	require.True(t, logger.Info())
	require.True(t, logger.Warn())
	require.True(t, logger.Err())

	logger.SetLevel(log.Everything)
	require.True(t, logger.Info())
	require.True(t, logger.Warn())
	require.True(t, logger.Err())

}

func TestLogReporter(t *testing.T) {

	var buf bytes.Buffer
	logLogger := golog.New(&buf, "prefix: ", golog.LstdFlags)

	logger := log.New("parent", log.Everything)
	logger.SetReporter(log.NewLogReporter(logLogger, false))
	child := logger.New("child")
	child.Info(errors.New("message"))

	time.Sleep(500 * time.Millisecond)

	require.Contains(t, buf.String(), `message`)
	require.Contains(t, buf.String(), `parent➤child:`)
	require.Contains(t, buf.String(), `prefix:`)

}
