package config

import (
	"github.com/kkrt-labs/kakarot-controller/pkg/common"
	"github.com/kkrt-labs/kakarot-controller/pkg/spf13"
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
		DefaultValue: common.Ptr("data/inputs"),
	}
	storageFlag = &spf13.StringFlag{
		ViperKey:     "store.location",
		Name:         "store-location",
		Env:          "STORE_LOCATION",
		Description:  "Storage type (file or s3)",
		DefaultValue: common.Ptr("file"),
	}
	contentTypeFlag = &spf13.StringFlag{
		ViperKey:     "store.content-type",
		Name:         "store-content-type",
		Env:          "STORE_CONTENT_TYPE",
		Description:  "Store content type",
		DefaultValue: common.Ptr("json"),
	}
	contentEncodingFlag = &spf13.StringFlag{
		ViperKey:     "store.content-encoding",
		Name:         "store-content-encoding",
		Env:          "STORE_CONTENT_ENCODING",
		Description:  "Store content encoding",
		DefaultValue: common.Ptr(""),
	}
	awsS3BucketFlag = &spf13.StringFlag{
		ViperKey:    "aws.s3.bucket",
		Name:        "aws-s3-bucket",
		Env:         "AWS_S3_BUCKET",
		Description: "AWS S3 bucket name",
	}
	awsS3KeyPrefixFlag = &spf13.StringFlag{
		ViperKey:    "aws.s3.key-prefix",
		Name:        "aws-s3-key-prefix",
		Env:         "AWS_S3_KEY_PREFIX",
		Description: "AWS S3 key prefix",
	}
	awsS3AccessKeyFlag = &spf13.StringFlag{
		ViperKey:    "aws.s3.access-key",
		Name:        "aws-s3-access-key",
		Env:         "AWS_S3_ACCESS_KEY",
		Description: "AWS S3 access key",
	}
	awsS3SecretKeyFlag = &spf13.StringFlag{
		ViperKey:    "aws.s3.secret-key",
		Name:        "aws-s3-secret-key",
		Env:         "AWS_S3_SECRET_KEY",
		Description: "AWS S3 secret key",
	}
	awsS3RegionFlag = &spf13.StringFlag{
		ViperKey:    "aws.s3.region",
		Name:        "aws-s3-region",
		Env:         "AWS_S3_REGION",
		Description: "AWS S3 region",
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
	contentTypeFlag.Add(v, f)
	contentEncodingFlag.Add(v, f)
	storageFlag.Add(v, f)
}

func AddProverInputsFlags(v *viper.Viper, f *pflag.FlagSet) {
	AddChainFlags(v, f)
	AddAWSFlags(v, f)
	AddStoreFlags(v, f)
	dataDirFlag.Add(v, f)
}
