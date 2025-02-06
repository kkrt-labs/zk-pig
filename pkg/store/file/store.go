package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkrt-labs/kakarot-controller/pkg/store"
)

type Store struct {
	cfg Config
}

func New(cfg Config) store.Store {
	return &Store{cfg: cfg}
}

func (f *Store) Store(_ context.Context, key string, reader io.Reader, headers *store.Headers) error {
	baseDir := f.baseDir(headers)
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

func (f *Store) Load(_ context.Context, key string, headers *store.Headers) (io.Reader, error) {
	baseDir := f.baseDir(headers)
	filePath := filepath.Join(baseDir, key)
	return os.Open(filePath)
}

func (f *Store) baseDir(headers *store.Headers) string {
	baseDir := f.cfg.DataDir
	if headers != nil && strings.Contains(baseDir, "default") { // baseDir includes "default" as a substring
		// replace default with chainID
		baseDir = strings.Replace(baseDir, "default", headers.KeyValue["chainID"], 1)
	}
	return baseDir
}
