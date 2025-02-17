package config

import (
	"fmt"

	"github.com/kkrt-labs/go-utils/common"
	"github.com/kkrt-labs/go-utils/spf13"
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
		ViperKey:     "data-dir",
		Name:         "data-dir",
		Env:          "DATA_DIR",
		Description:  "Path to data directory",
		DefaultValue: common.Ptr("data"),
	}
	preflightDirFlag = &spf13.StringFlag{
		ViperKey:     "preflight-data-store.file.dir",
		Name:         "preflight-dir",
		Env:          "PREFLIGHT_DIR",
		Description:  "Directory where to store preflight data within --data-dir. If set to \"\" then does not store preflight data",
		DefaultValue: common.Ptr("preflight"),
	}
	inputsDirFlag = &spf13.StringFlag{
		ViperKey:     "prover-input-store.file.dir",
		Name:         "inputs-dir",
		Env:          "INPUTS_DIR",
		Description:  "Directory where to store prover inputs within --data-dir. If set to \"\" then does not store file to dir",
		DefaultValue: common.Ptr("inputs"),
	}
	contentTypeFlag = &spf13.StringFlag{
		ViperKey:     "prover-input-store.content-type",
		Name:         "inputs-content-type",
		Env:          "INPUTS_CONTENT_TYPE",
		Description:  fmt.Sprintf("Content type for storing prover inputs (one of %q)", []string{"json", "protobuf"}),
		DefaultValue: common.Ptr("json"),
	}
	contentEncodingFlag = &spf13.StringFlag{
		ViperKey:     "prover-input-store.content-encoding",
		Name:         "inputs-content-encoding",
		Env:          "INPUTS_CONTENT_ENCODING",
		Description:  fmt.Sprintf("Optional content encoding to apply to prover inputs before storing (one of %q)", []string{"gzip", "flate"}),
		DefaultValue: common.Ptr(""),
	}
)

func AddChainFlags(v *viper.Viper, f *pflag.FlagSet) {
	chainIDFlag.Add(v, f)
	chainRPCURLFlag.Add(v, f)
}

var (
	awsS3BucketFlag = &spf13.StringFlag{
		ViperKey:    "prover-input-store.s3.bucket",
		Name:        "inputs-aws-s3-bucket",
		Env:         "INPUTS_AWS_S3_BUCKET",
		Description: "Optional AWS S3 bucket to store prover inputs",
	}
	awsS3BucketKeyPrefixFlag = &spf13.StringFlag{
		ViperKey:    "prover-input-store.s3.bucket-key-prefix",
		Name:        "inputs-aws-s3-bucket-key-prefix",
		Env:         "INPUTS_AWS_S3_BUCKET_KEY_PREFIX",
		Description: "Optional AWS S3 bucket key prefix where to store prover inputs",
	}
	awsS3AccessKeyFlag = &spf13.StringFlag{
		ViperKey:    "prover-input-store.s3.aws-provider.credentials.access-key",
		Name:        "inputs-aws-s3-access-key",
		Env:         "INPUTS_AWS_S3_ACCESS_KEY",
		Description: "Optional AWS Access Key to write prover inputs into S3 bucket",
	}
	awsS3SecretKeyFlag = &spf13.StringFlag{
		ViperKey:    "prover-input-store.s3.aws-provider.credentials.secret-key",
		Name:        "inputs-aws-s3-secret-key",
		Env:         "INPUTS_AWS_S3_SECRET_KEY",
		Description: "Optional AWS Secret Key to write prover inputs into S3 bucket",
	}
	awsS3RegionFlag = &spf13.StringFlag{
		ViperKey:    "prover-input-store.s3.aws-provider.region",
		Name:        "inputs-aws-s3-region",
		Env:         "INPUTS_AWS_S3_REGION",
		Description: "Optional AWS S3 bucket's region",
	}
)

func AddAWSFlags(v *viper.Viper, f *pflag.FlagSet) {
	awsS3BucketFlag.Add(v, f)
	awsS3RegionFlag.Add(v, f)
	awsS3AccessKeyFlag.Add(v, f)
	awsS3SecretKeyFlag.Add(v, f)
	awsS3BucketKeyPrefixFlag.Add(v, f)
}

func AddStoreFlags(v *viper.Viper, f *pflag.FlagSet) {
	dataDirFlag.Add(v, f)
	preflightDirFlag.Add(v, f)
	inputsDirFlag.Add(v, f)
	contentTypeFlag.Add(v, f)
	contentEncodingFlag.Add(v, f)
}
