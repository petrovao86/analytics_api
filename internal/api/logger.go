package api

import (
	"context"
	"fmt"
	"io"
	"slices"

	fiberlog "github.com/gofiber/fiber/v2/log"
	log "github.com/sirupsen/logrus"
)

type logger struct {
	l *log.Logger
}

var _ fiberlog.AllLogger = (*logger)(nil)

func (l *logger) newEntryW(args ...any) *log.Entry {
	e := log.NewEntry(l.l)
	if len(args) > 0 {
		if (len(args) & 1) == 1 {
			args = append(args, "__KEYVALS_UNPAIRED__")
		}

		for i := 0; i < len(args); i += 2 {
			e = e.WithField(fmt.Sprintf("%s", args[i]), args[i+1])
		}
	}
	return e
}

func (l *logger) Debug(args ...any) {
	l.l.Debug(args...)
}

func (l *logger) Debugf(format string, args ...any) {
	l.l.Debugf(format, args...)
}

func (l *logger) Debugw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Debug(format)
}

func (l *logger) Error(args ...any) {
	l.l.Error(args...)
}

func (l *logger) Errorf(format string, args ...any) {
	l.l.Errorf(format, args...)
}

func (l *logger) Errorw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Error(format)
}

func (l *logger) Fatal(args ...any) {
	l.l.Fatal(args...)
}

func (l *logger) Fatalf(format string, args ...any) {
	l.l.Fatalf(format, args...)
}

func (l *logger) Fatalw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Fatal(format)
}

func (l *logger) Info(args ...any) {
	l.l.Info(args...)
}

func (l *logger) Infof(format string, args ...any) {
	l.l.Infof(format, args...)
}

func (l *logger) Infow(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Info(format)
}

func (l *logger) Panic(args ...any) {
	l.l.Panic(args...)
}

func (l *logger) Panicf(format string, args ...any) {
	l.l.Panicf(format, args...)
}

func (l *logger) Panicw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Panic(format)
}

func (l *logger) Trace(args ...any) {
	l.l.Trace(args...)
}

func (l *logger) Tracef(format string, args ...any) {
	l.l.Tracef(format, args...)
}

func (l *logger) Tracew(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Trace(format)
}

func (l *logger) Warn(args ...any) {
	l.l.Warn(args...)
}

func (l *logger) Warnf(format string, args ...any) {
	l.l.Warnf(format, args...)
}

func (l *logger) Warnw(format string, args ...any) {
	e := l.newEntryW(args...)
	e.Warn(format)
}

func (l *logger) SetLevel(lvl fiberlog.Level) {
	levels := slices.Clone(log.AllLevels)
	slices.Reverse(levels)
	l.l.SetLevel(levels[lvl])
}

func (l *logger) SetOutput(out io.Writer) {
	l.l.SetOutput(out)
}

func (l *logger) WithContext(ctx context.Context) fiberlog.CommonLogger {
	return l
}
