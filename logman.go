package logman

import (
	"errors"
	"fmt"
)

var (
	DefaultChannelIsNotSetErr    = errors.New("Default channel is not set")
	DriverIsNotSetErr            = errors.New("Driver is not set")
	InvalidConfigValueErr        = errors.New("Invalid config value")
	MultipleInitErr              = errors.New("Already initialized")
	NoChannelsConfiguredErr      = errors.New("No channels configured")
	NoConfigForDefaultChannelErr = errors.New("No config for default channel")
	UnknownDriverErr             = errors.New("Unknown driver")
)

var logger = newDefault()

// Current returns the current logger used by the package-level functions.
func Current() Logger {
	return logger
}
func Debug(msg string, fields ...Fields) {
	logger.Debug(msg, fields...)
}
func Info(msg string, fields ...Fields) {
	logger.Info(msg, fields...)
}
func Warning(msg string, fields ...Fields) {
	logger.Warning(msg, fields...)
}
func Error(msg string, fields ...Fields) {
	logger.Error(msg, fields...)
}
func Critical(msg string, fields ...Fields) {
	logger.Critical(msg, fields...)
}
func Log(level Level, msg string, fields ...Fields) {
	logger.Log(level, msg, fields...)
}

type Logman struct {
	cfg      Config
	channels map[string]Logger
	isInited bool
}

func Init(cfg Config) error {
	if logger != nil && logger.isInited {
		return fmt.Errorf("Logman init failed <= %w", MultipleInitErr)
	}

	l, err := New(cfg)
	if err != nil {
		return fmt.Errorf("Logman init failed <= %w", err)
	}

	logger = l
	logger.isInited = true

	return nil
}
func InitOrPanic(cfg Config) {
	if err := Init(cfg); err != nil {
		panic(err)
	}
}
func New(cfg Config) (*Logman, error) {
	lm := &Logman{
		cfg:      cfg,
		channels: map[string]Logger{},
	}

	if err := lm.cfg.setDefaults().validate(); err != nil {
		return nil, fmt.Errorf("Invalid config <= %w", err)
	}

	err := createChannels(lm, cfg.Channels)
	if err != nil {
		return nil, fmt.Errorf("Failed to create channels <= %w", err)
	}

	return lm, nil
}
func NewOrPanic(cfg Config) *Logman {
	lm, err := New(cfg)
	if err != nil {
		panic(fmt.Errorf("Failed to create logger <= %w", err))
	}

	return lm
}
func newDefault() *Logman {
	RegisterDriver(DriverName, stdDriver{})

	return NewOrPanic(Config{
		DefaultChannel: "std",
		Level:          InfoLevel,
		Channels: ChannelConfigs{
			"std": stdLoggerConfig{},
		},
	})
}

func createChannels(lm *Logman, chCfgs map[string]ChannelConfig) error {
	for name, cfg := range chCfgs {
		logger, err := drivers[cfg.DriverName()].CreateLogger(lm, cfg)
		if err != nil {
			return fmt.Errorf("Failed to create logger: %s <= %w", name, err)
		}

		lm.channels[name] = logger
	}

	return nil
}

func (lm *Logman) Config() Config {
	return lm.cfg
}
func (lm *Logman) Channels(name ...string) map[string]Logger {
	chans := map[string]Logger{}
	for _, n := range name {
		if ch, exists := lm.channels[n]; exists {
			chans[n] = ch
		}
	}

	return chans
}

func (lm *Logman) Level() Level {
	return lm.cfg.Level
}
func (lm *Logman) Debug(msg string, fields ...Fields) {
	lm.Log(DebugLevel, msg, fields...)
}
func (lm *Logman) Info(msg string, fields ...Fields) {
	lm.Log(InfoLevel, msg, fields...)
}
func (lm *Logman) Warning(msg string, fields ...Fields) {
	lm.Log(WarningLevel, msg, fields...)
}
func (lm *Logman) Error(msg string, fields ...Fields) {
	lm.Log(ErrorLevel, msg, fields...)
}
func (lm *Logman) Critical(msg string, fields ...Fields) {
	lm.Log(CriticalLevel, msg, fields...)
}
func (lm *Logman) Log(level Level, msg string, fields ...Fields) {
	if lm.Level() < level {
		return
	}

	switch level {
	case DebugLevel:
		lm.channels[lm.cfg.DefaultChannel].Debug(msg, fields...)
	case InfoLevel:
		lm.channels[lm.cfg.DefaultChannel].Info(msg, fields...)
	case WarningLevel:
		lm.channels[lm.cfg.DefaultChannel].Warning(msg, fields...)
	case ErrorLevel:
		lm.channels[lm.cfg.DefaultChannel].Error(msg, fields...)
	case CriticalLevel:
		lm.channels[lm.cfg.DefaultChannel].Critical(msg, fields...)
	default:
		lm.channels[lm.cfg.DefaultChannel].Error(
			"Unknown log level",
			Fields{
				"level":          level,
				"originalMsg":    msg,
				"originalFields": fields,
			},
		)
	}
}
