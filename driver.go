package logman

var drivers = map[string]Driver{}

type Driver interface {
	CreateLogger(lm *Logman, loggerCfg ChannelConfig) (Logger, error)
}

func RegisterDriver(name string, driver Driver) {
	if name == "" {
		panic("logman: Empty driver name passed")
	}

	if driver == nil {
		panic("logman: Try to register nil driver")
	}

	if _, dup := drivers[name]; dup {
		panic("logman: Register called twice for driver " + name)
	}

	drivers[name] = driver
}
