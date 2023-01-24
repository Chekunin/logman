package stack

import (
	"errors"
	"fmt"

	"github.com/Chekunin/logman"
)

type ChannelConfig struct {
	Name          string
	DisableBubble bool
}

func (c *ChannelConfig) setDefaults() *ChannelConfig {
	return c
}
func (c ChannelConfig) validate() error {
	if c.Name == "" {
		return errors.New("No \"name\" defined")
	}

	return nil
}

type LoggerConfig struct {
	Level    logman.Level
	Channels []ChannelConfig
}

func (c LoggerConfig) DriverName() string {
	return DriverName
}
func (c *LoggerConfig) setDefaults(lm *logman.Logman) *LoggerConfig {
	if c.Level == logman.NotSet {
		c.Level = lm.Level()
	}

	for _, chCfg := range c.Channels {
		chCfg.setDefaults()
	}

	return c
}
func (c LoggerConfig) validate(lm *logman.Logman) error {
	if c.Level < logman.CriticalLevel || c.Level > logman.DebugLevel {
		return fmt.Errorf("Invalid log level: %d", c.Level)
	}

	if len(c.Channels) == 0 {
		return errors.New("No channels configured")
	}

	for _, chCfg := range c.Channels {
		if err := chCfg.validate(); err != nil {
			return fmt.Errorf(
				"Invalid config for channel \"%s\": %w", chCfg.Name, err,
			)
		}

		for lmChName, lmChCfg := range lm.Config().Channels {
			if chCfg.Name == lmChName && lmChCfg.DriverName() == DriverName {
				return fmt.Errorf(
					"Recursive usage of stack logger \"%s\"", chCfg.Name,
				)
			}
		}

		if _, exists := lm.Config().Channels[chCfg.Name]; !exists {
			return fmt.Errorf(
				"No configuration defined for channel \"%s\"", chCfg.Name,
			)
		}
	}

	return nil
}

func parseConfig(c logman.ChannelConfig) (LoggerConfig, error) {
	if cfg, ok := c.(LoggerConfig); ok {
		return cfg, nil
	}

	cfg := LoggerConfig{}

	rawCfg, ok := c.(logman.ChannelArbitraryConfig)
	if !ok {
		return cfg, errors.New("Invalid config structure")
	}

	cfg.Level = rawCfg.Level

	for option, val := range rawCfg.Extra {
		if option == "channels" {
			rawChCfgs, ok := val.([]interface{})
			if !ok {
				return cfg, fmt.Errorf(
					"Invalid structure for \"channels\" option",
				)
			}

			for n, rawChCfg := range rawChCfgs {
				chCfg := ChannelConfig{}

				rawChCfg, ok := rawChCfg.(map[interface{}]interface{})
				if !ok {
					return cfg, fmt.Errorf(
						"Failed to parse \"channels\" option #%d", n+1,
					)
				}

				for opt, v := range rawChCfg {
					opt, ok := opt.(string)
					if !ok {
						return cfg, fmt.Errorf(
							"Not string key used in \"channels\" option #%d",
							n+1,
						)
					}

					if opt == "disableBubble" {
						disableBubble, ok := v.(bool)
						if !ok {
							return cfg, fmt.Errorf(
								"Failed to parse \"disableBubble\" "+
									"in \"channels\" option #%d",
								n+1,
							)
						}
						chCfg.DisableBubble = disableBubble
						continue
					}

					if opt == "name" {
						name, ok := v.(string)
						if !ok {
							return cfg, fmt.Errorf(
								"Failed to parse \"name\" "+
									"in \"channels\" option #%d",
								n+1,
							)
						}
						chCfg.Name = name
						continue
					}

					return cfg, fmt.Errorf(
						"Unknown option passsed for \"channels\": %s", opt,
					)
				}

				cfg.Channels = append(cfg.Channels, chCfg)
			}
			continue
		}

		return cfg, fmt.Errorf("Unknown option passsed: %s", option)
	}

	return cfg, nil
}
