package logman

import (
	"fmt"
	"log"
	"os"
)

const DriverName = "std"

type stdDriver struct{}

func (d stdDriver) CreateLogger(_ *Logman, c ChannelConfig) (Logger, error) {
	cfg, err := parseConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config: %w", err)
	}
	return newLogger(cfg), nil
}

var levelLabels = map[Level]string{
	CriticalLevel: "CRT",
	ErrorLevel:    "ERR",
	WarningLevel:  "WRN",
	InfoLevel:     "INF",
	DebugLevel:    "DBG",
}

type stdLoggerConfig struct{}

func (lc stdLoggerConfig) DriverName() string {
	return DriverName
}

type stdLogger struct {
	*log.Logger
	level Level
}

func newLogger(_ stdLoggerConfig) *stdLogger {
	return &stdLogger{
		log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC),
		DebugLevel,
	}
}

func (l *stdLogger) Debug(msg string, fields ...Fields) {
	l.Log(DebugLevel, msg, fields...)
}
func (l *stdLogger) Info(msg string, fields ...Fields) {
	l.Log(InfoLevel, msg, fields...)

}
func (l *stdLogger) Warning(msg string, fields ...Fields) {
	l.Log(WarningLevel, msg, fields...)

}
func (l *stdLogger) Error(msg string, fields ...Fields) {
	l.Log(ErrorLevel, msg, fields...)

}
func (l *stdLogger) Critical(msg string, fields ...Fields) {
	l.Log(CriticalLevel, msg, fields...)

}
func (l *stdLogger) Log(level Level, msg string, fields ...Fields) {
	if l.level < level {
		return
	}

	l.Printf("[%s] %s %+v\n", levelLabels[level], msg, fields)
}
func (l *stdLogger) Level() Level {
	return l.level
}

func parseConfig(c ChannelConfig) (stdLoggerConfig, error) {
	if cfg, ok := c.(stdLoggerConfig); ok {
		return cfg, nil
	}

	return stdLoggerConfig{}, nil
}
