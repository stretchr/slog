// Package slog provides concurrent levelled logging capabilities
// with parent/child loggers.
//
// * Concurrent safe
// * Three levels; `slog.Err`, `slog.Warn`, and `slog.Info`
// * Children loggers for sub-processes
// * Built-in zero-memory `slog.NilLogger` to easily logging off without changing calling code
// * Custom reporters to send logs anywhere
// * `Reporters` function for reporting to many places
//
// Usage
//
// Using slog.Logger is very simple:
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
