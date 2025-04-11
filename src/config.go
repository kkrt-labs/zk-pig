package src

import (
	"fmt"

	"github.com/kkrt-labs/go-utils/app"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/spf13/viper"
)

type Config struct {
	Log    log.Config `mapstructure:"log"`
	App    app.Config `mapstructure:"app"`
	Config []string   `mapstructure:"config"`
	Chain  struct {
		ID  string `mapstructure:"id,omitempty"`
		RPC struct {
			URL string `mapstructure:"url"`
		} `mapstructure:"rpc,omitempty"`
	} `mapstructure:"chain"`

	Store struct {
		File struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"file,omitempty"`
		S3 struct {
			AWSProvider struct {
				Region      string `mapstructure:"region"`
				Credentials struct {
					AccessKey string `mapstructure:"access-key"`
					SecretKey string `mapstructure:"secret-key"`
				} `mapstructure:"credentials"`
			} `mapstructure:"aws-provider"`
			Bucket string `mapstructure:"bucket"`
			Prefix string `mapstructure:"prefix"`
		} `mapstructure:"s3,omitempty"`
		ContentEncoding string `mapstructure:"content-encoding"`
	} `mapstructure:"store"`
	ProverInputs struct {
		ContentType string `mapstructure:"content-type"`
	} `mapstructure:"inputs"`
	PreflightData struct {
		Enabled bool `mapstructure:"enabled"`
	} `mapstructure:"preflight"`
	Generator struct {
		Include []string `mapstructure:"include"`
		Filter  struct {
			Modulo struct {
				Value int `mapstructure:"value"`
			} `mapstructure:"modulo,omitempty"`
		} `mapstructure:"filter,omitempty"`
	} `mapstructure:"generator"`
	Extra map[string]interface{} `mapstructure:"_extra,remain,omitempty"`
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
