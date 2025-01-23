package s3

type Config struct {
	Bucket    string
	Region    string
	Endpoint  string
	KeyPrefix string
	AccessKey string
	SecretKey string
}
