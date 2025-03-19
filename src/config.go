package src

import (
	"fmt"

	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/zk-pig/pkg/app"
	"github.com/spf13/viper"
)

type Config struct {
	Chain struct {
		ID  string `mapstructure:"id,omitempty"`
		RPC struct {
			URL string `mapstructure:"url"`
		} `mapstructure:"rpc,omitempty"`
	} `mapstructure:"chain"`
	DataDir   string   `mapstructure:"data-dir"`
	Config    []string `mapstructure:"config"`
	Generator struct {
		Include []string `mapstructure:"include"`
	} `mapstructure:"generator"`
	PreflightDataStore struct {
		File struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"file"`
	} `mapstructure:"preflight-data-store"`
	ProverInputStore struct {
		ContentType     string `mapstructure:"content-type"`
		ContentEncoding string `mapstructure:"content-encoding"`
		File            struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"file"`
		S3 struct {
			AWSProvider struct {
				Region      string `mapstructure:"region"`
				Credentials struct {
					AccessKey string `mapstructure:"access-key"`
					SecretKey string `mapstructure:"secret-key"`
				} `mapstructure:"credentials"`
			} `mapstructure:"aws-provider"`
			Bucket          string `mapstructure:"bucket"`
			BucketKeyPrefix string `mapstructure:"bucket-key-prefix"`
		} `mapstructure:"s3,omitempty"`
	} `mapstructure:"prover-input-store"`
	Extra map[string]interface{} `mapstructure:"_extra,remain,omitempty"`
	App   app.Config             `mapstructure:"app"`
	Log   log.Config             `mapstructure:"log"`
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
