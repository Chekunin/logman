package logman

import "fmt"

type ChannelConfig interface {
	DriverName() string
}
type ChannelConfigs map[string]ChannelConfig

type ChannelArbitraryConfig struct {
	Driver string
	Level  Level
	Extra  map[string]interface{}
}

func (c ChannelArbitraryConfig) DriverName() string {
	return c.Driver
}

type ChannelArbitraryConfigs map[string]ChannelArbitraryConfig

type Config struct {
	DefaultChannel string
	Level          Level
	Channels       ChannelConfigs
}

func NewConfig() Config {
	cfg := Config{
		DefaultChannel: "stack",
		Level:          DebugLevel,
	}
	return cfg
}

func NewLoggerChannels() ChannelArbitraryConfigs {
	c := make(ChannelArbitraryConfigs)
	c["stack"] = ChannelArbitraryConfig{
		Driver: "stack",
		Extra: map[string]interface{}{
			"channels": []interface{}{
				map[interface{}]interface{}{"name": "stderr"},
			},
		},
	}
	c["stderr"] = ChannelArbitraryConfig{Driver: "zap"}
	return c
}

func (cfg Config) WithChannels(chs ChannelArbitraryConfigs) Config {
	if cfg.Channels == nil {
		cfg.Channels = map[string]ChannelConfig{}
	}

	for chName, chCfg := range chs {
		cfg.Channels[chName] = chCfg
	}

	return cfg
}
func (c *Config) setDefaults() *Config {
	if c.Level == NotSet {
		c.Level = InfoLevel
	}

	return c
}
func (cfg Config) validate() error {
	if cfg.Level < CriticalLevel || cfg.Level > DebugLevel {
		return fmt.Errorf("Level \"%d\": %w", cfg.Level, InvalidConfigValueErr)
	}

	if len(cfg.Channels) == 0 {
		return NoChannelsConfiguredErr
	}

	if cfg.DefaultChannel == "" {
		return DefaultChannelIsNotSetErr
	}

	if _, exists := cfg.Channels[cfg.DefaultChannel]; !exists {
		return fmt.Errorf(
			"Channel \"%s\": %w",
			cfg.DefaultChannel,
			NoConfigForDefaultChannelErr,
		)
	}

	for chName, chCfg := range cfg.Channels {
		if chCfg.DriverName() == "" {
			return fmt.Errorf("Channel \"%s\": %w", chName, DriverIsNotSetErr)
		}

		_, exists := drivers[chCfg.DriverName()]
		if !exists {
			return fmt.Errorf(
				"Channel \"%s\", driver \"%s\": %w",
				chName,
				chCfg.DriverName(),
				UnknownDriverErr,
			)
		}
	}

	return nil
}
