package slog_test

import (
	"bytes"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/pat/stop"
	"github.com/stretchr/slog"
	"github.com/stretchr/testify/require"
)

type TestReporter struct {
	logs    []*slog.Log
	logFunc func(l *slog.Log)
}

func (r *TestReporter) Log(l *slog.Log) {
	r.logFunc(l)
}

func NewTestReporter() *TestReporter {
	r := &TestReporter{}
	r.logFunc = func(l *slog.Log) {
		r.logs = append(r.logs, l)
	}
	return r
}

func TestLog(t *testing.T) {

	l := slog.New("parent", slog.Err)
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
	require.Equal(t, slog.Err, r.logs[0].Level)
	require.NotNil(t, r.logs[0].When)

}

func TestSetSource(t *testing.T) {

	var wg sync.WaitGroup

	p := slog.New("parent", slog.Err)
	l := p.New("child")
	defer func() {
		p.Stop(stop.NoWait)
		<-p.StopChan()
	}()

	r := NewTestReporter()
	f := r.logFunc
	r.logFunc = func(l *slog.Log) {
		f(l)
		wg.Done()
	}
	p.SetReporter(r)

	wg.Add(1)
	l.Err("test")
	wg.Wait()

	require.Equal(t, 1, len(r.logs))
	require.Equal(t, "parent", r.logs[0].Source[0])
	require.Equal(t, "child", r.logs[0].Source[1])

	l.SetSource("new-source")
	wg.Add(1)
	l.Err("test")
	wg.Wait()

	require.Equal(t, 2, len(r.logs))
	require.Equal(t, "parent", r.logs[0].Source[0])
	require.Equal(t, "new-source", r.logs[1].Source[1])

}

func TestLogChildren(t *testing.T) {

	var wg sync.WaitGroup

	parent := slog.New("parent", slog.Info)
	defer func() {
		parent.Stop(stop.NoWait)
		<-parent.StopChan()
	}()

	child := parent.New("child")
	grandchild := child.New("grandchild")

	require.NotNil(t, child)
	r := NewTestReporter()
	f := r.logFunc
	r.logFunc = func(l *slog.Log) {
		f(l)
		wg.Done()
	}
	parent.SetReporter(r)

	wg.Add(3)
	require.True(t, parent.Info("Something went", "wrong"))
	require.True(t, child.Info("something went wrong in the child too"))
	require.True(t, grandchild.Info("something went wrong in the grandchild too"))
	wg.Wait()

	require.Equal(t, 3, len(r.logs))

	require.Equal(t, "parent", r.logs[0].Source[0])
	require.Equal(t, "Something went", r.logs[0].Data[0])
	require.Equal(t, "wrong", r.logs[0].Data[1])
	require.Equal(t, slog.Info, r.logs[0].Level)
	require.NotNil(t, r.logs[0].When)

	require.Equal(t, "parent", r.logs[1].Source[0])
	require.Equal(t, "child", r.logs[1].Source[1])
	require.Equal(t, "something went wrong in the child too", r.logs[1].Data[0])
	require.Equal(t, slog.Info, r.logs[1].Level)
	require.NotNil(t, r.logs[1].When)

	require.Equal(t, "parent", r.logs[2].Source[0])
	require.Equal(t, "child", r.logs[2].Source[1])
	require.Equal(t, "grandchild", r.logs[2].Source[2])
	require.Equal(t, "something went wrong in the grandchild too", r.logs[2].Data[0])
	require.Equal(t, slog.Info, r.logs[2].Level)
	require.NotNil(t, r.logs[2].When)

}

func TestLevels(t *testing.T) {

	logger := slog.New("parent", slog.Info)
	defer func() {
		logger.Stop(stop.NoWait)
		<-logger.StopChan()
	}()
	r := NewTestReporter()
	logger.SetReporter(r)

	logger.SetLevel(slog.Nothing)
	require.False(t, logger.Info())
	require.False(t, logger.Warn())
	require.False(t, logger.Err())

	logger.SetLevel(slog.Err)
	require.False(t, logger.Info())
	require.False(t, logger.Warn())
	require.True(t, logger.Err())

	logger.SetLevel(slog.Warn)
	require.False(t, logger.Info())
	require.True(t, logger.Warn())
	require.True(t, logger.Err())

	logger.SetLevel(slog.Info)
	require.True(t, logger.Info())
	require.True(t, logger.Warn())
	require.True(t, logger.Err())

	logger.SetLevel(slog.Everything)
	require.True(t, logger.Info())
	require.True(t, logger.Warn())
	require.True(t, logger.Err())

}

func TestLogReporter(t *testing.T) {

	var buf bytes.Buffer
	logLogger := log.New(&buf, "prefix: ", log.LstdFlags)

	logger := slog.New("parent", slog.Everything)
	logger.SetReporter(slog.NewLogReporter(logLogger, false))
	child := logger.New("child")
	child.Info(errors.New("message"))

	time.Sleep(500 * time.Millisecond)

	require.Contains(t, buf.String(), `message`)
	require.Contains(t, buf.String(), "parent»child:")
	require.Contains(t, buf.String(), `prefix:`)

}

func TestReporterFunc(t *testing.T) {

	l := slog.New("parent", slog.Err)
	defer func() {
		l.Stop(stop.NoWait)
		<-l.StopChan()
	}()

	var logs []*slog.Log
	l.SetReporterFunc(func(l *slog.Log) {
		logs = append(logs, l)
	})

	require.False(t, l.Warn())
	require.False(t, l.Info())
	require.True(t, l.Err())
	require.True(t, l.Err("Something went", "wrong"))
	require.False(t, l.Warn("this should be ignored"))
	require.False(t, l.Info("this should be ignored"))

	require.Equal(t, 1, len(logs))

	require.Equal(t, "parent", logs[0].Source[0])
	require.Equal(t, "Something went", logs[0].Data[0])
	require.Equal(t, "wrong", logs[0].Data[1])
	require.Equal(t, slog.Err, logs[0].Level)
	require.NotNil(t, logs[0].When)

}

func TestMultipleReporters(t *testing.T) {

	var logs1 []*slog.Log
	r1 := slog.ReporterFunc(func(l *slog.Log) {
		logs1 = append(logs1, l)
	})
	var logs2 []*slog.Log
	r2 := slog.ReporterFunc(func(l *slog.Log) {
		logs2 = append(logs2, l)
	})
	var logs3 []*slog.Log
	r3 := slog.ReporterFunc(func(l *slog.Log) {
		logs3 = append(logs3, l)
	})

	rs := slog.Reporters(r1, r2, r3)
	l := &slog.Log{}
	rs.Log(l)

	require.Equal(t, 1, len(logs1))
	require.Equal(t, 1, len(logs2))
	require.Equal(t, 1, len(logs3))
	require.Equal(t, l, logs1[0])
	require.Equal(t, l, logs2[0])
	require.Equal(t, l, logs3[0])

}
