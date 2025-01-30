package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

type CredentialsConfig struct {
	AccessKey string
	SecretKey string
}

type ProviderConfig struct {
	Region      string
	Credentials *CredentialsConfig
}

func LoadConfig(cfg *ProviderConfig) (aws.Config, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.Credentials.AccessKey,
			cfg.Credentials.SecretKey,
			"",
		)),
	)
	if err != nil {
		return aws.Config{}, err
	}

	return awsCfg, nil
}
