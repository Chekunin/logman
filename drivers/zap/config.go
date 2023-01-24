package zap

import (
	"errors"
	"fmt"

	"github.com/Chekunin/logman"
)

type LoggerConfig struct {
	EnableStackTrace bool
	EnableCaller     bool
	Encoding         string
	Level            logman.Level
	Output           []string
}

func (c LoggerConfig) DriverName() string {
	return DriverName
}
func (c *LoggerConfig) setDefaults(lm *logman.Logman) *LoggerConfig {
	if c.Encoding == "" {
		c.Encoding = "json"
	}

	if c.Level == logman.NotSet {
		c.Level = lm.Level()
	}

	if len(c.Output) == 0 {
		c.Output = []string{"stderr"}
	}

	return c
}
func (c LoggerConfig) validate(_ *logman.Logman) error {
	if c.Level < logman.CriticalLevel || c.Level > logman.DebugLevel {
		return fmt.Errorf("Invalid log level: %d", c.Level)
	}

	if c.Encoding != "console" && c.Encoding != "json" {
		return fmt.Errorf("Invalid encoding: %s", c.Encoding)
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
		if option == "enableCaller" {
			enableCaller, ok := val.(bool)
			if !ok {
				return cfg, fmt.Errorf(
					"Failed to parse \"enableCaller\" option",
				)
			}
			cfg.EnableCaller = enableCaller
			continue
		}

		if option == "encoding" {
			encoding, ok := val.(string)
			if !ok {
				return cfg, fmt.Errorf("Failed to parse \"encoding\" option")
			}
			cfg.Encoding = encoding
			continue
		}

		if option == "output" {
			rawOutputs, ok := val.([]interface{})
			if !ok {
				return cfg, fmt.Errorf(
					"Invalid structure for \"output\" option",
				)
			}

			for n, rawOutput := range rawOutputs {
				output, ok := rawOutput.(string)
				if !ok {
					return cfg, fmt.Errorf(
						"Failed to parse \"output\" option (item #%d)", n+1,
					)
				}
				cfg.Output = append(cfg.Output, output)
			}
			continue
		}

		return cfg, fmt.Errorf("Unknown option passsed: %s", option)
	}

	return cfg, nil
}
