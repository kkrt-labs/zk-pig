package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
)

type fileStore struct {
	cfg Config
}

func New(cfg Config) store.Store {
	return &fileStore{cfg: cfg}
}

func (f *fileStore) Store(_ context.Context, key string, reader io.Reader, _ *store.Headers) error {
	path := filepath.Join(f.cfg.DataDir, key)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (f *fileStore) Load(_ context.Context, key string, _ *store.Headers) (io.Reader, error) {
	path := filepath.Join(f.cfg.DataDir, key)
	return os.Open(path)
}
