package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	if baseDir == "" {
		return nil, errors.New("base directory is required")
	}
	clean := filepath.Clean(baseDir)
	if err := os.MkdirAll(clean, 0o755); err != nil {
		return nil, err
	}
	return &LocalStorage{baseDir: clean}, nil
}

func (l *LocalStorage) Put(ctx context.Context, key string, body io.Reader, contentType string) error {
	_ = ctx
	_ = contentType

	fullPath, err := l.pathForKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(fullPath), "imagekit-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	if _, err := io.Copy(tmp, body); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), fullPath)
}

func (l *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	_ = ctx
	fullPath, err := l.pathForKey(key)
	if err != nil {
		return nil, err
	}
	return os.Open(fullPath)
}

func (l *LocalStorage) pathForKey(key string) (string, error) {
	if key == "" {
		return "", errors.New("object key is required")
	}

	clean := filepath.Clean(string(os.PathSeparator) + key)
	clean = strings.TrimPrefix(clean, string(os.PathSeparator))
	fullPath := filepath.Join(l.baseDir, clean)

	rel, err := filepath.Rel(l.baseDir, fullPath)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", errors.New("invalid object key")
	}

	return fullPath, nil
}
