package logger

import (
	"context"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
	"os"

	"github.com/fatih/color"
)

type Log struct {
	Logger *slog.Logger
}

func (l *Log) Initialize(level string) *slog.Logger {
	const (
		info    = "INFO"
		debug   = "DEBUG"
		warning = "WARNING"
		errorL  = "ERROR"
	)
	var logLevel slog.Level
	switch level {
	case info:
		logLevel = slog.LevelInfo
	case debug:
		logLevel = slog.LevelDebug
	case warning:
		logLevel = slog.LevelWarn
	case errorL:
		logLevel = slog.LevelError
	}
	opts := PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: logLevel,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)
	l.Logger = slog.New(handler)

	return l.Logger
}

func (l *Log) Fatal(v ...any) {
	l.Logger.Error(fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Log) Err(msg string, value interface{}) {
	l.Logger.Error(msg, slog.String("err", fmt.Sprintf("%v", value)))
}

func (l *Log) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Log) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l *Log) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewTextHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}

	return h
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error { //nolint:gocritic //huge param passed
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var msg string

	if len(fields) > 0 {
		for k, v := range fields {
			msg += fmt.Sprintf("%s=%v ", k, v)
		}
	}

	timeStr := r.Time.Format("[15:04:05.000]")
	msg = color.CyanString(r.Message) + " " + msg

	h.l.Println(
		timeStr,
		level,
		msg,
	)

	return nil
}
