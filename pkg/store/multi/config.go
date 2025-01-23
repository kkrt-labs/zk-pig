package multi

import (
	"github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
)

type Config struct {
	FileConfig *file.Config
	S3Config   *s3.Config
}
