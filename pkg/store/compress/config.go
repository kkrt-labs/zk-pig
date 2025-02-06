package compress

import (
	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	multistore "github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
)

type Config struct {
	ContentEncoding  store.ContentEncoding
	MultiStoreConfig multistore.Config
}
