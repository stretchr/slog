// Package slog provides concurrent levelled logging capabilities
// with parent/child loggers.
//
// Example
//
//     // make a logger to output to stdout
//     l := slog.New("parent", slog.Warn)
//     l.SetReporter(slog.Stdout)
//     if l.Info() {
//       l.Info("This is some information")
//     }
//     if l.Err() {
//       l.Err("failed to do something:", err)
//     }
//
//     // start a sub process
//     cl := l.New("child")
//     StartSubProcess(cl)
//     // reports as parent>child
//
//     // stop the logger
//     l.Stop()
//     <-l.StopChan() // wait for it to stop
package slog
