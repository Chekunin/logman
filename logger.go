package logman

type Logger interface {
	// Debug logs a detailed debug information.
	Debug(msg string, fields ...Fields)
	// Info logs general informational events that require no action.
	Info(msg string, fields ...Fields)
	// Warning logs exceptional occurrences that are not errors
	// and should be taken care of.
	Warning(msg string, fields ...Fields)
	// Error logs runtime errors that do not require immediate action
	// but should typically be monitored and investigated.
	Error(msg string, fields ...Fields)
	// Critical logs critical events such as overall application failure
	// or unusability and usually forcing a shutdown of the application
	// to prevent data loss.
	Critical(msg string, fields ...Fields)
	// Log logs with an arbitrary level.
	Log(level Level, msg string, fields ...Fields)
	// level return current log level.
	Level() Level
}

type Fields map[string]interface{}

type Level uint8

const (
	NotSet Level = iota
	CriticalLevel
	ErrorLevel
	WarningLevel
	InfoLevel
	DebugLevel
)
