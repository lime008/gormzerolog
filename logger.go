package gormzerolog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var _ gormLogger.Interface = &logger{}

// Config can be used to configure the logger
type Config struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
}

var (
	// Default is a logger with the default configuration
	Default = New(nil, Config{
		SlowThreshold:             200 * time.Millisecond,
		IgnoreRecordNotFoundError: false,
	})
	// Recorder stores the trace recorder
	Recorder = traceRecorder{Interface: Default, BeginAt: time.Now()}
)

// New returns a new logger instance with the provided configuration
func New(zl *zerolog.Logger, config Config) gormLogger.Interface {
	if zl == nil {
		zl = &log.Logger
	}

	return &logger{
		Logger: zl,
		Config: config,
	}
}

type logger struct {
	Logger *zerolog.Logger
	Config
}

// LogMode log mode
func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

// Info print info
func (l logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info().Caller(9).Interface("data", data).Msg(msg)
}

// Warn print warn messages
func (l logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Warn().Caller(9).Interface("data", data).Msg(msg)
}

// Error print error messages
func (l logger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Error().Caller(9).Interface("data", data).Msg(msg)
}

// Trace print sql message
func (l logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	lg := l.Logger.With().CallerWithSkipFrameCount(9).Dur("elapsed", elapsed).Logger()
	var lo *zerolog.Event
	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		lo = lg.Error()
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		lo = lg.Warn().Str("slowSql", slowLog)
	default:
		lo = lg.Trace()
	}
	if rows == -1 {
		lo.Msg(sql)
		return
	}
	lo.Int64("rows", rows).Msg(sql)
}

type traceRecorder struct {
	gormLogger.Interface
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

func (l traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: l.Interface, BeginAt: time.Now()}
}

func (l *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.BeginAt = begin
	l.SQL, l.RowsAffected = fc()
	l.Err = err
}
