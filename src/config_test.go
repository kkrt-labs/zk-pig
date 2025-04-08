package src

import (
	"strings"
	"testing"
	"time"

	"github.com/kkrt-labs/go-utils/app"
	"github.com/kkrt-labs/go-utils/common"
	"github.com/kkrt-labs/go-utils/config"
	"github.com/kkrt-labs/go-utils/log"
	kkrthttp "github.com/kkrt-labs/go-utils/net/http"
	store "github.com/kkrt-labs/go-utils/store"
	"github.com/kkrt-labs/zk-pig/src/steps"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViperConfig(t *testing.T) {
	v := config.NewViper()
	v.Set("app.main-ep.addr", "localhost:8881")
	v.Set("app.main-ep.http.read-timeout", "40s")
	v.Set("app.main-ep.http.read-header-timeout", "41s")
	v.Set("app.main-ep.http.write-timeout", "42s")
	v.Set("app.main-ep.http.idle-timeout", "43s")
	v.Set("app.main-ep.net.keep-alive", "44s")
	v.Set("app.main-ep.net.keep-alive-probe.enable", "true")
	v.Set("app.main-ep.net.keep-alive-probe.idle", "45s")
	v.Set("app.main-ep.net.keep-alive-probe.interval", "46s")
	v.Set("app.main-ep.net.keep-alive-probe.count", "47")
	v.Set("app.main-ep.http.max-header-bytes", "40000")
	v.Set("app.healthz-ep.addr", "localhost:8882")
	v.Set("app.healthz-ep.http.read-timeout", "50s")
	v.Set("app.healthz-ep.http.read-header-timeout", "51s")
	v.Set("app.healthz-ep.http.write-timeout", "52s")
	v.Set("app.healthz-ep.http.idle-timeout", "53s")
	v.Set("app.healthz-ep.net.keep-alive", "54s")
	v.Set("app.healthz-ep.net.keep-alive-probe.enable", "true")
	v.Set("app.healthz-ep.net.keep-alive-probe.idle", "55s")
	v.Set("app.healthz-ep.net.keep-alive-probe.interval", "56s")
	v.Set("app.healthz-ep.net.keep-alive-probe.count", "57")
	v.Set("app.healthz-ep.http.max-header-bytes", "50000")
	v.Set("app.log.level", "info")
	v.Set("app.start-timeout", "10s")
	v.Set("app.stop-timeout", "20s")
	v.Set("chain.id", "1")
	v.Set("chain.rpc.url", "https://test.com")
	v.Set("store.file.dir", "testdata")
	v.Set("store.s3.provider.region", "us-east-1")
	v.Set("store.s3.provider.credentials.access-key", "test-access-key")
	v.Set("store.s3.provider.credentials.secret-key", "test-secret-key")
	v.Set("store.s3.bucket", "test-bucket")
	v.Set("store.s3.prefix", "test-prefix")
	v.Set("store.content-encoding", "gzip")
	v.Set("inputs.content-type", "application/protobuf")
	v.Set("preflight.enabled", "true")
	v.Set("generator.filter-modulo", "15")
	v.Set("generator.include", "preState,accessList")

	cfg := new(Config)
	err := cfg.Unmarshal(v)
	require.NoError(t, err)

	expectedCfg := &Config{
		App: &app.Config{
			MainEntrypoint: &kkrthttp.EntrypointConfig{
				Addr: common.Ptr("localhost:8881"),
				HTTP: &kkrthttp.ServerConfig{
					ReadTimeout:       common.Ptr(40 * time.Second),
					ReadHeaderTimeout: common.Ptr(41 * time.Second),
					WriteTimeout:      common.Ptr(42 * time.Second),
					IdleTimeout:       common.Ptr(43 * time.Second),
					MaxHeaderBytes:    common.Ptr(40000),
				},
				Net: &kkrthttp.ListenConfig{
					KeepAlive: common.Ptr(44 * time.Second),
					KeepAliveProbe: &kkrthttp.KeepAliveProbeConfig{
						Enable:   common.Ptr(true),
						Idle:     common.Ptr(45 * time.Second),
						Interval: common.Ptr(46 * time.Second),
						Count:    common.Ptr(47),
					},
				},
			},
			HealthzEntrypoint: &kkrthttp.EntrypointConfig{
				Addr: common.Ptr("localhost:8882"),
				HTTP: &kkrthttp.ServerConfig{
					ReadTimeout:       common.Ptr(50 * time.Second),
					ReadHeaderTimeout: common.Ptr(51 * time.Second),
					WriteTimeout:      common.Ptr(52 * time.Second),
					IdleTimeout:       common.Ptr(53 * time.Second),
					MaxHeaderBytes:    common.Ptr(50000),
				},
				Net: &kkrthttp.ListenConfig{
					KeepAlive: common.Ptr(54 * time.Second),
					KeepAliveProbe: &kkrthttp.KeepAliveProbeConfig{
						Enable:   common.Ptr(true),
						Idle:     common.Ptr(55 * time.Second),
						Interval: common.Ptr(56 * time.Second),
						Count:    common.Ptr(57),
					},
				},
			},
			Log: &log.Config{
				Level: common.Ptr(log.InfoLevel),
			},
			StartTimeout: common.Ptr("10s"),
			StopTimeout:  common.Ptr("20s"),
		},
		Chain: &ChainConfig{
			ID: common.Ptr("1"),
			RPC: &ChainRPCConfig{
				URL: common.Ptr("https://test.com"),
			},
		},
		Store: &StoreConfig{
			File: &FileStoreConfig{
				Dir: common.Ptr("testdata"),
			},
			S3: &S3StoreConfig{
				Provider: &AWSProviderConfig{
					Region: common.Ptr("us-east-1"),
					Credentials: &CredentialsConfig{
						AccessKey: common.Ptr("test-access-key"),
						SecretKey: common.Ptr("test-secret-key"),
					},
				},
				Bucket: common.Ptr("test-bucket"),
				Prefix: common.Ptr("test-prefix"),
			},
			ContentEncoding: common.Ptr(store.ContentEncodingGzip),
		},
		ProverInputs: &ProverInputsConfig{
			ContentType: common.Ptr(store.ContentTypeProtobuf),
		},
		PreflightData: &PreflightDataConfig{
			Enabled: common.Ptr(true),
		},
		Generator: &GeneratorConfig{
			FilterModulo:      common.Ptr(uint64(15)),
			IncludeExtensions: common.Ptr(steps.IncludePreState | steps.IncludeAccessList),
		},
	}
	assert.Equal(t, expectedCfg, cfg)
}

func TestEnv(t *testing.T) {
	env, err := (&Config{
		App: &app.Config{
			MainEntrypoint: &kkrthttp.EntrypointConfig{
				Addr: common.Ptr("localhost:8881"),
				HTTP: &kkrthttp.ServerConfig{
					ReadTimeout:       common.Ptr(40 * time.Second),
					ReadHeaderTimeout: common.Ptr(41 * time.Second),
					WriteTimeout:      common.Ptr(42 * time.Second),
					IdleTimeout:       common.Ptr(43 * time.Second),
					MaxHeaderBytes:    common.Ptr(40000),
				},
				Net: &kkrthttp.ListenConfig{
					KeepAlive: common.Ptr(44 * time.Second),
					KeepAliveProbe: &kkrthttp.KeepAliveProbeConfig{
						Enable:   common.Ptr(true),
						Idle:     common.Ptr(45 * time.Second),
						Interval: common.Ptr(46 * time.Second),
						Count:    common.Ptr(47),
					},
				},
			},
			HealthzEntrypoint: &kkrthttp.EntrypointConfig{
				Addr: common.Ptr("localhost:8882"),
				HTTP: &kkrthttp.ServerConfig{
					ReadTimeout:       common.Ptr(50 * time.Second),
					ReadHeaderTimeout: common.Ptr(51 * time.Second),
					WriteTimeout:      common.Ptr(52 * time.Second),
					IdleTimeout:       common.Ptr(53 * time.Second),
					MaxHeaderBytes:    common.Ptr(50000),
				},
				Net: &kkrthttp.ListenConfig{
					KeepAlive: common.Ptr(54 * time.Second),
					KeepAliveProbe: &kkrthttp.KeepAliveProbeConfig{
						Enable:   common.Ptr(true),
						Idle:     common.Ptr(55 * time.Second),
						Interval: common.Ptr(56 * time.Second),
						Count:    common.Ptr(57),
					},
				},
			},
			Log: &log.Config{
				Level: common.Ptr(log.InfoLevel),
			},
			StartTimeout: common.Ptr("10s"),
			StopTimeout:  common.Ptr("20s"),
		},
		Chain: &ChainConfig{
			ID: common.Ptr("1"),
			RPC: &ChainRPCConfig{
				URL: common.Ptr("https://test.com"),
			},
		},
		Store: &StoreConfig{
			File: &FileStoreConfig{
				Dir: common.Ptr("testdata"),
			},
			S3: &S3StoreConfig{
				Provider: &AWSProviderConfig{
					Region: common.Ptr("us-east-1"),
					Credentials: &CredentialsConfig{
						AccessKey: common.Ptr("test-access-key"),
						SecretKey: common.Ptr("test-secret-key"),
					},
				},
				Bucket: common.Ptr("test-bucket"),
				Prefix: common.Ptr("test-prefix"),
			},
			ContentEncoding: common.Ptr(store.ContentEncodingGzip),
		},
		ProverInputs: &ProverInputsConfig{
			ContentType: common.Ptr(store.ContentTypeProtobuf),
		},
		PreflightData: &PreflightDataConfig{
			Enabled: common.Ptr(true),
		},
		Generator: &GeneratorConfig{
			FilterModulo:      common.Ptr(uint64(15)),
			IncludeExtensions: common.Ptr(steps.IncludePreState | steps.IncludeAccessList),
		},
	}).Env()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"MAIN_EP_ADDR":                             "localhost:8881",
		"MAIN_EP_HTTP_READ_TIMEOUT":                "40s",
		"MAIN_EP_HTTP_READ_HEADER_TIMEOUT":         "41s",
		"MAIN_EP_HTTP_WRITE_TIMEOUT":               "42s",
		"MAIN_EP_HTTP_IDLE_TIMEOUT":                "43s",
		"MAIN_EP_NET_KEEP_ALIVE":                   "44s",
		"MAIN_EP_NET_KEEP_ALIVE_PROBE_ENABLE":      "true",
		"MAIN_EP_NET_KEEP_ALIVE_PROBE_IDLE":        "45s",
		"MAIN_EP_NET_KEEP_ALIVE_PROBE_INTERVAL":    "46s",
		"MAIN_EP_NET_KEEP_ALIVE_PROBE_COUNT":       "47",
		"MAIN_EP_HTTP_MAX_HEADER_BYTES":            "40000",
		"HEALTHZ_EP_ADDR":                          "localhost:8882",
		"HEALTHZ_EP_HTTP_READ_TIMEOUT":             "50s",
		"HEALTHZ_EP_HTTP_READ_HEADER_TIMEOUT":      "51s",
		"HEALTHZ_EP_HTTP_WRITE_TIMEOUT":            "52s",
		"HEALTHZ_EP_HTTP_IDLE_TIMEOUT":             "53s",
		"HEALTHZ_EP_NET_KEEP_ALIVE":                "54s",
		"HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_ENABLE":   "true",
		"HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_IDLE":     "55s",
		"HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_INTERVAL": "56s",
		"HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_COUNT":    "57",
		"HEALTHZ_EP_HTTP_MAX_HEADER_BYTES":         "50000",
		"LOG_LEVEL":                                "info",
		"START_TIMEOUT":                            "10s",
		"STOP_TIMEOUT":                             "20s",
		"CHAIN_ID":                                 "1",
		"CHAIN_RPC_URL":                            "https://test.com",
		"STORE_FILE_DIR":                           "testdata",
		"STORE_AWS_S3_PROVIDER_REGION":             "us-east-1",
		"STORE_AWS_S3_PROVIDER_ACCESS_KEY":         "test-access-key",
		"STORE_AWS_S3_PROVIDER_SECRET_KEY":         "test-secret-key",
		"STORE_AWS_S3_BUCKET":                      "test-bucket",
		"STORE_AWS_S3_PREFIX":                      "test-prefix",
		"STORE_CONTENT_ENCODING":                   "gzip",
		"INPUTS_CONTENT_TYPE":                      "application/protobuf",
		"PREFLIGHT_ENABLED":                        "true",
		"GENERATOR_FILTER_MODULO":                  "15",
		"GENERATOR_INCLUDE_EXTENSIONS":             "accessList,preState",
	}, env)
}

func TestFlagsUsage(t *testing.T) {
	v := config.NewViper()
	set := pflag.NewFlagSet("test", pflag.ContinueOnError)
	set.SortFlags = true
	err := AddFlags(v, set)
	require.NoError(t, err)

	expectedUsage := `      --chain-id string                                   Chain ID (decimal) [env: CHAIN_ID]
      --chain-rpc-url string                              Chain JSON-RPC URL [env: CHAIN_RPC_URL]
  -c, --config strings                                     [env: CONFIG] (default [config.yaml,config.yml])
      --generator-filter-modulo uint                      Generate prover input for blocks which number is divisible by the given modulo [env: GENERATOR_FILTER_MODULO] (default 5)
      --generator-includeextensions string                Optionnal extended data to include in the generated Prover Input (e.g. accessList [env: GENERATOR_INCLUDE_EXTENSIONS] (default "none")
      --healthz-ep-addr string                            healthz entrypoint: TCP Address to listen on [env: HEALTHZ_EP_ADDR] (default ":8081")
      --healthz-ep-http-idle-timeout string               healthz entrypoint: Maximum duration to wait for the next request when keep-alives are enabled (zero uses the value of read timeout) [env: HEALTHZ_EP_HTTP_IDLE_TIMEOUT] (default "30s")
      --healthz-ep-http-max-header-bytes int              healthz entrypoint: Maximum number of bytes the server will read parsing the request header's keys and values [env: HEALTHZ_EP_HTTP_MAX_HEADER_BYTES] (default 1048576)
      --healthz-ep-http-read-header-timeout string        healthz entrypoint: Maximum duration for reading request headers (zero uses the value of read timeout) [env: HEALTHZ_EP_HTTP_READ_HEADER_TIMEOUT] (default "30s")
      --healthz-ep-http-read-timeout string               healthz entrypoint: Maximum duration for reading the entire request including the body (zero means no timeout) [env: HEALTHZ_EP_HTTP_READ_TIMEOUT] (default "30s")
      --healthz-ep-http-write-timeout string              healthz entrypoint: Maximum duration before timing out writes of the response (zero means no timeout) [env: HEALTHZ_EP_HTTP_WRITE_TIMEOUT] (default "30s")
      --healthz-ep-net-keep-alive string                  healthz entrypoint: Keep alive period for network connections accepted by this entrypoint [env: HEALTHZ_EP_NET_KEEP_ALIVE] (default "-1s")
      --healthz-ep-net-keep-alive-probe-count int         healthz entrypoint: Maximum number of keep-alive probes that can go unanswered before dropping a connection [env: HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_COUNT] (default 9)
      --healthz-ep-net-keep-alive-probe-enable            healthz entrypoint: Enable keep alive probes [env: HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_ENABLE]
      --healthz-ep-net-keep-alive-probe-idle string       healthz entrypoint: Time that the connection must be idle before the first keep-alive probe is sent [env: HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_IDLE] (default "15s")
      --healthz-ep-net-keep-alive-probe-interval string   healthz entrypoint: Time between keep-alive probes [env: HEALTHZ_EP_NET_KEEP_ALIVE_PROBE_INTERVAL] (default "15s")
      --inputs-content-type string                        Content type (e.g. json) [env: INPUTS_CONTENT_TYPE] (default "application/json")
      --log-enable-caller                                 Enable caller [env: LOG_ENABLE_CALLER]
      --log-enable-stacktrace                             Enable automatic stacktrace capturing [env: LOG_ENABLE_STACKTRACE]
      --log-encoding-caller-encoder string                Encoding: Primitive representation for the log caller (e.g. 'full' [env: LOG_ENCODING_CALLER_ENCODER] (default "short")
      --log-encoding-caller-key string                    Encoding: Key for the log caller (if empty [env: LOG_ENCODING_CALLER_KEY] (default "caller")
      --log-encoding-console-separator string             Encoding: Field separator used by the console encoder [env: LOG_ENCODING_CONSOLE_SEPARATOR] (default "\t")
      --log-encoding-duration-encoder string              Encoding: Primitive representation for the log duration (e.g. 'string' [env: LOG_ENCODING_DURATION_ENCODER] (default "s")
      --log-encoding-function-key string                  Encoding: Key for the log function (if empty [env: LOG_ENCODING_FUNCTION_KEY]
      --log-encoding-level-encoder string                 Encoding: Primitive representation for the log level (e.g. 'capital' [env: LOG_ENCODING_LEVEL_ENCODER] (default "capitalColor")
      --log-encoding-level-key string                     Encoding: Key for the log level (if empty [env: LOG_ENCODING_LEVEL_KEY] (default "level")
      --log-encoding-line-ending string                   Encoding: Line ending [env: LOG_ENCODING_LINE_ENDING] (default "\n")
      --log-encoding-message-key string                   Encoding: Key for the log message (if empty [env: LOG_ENCODING_MESSAGE_KEY] (default "msg")
      --log-encoding-name-encoder string                  Encoding: Primitive representation for the log logger name (e.g. 'full' [env: LOG_ENCODING_NAME_ENCODER] (default "full")
      --log-encoding-name-key string                      Encoding: Key for the log logger name (if empty [env: LOG_ENCODING_NAME_KEY] (default "logger")
      --log-encoding-skip-line-ending                     Encoding: Skip the line ending [env: LOG_ENCODING_SKIP_LINE_ENDING]
      --log-encoding-stacktrace-key string                Encoding: Key for the log stacktrace (if empty [env: LOG_ENCODING_STACKTRACE_KEY] (default "stacktrace")
      --log-encoding-time-encoder string                  Encoding: Primitive representation for the log timestamp (e.g. 'rfc3339nano' [env: LOG_ENCODING_TIME_ENCODER] (default "rfc3339")
      --log-encoding-time-key string                      Encoding: Key for the log timestamp (if empty [env: LOG_ENCODING_TIME_KEY] (default "ts")
      --log-err-output strings                            List of URLs to write internal logger errors to [env: LOG_ERROR_OUTPUT_PATHS] (default [stderr])
      --log-format string                                 Log format [env: LOG_FORMAT] (default "text")
      --log-level string                                  Minimum enabled logging level [env: LOG_LEVEL] (default "info")
      --log-output strings                                List of URLs or file paths to write logging output to [env: LOG_OUTPUT_PATHS] (default [stderr])
      --log-sampling-initial int                          Sampling: Number of log entries with the same level and message to log before dropping entries [env: LOG_SAMPLING_INITIAL] (default 100)
      --log-sampling-thereafter int                       Sampling: After the initial number of entries [env: LOG_SAMPLING_THEREAFTER] (default 100)
      --main-ep-addr string                               main entrypoint: TCP Address to listen on [env: MAIN_EP_ADDR] (default ":8080")
      --main-ep-http-idle-timeout string                  main entrypoint: Maximum duration to wait for the next request when keep-alives are enabled (zero uses the value of read timeout) [env: MAIN_EP_HTTP_IDLE_TIMEOUT] (default "30s")
      --main-ep-http-max-header-bytes int                 main entrypoint: Maximum number of bytes the server will read parsing the request header's keys and values [env: MAIN_EP_HTTP_MAX_HEADER_BYTES] (default 1048576)
      --main-ep-http-read-header-timeout string           main entrypoint: Maximum duration for reading request headers (zero uses the value of read timeout) [env: MAIN_EP_HTTP_READ_HEADER_TIMEOUT] (default "30s")
      --main-ep-http-read-timeout string                  main entrypoint: Maximum duration for reading the entire request including the body (zero means no timeout) [env: MAIN_EP_HTTP_READ_TIMEOUT] (default "30s")
      --main-ep-http-write-timeout string                 main entrypoint: Maximum duration before timing out writes of the response (zero means no timeout) [env: MAIN_EP_HTTP_WRITE_TIMEOUT] (default "30s")
      --main-ep-net-keep-alive string                     main entrypoint: Keep alive period for network connections accepted by this entrypoint [env: MAIN_EP_NET_KEEP_ALIVE] (default "-1s")
      --main-ep-net-keep-alive-probe-count int            main entrypoint: Maximum number of keep-alive probes that can go unanswered before dropping a connection [env: MAIN_EP_NET_KEEP_ALIVE_PROBE_COUNT] (default 9)
      --main-ep-net-keep-alive-probe-enable               main entrypoint: Enable keep alive probes [env: MAIN_EP_NET_KEEP_ALIVE_PROBE_ENABLE]
      --main-ep-net-keep-alive-probe-idle string          main entrypoint: Time that the connection must be idle before the first keep-alive probe is sent [env: MAIN_EP_NET_KEEP_ALIVE_PROBE_IDLE] (default "15s")
      --main-ep-net-keep-alive-probe-interval string      main entrypoint: Time between keep-alive probes [env: MAIN_EP_NET_KEEP_ALIVE_PROBE_INTERVAL] (default "15s")
      --preflight-enabled                                  [env: PREFLIGHT_ENABLED]
      --start-timeout string                              Start timeout [env: START_TIMEOUT] (default "10s")
      --stop-timeout string                               Stop timeout [env: STOP_TIMEOUT] (default "10s")
      --store-aws-s3-bucket string                        AWS S3 bucket [env: STORE_AWS_S3_BUCKET]
      --store-aws-s3-prefix string                        AWS S3 bucket key prefix [env: STORE_AWS_S3_PREFIX]
      --store-aws-s3-provider-access-key string           AWS access key [env: STORE_AWS_S3_PROVIDER_ACCESS_KEY]
      --store-aws-s3-provider-region string               AWS region [env: STORE_AWS_S3_PROVIDER_REGION]
      --store-aws-s3-provider-secret-key string           AWS secret key [env: STORE_AWS_S3_PROVIDER_SECRET_KEY]
      --store-content-encoding string                     Content encoding (e.g. gzip) [env: STORE_CONTENT_ENCODING] (default "plain")
      --store-file-dir string                             Path to local data directory [env: STORE_FILE_DIR] (default "data")
`

	expectedRaws := strings.Split(expectedUsage, "\n")
	actualRaws := strings.Split(set.FlagUsages(), "\n")

	require.Equal(t, len(expectedRaws), len(actualRaws))
	for i, raw := range expectedRaws {
		require.Equal(t, raw, actualRaws[i])
	}
}

func TestAddFlagsAndLoadEnv(t *testing.T) {
	cfg := &Config{
		App: &app.Config{
			MainEntrypoint: &kkrthttp.EntrypointConfig{
				Addr: common.Ptr("localhost:8885"),
				HTTP: &kkrthttp.ServerConfig{
					ReadTimeout:       common.Ptr(40 * time.Second),
					ReadHeaderTimeout: common.Ptr(41 * time.Second),
					WriteTimeout:      common.Ptr(42 * time.Second),
					IdleTimeout:       common.Ptr(43 * time.Second),
					MaxHeaderBytes:    common.Ptr(40000),
				},
				Net: &kkrthttp.ListenConfig{
					KeepAlive: common.Ptr(44 * time.Second),
					KeepAliveProbe: &kkrthttp.KeepAliveProbeConfig{
						Enable:   common.Ptr(true),
						Idle:     common.Ptr(45 * time.Second),
						Interval: common.Ptr(46 * time.Second),
						Count:    common.Ptr(47),
					},
				},
			},
			HealthzEntrypoint: &kkrthttp.EntrypointConfig{
				Addr: common.Ptr("localhost:8886"),
				HTTP: &kkrthttp.ServerConfig{
					ReadTimeout:       common.Ptr(50 * time.Second),
					ReadHeaderTimeout: common.Ptr(51 * time.Second),
					WriteTimeout:      common.Ptr(52 * time.Second),
					IdleTimeout:       common.Ptr(53 * time.Second),
					MaxHeaderBytes:    common.Ptr(50000),
				},
				Net: &kkrthttp.ListenConfig{
					KeepAlive: common.Ptr(54 * time.Second),
					KeepAliveProbe: &kkrthttp.KeepAliveProbeConfig{
						Enable:   common.Ptr(true),
						Idle:     common.Ptr(55 * time.Second),
						Interval: common.Ptr(56 * time.Second),
						Count:    common.Ptr(57),
					},
				},
			},
			Log:          log.DefaultConfig(),
			StartTimeout: common.Ptr("10s"),
			StopTimeout:  common.Ptr("20s"),
		},
		Config: &[]*string{common.Ptr("config.yaml")},
		Chain: &ChainConfig{
			RPC: &ChainRPCConfig{
				URL: common.Ptr("https://test.com"),
			},
		},
		Store: &StoreConfig{
			File: &FileStoreConfig{
				Dir: common.Ptr("testdata"),
			},
			S3: &S3StoreConfig{
				Provider: &AWSProviderConfig{
					Region: common.Ptr("us-east-1"),
					Credentials: &CredentialsConfig{
						AccessKey: common.Ptr("test-access-key"),
						SecretKey: common.Ptr("test-secret-key"),
					},
				},
				Bucket: common.Ptr("test-bucket"),
				Prefix: common.Ptr("test-prefix"),
			},
			ContentEncoding: common.Ptr(store.ContentEncodingGzip),
		},
		PreflightData: &PreflightDataConfig{
			Enabled: common.Ptr(true),
		},
		ProverInputs: &ProverInputsConfig{
			ContentType: common.Ptr(store.ContentTypeJSON),
		},
		Generator: &GeneratorConfig{
			FilterModulo:      common.Ptr(uint64(15)),
			IncludeExtensions: common.Ptr(steps.IncludePreState | steps.IncludeAccessList),
		},
	}

	v := config.NewViper()
	err := AddFlags(v, pflag.NewFlagSet("test", pflag.ContinueOnError))
	require.NoError(t, err)

	env, err := cfg.Env()
	require.NoError(t, err)
	for k, v := range env {
		t.Setenv(k, v)
	}

	loadedCfg := new(Config)
	err = loadedCfg.Unmarshal(v)
	require.NoError(t, err)
	assert.Equal(t, cfg, loadedCfg)
}

func TestUnmarshalFromDefaults(t *testing.T) {
	v := config.NewViper()
	err := AddFlags(v, pflag.NewFlagSet("test", pflag.ContinueOnError))
	require.NoError(t, err)

	cfg := new(Config)
	err = cfg.Load(v)
	require.NoError(t, err)
	assert.Equal(t, DefaultConfig(), cfg)
}
