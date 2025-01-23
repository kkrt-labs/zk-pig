package compress

import (
	"github.com/kkrt-labs/kakarot-controller/pkg/store"
	"github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
)

type Config struct {
	ContentEncoding store.ContentEncoding
	MultiConfig     multi.Config
}
