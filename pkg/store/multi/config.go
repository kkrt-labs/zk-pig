package multistore

import (
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	s3store "github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
)

type Config struct {
	FileConfig *filestore.Config
	S3Config   *s3store.Config
}
