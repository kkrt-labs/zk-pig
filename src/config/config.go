package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Chain struct {
		ID  string `mapstructure:"id,omitempty"`
		RPC struct {
			URL string `mapstructure:"url"`
		} `mapstructure:"rpc,omitempty"`
	} `mapstructure:"chain"`
	Log struct {
		Format string `mapstructure:"format"`
		Level  string `mapstructure:"level"`
	} `mapstructure:"log"`
	DataDir string                 `mapstructure:"data-dir"`
	Config  []string               `mapstructure:"config"`
	Extra   map[string]interface{} `mapstructure:"_extra,remain,omitempty"`
}

func (config *Config) Load(v *viper.Viper) error {
	for _, configPath := range v.GetStringSlice("config") {
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")

		if err := v.MergeInConfig(); err != nil {
			// Don't return error to keep compatibility with previous env
			// return config, fmt.Errorf("unable to read config file: %w", err)
			return err
		}
	}

	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	return nil
}
