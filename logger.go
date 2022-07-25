package gormloggerlogrus

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type Options struct {
	Logger                *logrus.Logger
	SkipErrRecordNotFound bool
	Debug                 bool
	SlowThreshold         time.Duration
	SourceField           string
}

type Logger struct {
	Options
}

func New(opts Options) *Logger {
	l := &Logger{Options: opts}
	if l.Logger == nil {
		l.Logger = logrus.New()
	}

	return l
}

func (l *Logger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, s string, args ...interface{}) {
	l.Logger.WithContext(ctx).Infof(s, args...)
}

func (l *Logger) Warn(ctx context.Context, s string, args ...interface{}) {
	l.Logger.WithContext(ctx).Warnf(s, args...)
}

func (l *Logger) Error(ctx context.Context, s string, args ...interface{}) {
	l.Logger.WithContext(ctx).Errorf(s, args...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := logrus.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[logrus.ErrorKey] = err
		l.Logger.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}
	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		l.Logger.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		return
	}
	if l.Debug {
		l.Logger.WithContext(ctx).WithFields(fields).Debugf("%s [%s]", sql, elapsed)
	}
}
