package s3store

import aws "github.com/kkrt-labs/kakarot-controller/pkg/aws"

type Config struct {
	ProviderConfig *aws.ProviderConfig
	Bucket         string
	KeyPrefix      string
}
