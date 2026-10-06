package vacuum

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Storage — Input: s3_key (string) -> Process: open byte stream -> Output: io.ReadCloser (no data loss, streaming)
// Mental model: same pipeline works for local FS and S3. For dev we use filesystem root; inline: prefix for tests.
type Storage interface {
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Put(ctx context.Context, key string, r io.Reader) error
}

type FSStorage struct{ Root string }

func NewFSStorage(root string) *FSStorage {
	if root == "" {
		root = os.Getenv("VACUUM_STORAGE_ROOT")
		if root == "" {
			root = "/tmp/vacuum_storage"
		}
	}
	_ = os.MkdirAll(root, 0o755)
	return &FSStorage{Root: root}
}

func (s *FSStorage) resolve(key string) string {
	k := strings.TrimPrefix(key, "/")
	// prevent directory traversal
	k = filepath.Clean(k)
	if strings.HasPrefix(k, "..") {
		k = strings.TrimPrefix(k, "..")
		k = strings.TrimPrefix(k, string(filepath.Separator))
	}
	return filepath.Join(s.Root, k)
}

func (s *FSStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.HasPrefix(key, "inline:") {
		txt := strings.TrimPrefix(key, "inline:")
		return io.NopCloser(strings.NewReader(txt)), nil
	}
	p := s.resolve(key)
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("storage open %s: %w", key, err)
	}
	return f, nil
}

func (s *FSStorage) Put(ctx context.Context, key string, r io.Reader) error {
	if strings.HasPrefix(key, "inline:") {
		return fmt.Errorf("inline key not writable")
	}
	p := s.resolve(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
