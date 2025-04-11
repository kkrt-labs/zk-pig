package src

import (
	"fmt"

	"github.com/kkrt-labs/go-utils/common"
	"github.com/kkrt-labs/go-utils/spf13"
	"github.com/kkrt-labs/zk-pig/src/steps"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	configFileFlag = &spf13.StringArrayFlag{
		ViperKey:    "config",
		Name:        "config",
		Shorthand:   "c",
		Env:         "CONFIG",
		Description: "Configuration file (yaml format)",
	}
)

func AddConfigFileFlag(v *viper.Viper, f *pflag.FlagSet) {
	configFileFlag.Add(v, f)
}

var (
	generatorInclusionsFlag = &spf13.StringArrayFlag{
		ViperKey:    "generator.include",
		Name:        "include",
		Env:         "INCLUDE",
		Description: fmt.Sprintf("Data to include in the generated Prover Input (valid options: %q)", steps.ValidIncludes),
	}
	generatorFilterModuloFlag = &spf13.IntFlag{
		ViperKey:     "generator.filter.modulo.value",
		Name:         "filter-modulo",
		Env:          "FILTER_MODULO",
		Description:  "Does not generate prover input for blocks which number is not divisible by the given modulo",
		DefaultValue: common.Ptr(5),
	}
)

func AddGeneratorFlags(v *viper.Viper, f *pflag.FlagSet) {
	generatorInclusionsFlag.Add(v, f)
	generatorFilterModuloFlag.Add(v, f)
}

var (
	chainIDFlag = &spf13.StringFlag{
		ViperKey:    "chain.id",
		Name:        "chain-id",
		Env:         "CHAIN_ID",
		Description: "Chain ID (decimal)",
	}
	chainRPCURLFlag = &spf13.StringFlag{
		ViperKey:    "chain.rpc.url",
		Name:        "chain-rpc-url",
		Env:         "CHAIN_RPC_URL",
		Description: "Chain JSON-RPC URL",
	}
	dataDirFlag = &spf13.StringFlag{
		ViperKey:     "store.file.dir",
		Name:         "store-file-dir",
		Env:          "STORE_FILE_DIR",
		Description:  "Path to local data directory",
		DefaultValue: common.Ptr("data"),
	}
	awsS3BucketFlag = &spf13.StringFlag{
		ViperKey:    "store.s3.bucket",
		Name:        "store-s3-bucket",
		Env:         "STORE_S3_BUCKET",
		Description: "Optional AWS S3 bucket to store prover inputs",
	}
	awsS3BucketKeyPrefixFlag = &spf13.StringFlag{
		ViperKey:    "store.s3.prefix",
		Name:        "store-s3-prefix",
		Env:         "STORE_S3_PREFIX",
		Description: "Optional AWS S3 bucket key prefix where to store prover inputs",
	}
	awsS3AccessKeyFlag = &spf13.StringFlag{
		ViperKey:    "store.s3.aws-provider.credentials.access-key",
		Name:        "store-s3-access-key",
		Env:         "STORE_S3_ACCESS_KEY",
		Description: "Optional AWS Access Key to write prover inputs into S3 bucket",
	}
	awsS3SecretKeyFlag = &spf13.StringFlag{
		ViperKey:    "store.s3.aws-provider.credentials.secret-key",
		Name:        "store-s3-secret-key",
		Env:         "STORE_S3_SECRET_KEY",
		Description: "Optional AWS Secret Key to write prover inputs into S3 bucket",
	}
	awsS3RegionFlag = &spf13.StringFlag{
		ViperKey:    "store.s3.aws-provider.region",
		Name:        "store-s3-region",
		Env:         "STORE_S3_REGION",
		Description: "Optional AWS S3 bucket's region",
	}
	contentEncodingFlag = &spf13.StringFlag{
		ViperKey:     "store.content-encoding",
		Name:         "store-content-encoding",
		Env:          "STORE_CONTENT_ENCODING",
		Description:  fmt.Sprintf("Optional content encoding to apply to prover inputs before storing (one of %q)", []string{"gzip", "flate", "plain"}),
		DefaultValue: common.Ptr("plain"),
	}
	inputsContentTypeFlag = &spf13.StringFlag{
		ViperKey:     "inputs.content-type",
		Name:         "inputs-content-type",
		Env:          "INPUTS_CONTENT_TYPE",
		Description:  fmt.Sprintf("Content type for storing prover inputs (one of %q)", []string{"application/json", "application/protobuf"}),
		DefaultValue: common.Ptr("application/json"),
	}
	preflightDataEnabledFlag = &spf13.BoolFlag{
		ViperKey:     "preflight.enabled",
		Name:         "preflight-data-enabled",
		Env:          "PREFLIGHT_DATA_ENABLED",
		Description:  "Enable preflight data",
		DefaultValue: common.Ptr(true),
	}
)

func AddChainFlags(v *viper.Viper, f *pflag.FlagSet) {
	chainIDFlag.Add(v, f)
	chainRPCURLFlag.Add(v, f)
}

func AddAWSFlags(v *viper.Viper, f *pflag.FlagSet) {
	awsS3BucketFlag.Add(v, f)
	awsS3RegionFlag.Add(v, f)
	awsS3AccessKeyFlag.Add(v, f)
	awsS3SecretKeyFlag.Add(v, f)
	awsS3BucketKeyPrefixFlag.Add(v, f)
}

func AddStoreFlags(v *viper.Viper, f *pflag.FlagSet) {
	dataDirFlag.Add(v, f)
	AddAWSFlags(v, f)
	contentEncodingFlag.Add(v, f)
	inputsContentTypeFlag.Add(v, f)
	preflightDataEnabledFlag.Add(v, f)
}

func AddFlags(v *viper.Viper, f *pflag.FlagSet) {
	AddConfigFileFlag(v, f)
	AddChainFlags(v, f)
	AddStoreFlags(v, f)
	AddGeneratorFlags(v, f)
}
