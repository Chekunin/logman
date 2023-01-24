package stack

import (
	"fmt"

	"github.com/Chekunin/logman"
)

const DriverName = "stack"

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

type channel struct {
	cfg    ChannelConfig
	logger logman.Logger
}

type logger struct {
	cfg      LoggerConfig
	channels []channel
	logman   *logman.Logman
}

func newLogger(cfg LoggerConfig, lm *logman.Logman) (*logger, error) {
	if err := cfg.setDefaults(lm).validate(lm); err != nil {
		return nil, fmt.Errorf("Invalid config: %w", err)
	}

	return &logger{
		cfg:    cfg,
		logman: lm,
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

	// deferred init (@TODO: add smth like `onCreated` hook to logman?)
	if len(l.channels) == 0 {
		for _, c := range l.cfg.Channels {
			l.channels = append(l.channels, channel{
				logger: l.logman.Channels(c.Name)[c.Name],
				cfg:    c,
			})
		}
	}

	for _, c := range l.channels {
		if c.logger.Level() < level {
			continue
		}

		c.logger.Log(level, msg, fields...) // @TODO: use goroutines?

		if c.cfg.DisableBubble {
			break
		}
	}
}
func (l *logger) Level() logman.Level {
	return l.cfg.Level
}

func init() {
	logman.RegisterDriver(DriverName, driver{})
}
