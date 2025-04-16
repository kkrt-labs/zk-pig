package src

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kkrt-labs/go-utils/app"
	"github.com/kkrt-labs/go-utils/common"
	"github.com/kkrt-labs/go-utils/config"
	store "github.com/kkrt-labs/go-utils/store"
	"github.com/kkrt-labs/zk-pig/src/steps"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	config.RegisterGlobalDecodeHooks(
		func(f, t reflect.Type, data any) (any, error) {
			if f.Kind() != reflect.String {
				return data, nil
			}

			if t == reflect.TypeOf(store.ContentType(0)) {
				return store.ParseContentType(data.(string))
			}

			if t == reflect.TypeOf(store.ContentEncoding(0)) {
				return store.ParseContentEncoding(data.(string))
			}

			if t == reflect.TypeOf(steps.Include(0)) {
				return steps.ParseIncludes(strings.Split(data.(string), ",")...)
			}

			return data, nil
		},
	)
}

func DefaultConfig() *Config {
	return &Config{
		App:    app.DefaultConfig(),
		Config: common.PtrSlice("config.yaml", "config.yml"),
		Store: &StoreConfig{
			File: &FileStoreConfig{
				Enabled: common.Ptr(true),
				Dir:     common.Ptr("data"),
			},
			S3: &S3StoreConfig{
				Enabled: common.Ptr(false),
				Provider: &AWSProviderConfig{
					Credentials: &CredentialsConfig{},
				},
			},
			ContentEncoding: common.Ptr(store.ContentEncodingPlain),
		},
		Chain: &ChainConfig{
			RPC: &ChainRPCConfig{},
		},
		ProverInputs: &ProverInputsConfig{
			ContentType: common.Ptr(store.ContentTypeJSON),
		},
		Generator: &GeneratorConfig{
			StorePreflightData: common.Ptr(false),
			FilterModulo:       common.Ptr(uint64(5)),
		},
	}
}

type Config struct {
	App          *app.Config         `key:"app" env:"-" flag:"-"`
	Config       *[]*string          `key:"config" short:"c"`
	Chain        *ChainConfig        `key:"chain"`
	Store        *StoreConfig        `key:"store"`
	ProverInputs *ProverInputsConfig `key:"inputs" env:"INPUTS" flag:"inputs"`
	Generator    *GeneratorConfig    `key:"generator"`
}

func (cfg *Config) Load(v *viper.Viper) error {
	for _, configPath := range v.GetStringSlice("config") {
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")

		if err := v.MergeInConfig(); err != nil {
			// Don't return error to keep compatibility with previous env
			// return config, fmt.Errorf("unable to read config file: %w", err)
			return err
		}
	}

	if err := cfg.Unmarshal(v); err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	return nil
}

// Env returns the environment variables for the config.
func (cfg *Config) Env() (map[string]string, error) {
	return config.Env(cfg, nil)
}

// Unmarshal unmarshals the config from the given viper.Viper.
func (cfg *Config) Unmarshal(v *viper.Viper) error {
	return config.Unmarshal(cfg, v)
}

// AddFlags adds the flags for the config.
func AddFlags(v *viper.Viper, f *pflag.FlagSet) error {
	return config.AddFlags(DefaultConfig(), v, f, nil)
}

type ChainConfig struct {
	ID  *string         `key:"id,omitempty" desc:"Chain ID (decimal)"`
	RPC *ChainRPCConfig `key:"rpc,omitempty"`
}

type ChainRPCConfig struct {
	URL *string `key:"url" desc:"Chain JSON-RPC URL"`
}

type StoreConfig struct {
	File            *FileStoreConfig       `key:"file,omitempty"`
	S3              *S3StoreConfig         `key:"s3,omitempty" env:"AWS_S3" flag:"aws-s3"`
	ContentEncoding *store.ContentEncoding `key:"content-encoding" env:"CONTENT_ENCODING" flag:"content-encoding" desc:"Content encoding (e.g. gzip)"`
}

type FileStoreConfig struct {
	Enabled *bool   `key:"enabled" desc:"Enable file store"`
	Dir     *string `key:"dir" desc:"Path to local data directory"`
}

type S3StoreConfig struct {
	Enabled  *bool              `key:"enabled" desc:"Enable S3 store"`
	Provider *AWSProviderConfig `key:"provider" env:"PROVIDER" flag:"provider"`
	Bucket   *string            `key:"bucket" desc:"AWS S3 bucket"`
	Prefix   *string            `key:"prefix" desc:"AWS S3 bucket key prefix"`
}

type AWSProviderConfig struct {
	Region      *string            `key:"region" desc:"AWS region"`
	Credentials *CredentialsConfig `key:"credentials" env:"-" flag:"-"`
}

type CredentialsConfig struct {
	AccessKey *string `key:"access-key" env:"ACCESS_KEY" flag:"access-key" desc:"AWS access key"`
	SecretKey *string `key:"secret-key" env:"SECRET_KEY" flag:"secret-key" desc:"AWS secret key"`
}

type ProverInputsConfig struct {
	ContentType *store.ContentType `key:"content-type" env:"CONTENT_TYPE" flag:"content-type" desc:"Content type (e.g. json)"`
}

type GeneratorConfig struct {
	StorePreflightData *bool          `key:"store-preflight-data" env:"STORE_PREFLIGHT_DATA" flag:"store-preflight-data" desc:"Store intermediate preflight data when generating prover inputs"`
	IncludeExtensions  *steps.Include `key:"include" env:"INCLUDE_EXTENSIONS" desc:"Optionnal extended data to include in the generated Prover Input (e.g. accessList, preState, stateDiffs, committed, all)"`
	FilterModulo       *uint64        `key:"filter-modulo" env:"FILTER_MODULO" flag:"filter-modulo" desc:"Generate prover input for blocks which number is divisible by the given modulo"`
}
