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
	DataDir struct {
		Root      string `mapstructure:"root-dir"`
		Preflight string `mapstructure:"preflight-dir"`
		Inputs    string `mapstructure:"inputs-dir"`
	} `mapstructure:"data-dir"`
	Config           []string `mapstructure:"config"`
	ProverInputStore struct {
		ContentType     string `mapstructure:"content-type"`
		ContentEncoding string `mapstructure:"content-encoding"`
		S3              struct {
			AWSProvider struct {
				Region      string `mapstructure:"region"`
				Credentials struct {
					AccessKey string `mapstructure:"access-key"`
					SecretKey string `mapstructure:"secret-key"`
				}
				Bucket    string `mapstructure:"bucket"`
				KeyPrefix string `mapstructure:"key-prefix"`
			} `mapstructure:"aws-provider"`
		} `mapstructure:"s3,omitempty"`
	} `mapstructure:"prover-inputs-store"`
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
