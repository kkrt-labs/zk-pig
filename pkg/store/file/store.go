package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
)

type FileStore struct {
	cfg Config
}

func New(cfg Config) store.Store {
	return &FileStore{cfg: cfg}
}

func (f *FileStore) Store(_ context.Context, key string, reader io.Reader, _ *store.Headers) error {
	if err := os.MkdirAll(filepath.Dir(key), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(key)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (f *FileStore) Load(_ context.Context, key string, _ *store.Headers) (io.Reader, error) {
	return os.Open(key)
}
