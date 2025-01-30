package s3

import aws "github.com/kkrt-labs/kakarot-controller/pkg/aws"

type Config struct {
	ProviderConfig *aws.ProviderConfig
	Bucket         string
	KeyPrefix      string
}
