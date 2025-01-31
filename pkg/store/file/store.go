package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
)

type Store struct {
	cfg Config
}

func New(cfg Config) store.Store {
	return &Store{cfg: cfg}
}

func (f *Store) Store(_ context.Context, key string, reader io.Reader, _ *store.Headers) error {
	baseDir := f.cfg.DataDir
	dir := filepath.Dir(filepath.Join(baseDir, key))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(baseDir, key)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (f *Store) Load(_ context.Context, key string, _ *store.Headers) (io.Reader, error) {
	filePath := filepath.Join(f.cfg.DataDir, key)
	return os.Open(filePath)
}
