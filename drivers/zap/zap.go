package zap

import (
	"fmt"

	"github.com/Chekunin/logman"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const DriverName = "zap"

const zapCriticalLevel = 99

type driver struct{}

func (d driver) CreateLogger(
	lm *logman.Logman,
	c logman.ChannelConfig,
) (logman.Logger, error) {
	cfg, err := parseConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config: %w", err)
	}

	return newLogger(cfg, lm)
}

type logger struct {
	cfg    LoggerConfig
	logger *zap.Logger
}

func newLogger(cfg LoggerConfig, lm *logman.Logman) (*logger, error) {
	if err := cfg.setDefaults(lm).validate(lm); err != nil {
		return nil, fmt.Errorf("Invalid config: %w", err)
	}

	zapLogger, err := zap.Config{
		Level:             zap.NewAtomicLevelAt(toZapLevel(cfg.Level)),
		Development:       false,
		Sampling:          nil,
		DisableStacktrace: !cfg.EnableStackTrace,
		Encoding:          cfg.Encoding,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapCriticalAwareLowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		DisableCaller:    !cfg.EnableCaller,
		OutputPaths:      cfg.Output,
		ErrorOutputPaths: []string{"stderr"},
	}.Build()

	if err != nil {
		return nil, err
	}

	return &logger{
		cfg:    cfg,
		logger: zapLogger,
	}, nil
}

func (l *logger) Debug(msg string, fields ...logman.Fields) {
	l.Log(logman.DebugLevel, msg, fields...)
}
func (l *logger) Info(msg string, fields ...logman.Fields) {
	l.Log(logman.InfoLevel, msg, fields...)
}
func (l *logger) Warning(msg string, fields ...logman.Fields) {
	l.Log(logman.WarningLevel, msg, fields...)
}
func (l *logger) Error(msg string, fields ...logman.Fields) {
	l.Log(logman.ErrorLevel, msg, fields...)
}
func (l *logger) Critical(msg string, fields ...logman.Fields) {
	l.Log(logman.CriticalLevel, msg, fields...)
}
func (l *logger) Log(level logman.Level, msg string, fields ...logman.Fields) {
	if l.Level() < level {
		return
	}

	switch level {
	case logman.DebugLevel:
		l.logger.Debug(msg, toZapFields(fields)...)
	case logman.InfoLevel:
		l.logger.Info(msg, toZapFields(fields)...)
	case logman.WarningLevel:
		l.logger.Warn(msg, toZapFields(fields)...)
	case logman.ErrorLevel:
		l.logger.Error(msg, toZapFields(fields)...)
	case logman.CriticalLevel:
		if e := l.logger.Check(zapcore.Level(zapCriticalLevel), msg); e != nil {
			e.Write(toZapFields(fields)...)
		}
	default:
		l.logger.Error(
			"Unknown log level",
			toZapFields([]logman.Fields{{
				"level":          level,
				"originalMsg":    msg,
				"originalFields": fields,
			}})...,
		)
	}
}
func (l *logger) Level() logman.Level {
	return l.cfg.Level
}

func toZapFields(fields []logman.Fields) []zap.Field {
	zapFields := []zap.Field{}
	for _, fieldSet := range fields {
		for field, val := range fieldSet {
			zapFields = append(zapFields, zap.Any(field, val))
		}
	}

	return zapFields
}

func toZapLevel(l logman.Level) zapcore.Level {
	switch l {
	case logman.DebugLevel:
		return zapcore.DebugLevel
	case logman.InfoLevel:
		return zapcore.InfoLevel
	case logman.WarningLevel:
		return zapcore.WarnLevel
	case logman.ErrorLevel:
		return zapcore.ErrorLevel
	case logman.CriticalLevel:
		return zapCriticalLevel
	}

	panic(fmt.Sprintf("Unknown log level: %d", l))
}

func zapCriticalAwareLowercaseLevelEncoder(
	l zapcore.Level,
	enc zapcore.PrimitiveArrayEncoder,
) {
	if l == zapCriticalLevel {
		enc.AppendString("critical")
	} else {
		enc.AppendString(l.String())
	}
}

func init() {
	logman.RegisterDriver(DriverName, driver{})
}
