package config

import (
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
		ViperKey:     "data-dir.root-dir",
		Name:         "data-dir",
		Env:          "DATA_DIR",
		Description:  "Path to data directory",
		DefaultValue: common.Ptr("data"),
	}
	preflightDirFlag = &spf13.StringFlag{
		ViperKey:     "data-dir.preflight-dir",
		Name:         "preflight-dir",
		Env:          "PREFLIGHT_DIR",
		Description:  "Path to preflight directory",
		DefaultValue: common.Ptr("preflight"),
	}
	inputsDirFlag = &spf13.StringFlag{
		ViperKey:     "data-dir.inputs-dir",
		Name:         "inputs-dir",
		Env:          "INPUTS_DIR",
		Description:  "Path to inputs directory",
		DefaultValue: common.Ptr("inputs"),
	}
	contentTypeFlag = &spf13.StringFlag{
		ViperKey:     "prover-inputs-store.content-type",
		Name:         "store-content-type",
		Env:          "PROVER_INPUTS_STORE_CONTENT_TYPE",
		Description:  "Prover inputs store content type",
		DefaultValue: common.Ptr("json"),
	}
	contentEncodingFlag = &spf13.StringFlag{
		ViperKey:     "prover-inputs-store.content-encoding",
		Name:         "store-content-encoding",
		Env:          "PROVER_INPUTS_STORE_CONTENT_ENCODING",
		Description:  "Prover inputs store content encoding",
		DefaultValue: common.Ptr(""),
	}
	awsS3BucketFlag = &spf13.StringFlag{
		ViperKey:    "prover-inputs-store.s3.aws-provider.bucket",
		Name:        "aws-s3-bucket",
		Env:         "AWS_S3_AWS_PROVIDER_BUCKET",
		Description: "AWS S3 AWS provider bucket name",
	}
	awsS3KeyPrefixFlag = &spf13.StringFlag{
		ViperKey:    "prover-inputs-store.s3.aws-provider.key-prefix",
		Name:        "aws-s3-key-prefix",
		Env:         "AWS_S3_AWS_PROVIDER_KEY_PREFIX",
		Description: "AWS S3 AWS provider key prefix",
	}
	awsS3AccessKeyFlag = &spf13.StringFlag{
		ViperKey:    "prover-inputs-store.s3.aws-provider.credentials.access-key",
		Name:        "aws-s3-access-key",
		Env:         "AWS_S3_AWS_PROVIDER_CREDENTIALS_ACCESS_KEY",
		Description: "AWS S3 AWS provider credentials access key",
	}
	awsS3SecretKeyFlag = &spf13.StringFlag{
		ViperKey:    "prover-inputs-store.s3.aws-provider.credentials.secret-key",
		Name:        "aws-s3-secret-key",
		Env:         "AWS_S3_AWS_PROVIDER_CREDENTIALS_SECRET_KEY",
		Description: "AWS S3 AWS provider credentials secret key",
	}
	awsS3RegionFlag = &spf13.StringFlag{
		ViperKey:    "prover-inputs-store.s3.aws-provider.region",
		Name:        "aws-s3-region",
		Env:         "AWS_S3_AWS_PROVIDER_REGION",
		Description: "AWS S3 AWS provider region",
	}
)

func AddChainFlags(v *viper.Viper, f *pflag.FlagSet) {
	chainIDFlag.Add(v, f)
	chainRPCURLFlag.Add(v, f)
}

func AddAWSFlags(v *viper.Viper, f *pflag.FlagSet) {
	awsS3BucketFlag.Add(v, f)
	awsS3KeyPrefixFlag.Add(v, f)
	awsS3AccessKeyFlag.Add(v, f)
	awsS3SecretKeyFlag.Add(v, f)
	awsS3RegionFlag.Add(v, f)
}

func AddStoreFlags(v *viper.Viper, f *pflag.FlagSet) {
	dataDirFlag.Add(v, f)
	preflightDirFlag.Add(v, f)
	inputsDirFlag.Add(v, f)
	contentTypeFlag.Add(v, f)
	contentEncodingFlag.Add(v, f)
}

func AddProverInputFlags(v *viper.Viper, f *pflag.FlagSet) {
	AddChainFlags(v, f)
	AddAWSFlags(v, f)
	AddStoreFlags(v, f)
}
