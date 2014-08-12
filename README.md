slog [![GoDoc](https://godoc.org/github.com/stretchr/slog?status.png)](http://godoc.org/github.com/stretchr/slog)
====

Logging package for Go.

  * Concurrent safe
  * Three levels; `slog.Err`, `slog.Warn`, and `slog.Info`
  * Children loggers for sub-processes
  * Built-in zero-memory `slog.NilLogger` to easily logging off without changing calling code
  * Custom reporters to send logs anywhere
  * `Reporters` function for reporting to many places

### Simple

The `slog.Logger` interface is simple.

```
// make a logger that logs everything
logger := slog.New("prefix", slog.Everything)

// throughout your code, call the method with empty
// params to see whether the logger cares about that
// level or not.
if logger.Info() {
  logger.Info("something happened")
}

// errors too...
if err != nil && logger.Err() {
  logger.Err("error occurred", err)
}

// when you're finished with it - stop it
logger.Stop()
<-logger.StopChan() // wait for it to stop
```

### Different levels

If you only care about errors, use the `slog.Err` level:

```
logger := slog.New("prefix", slog.Err)
```

You can change the level at runtime:

```
logger.SetLevel(slog.Info)
```

### Children

Children loggers report their findings to the parent, and changes to the parent will also affect the children.

```
logger := slog.New("parent", slog.Everything)

childLogger := logger.New("child")
go otherProcess(childLogger)

// stopping the parent will make sure children
// are stopped too.
logger.Stop()
<-logger.StopChan() // wait for everything to stop
```

### NilLogger

If you want to disable logging entirely, the most memory efficient way to do so is to pass a `slog.NilLogger` wherever a `Logger` is needed.

```
thing := &MyThing{
  Logger: slog.NilLogger,
}

thing.DoStuff() // will do stuff without logging
```

### Custom reporting

If you want to control where the logs get reported to, you can call `SetReporter` or `SetReporterFunc` on a RootLogger.

```
logger := slog.New("prefix", slog.Info)
logger.SetReporterFunc(func(l *slog.Log) {
  // publish logs to a messaging queue
  queuePublisher.Publish("log-channel", l)
})
```

  * You can only change the `Reporter` of a RootLogger (i.e. parent), children loggers will automatically report through the specified method too.

### Multiple reporters

If you want to report logs to multiple locations, you can use the `Reporters` function.

```
logger := slog.New("prefix", slog.Info)
logger.SetReporter(slog.Reporters(slog.Stdout, msgQueueReporter, databaseReporter))
```
