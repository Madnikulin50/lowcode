package agent

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/studio-b12/gowebdav"
)

// WebDAVStore — WebDAV-сервер (NAS, Nextcloud и т.п.) как назначение.
// Учётка из credentials (username + пароль в BACKUP_SECRET_<handle>);
// без записи — анонимный доступ.

type WebDAVStore struct {
	cli  *gowebdav.Client
	base string
	loc  string
}

func NewWebDAVStore(endpoint, username, password, baseDir string) (*WebDAVStore, error) {
	endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	if endpoint == "" {
		return nil, fmt.Errorf("webdav endpoint (URL) is empty")
	}
	cli := gowebdav.NewClient(endpoint, username, password)
	base := "/" + strings.Trim(strings.TrimSpace(baseDir), "/")
	if base == "/" {
		base = ""
	}
	loc := "webdav+" + endpoint + base
	return &WebDAVStore{cli: cli, base: base, loc: loc}, nil
}

func (s *WebDAVStore) EnsureBucket(ctx context.Context) error {
	if s.base == "" {
		return nil
	}
	return s.cli.MkdirAll(s.base, 0o755)
}

func (s *WebDAVStore) Bucket() string   { return s.base }
func (s *WebDAVStore) Location() string { return s.loc }
func (s *WebDAVStore) Config() MinioConfig {
	return MinioConfig{}
}
func (s *WebDAVStore) Ping(ctx context.Context) error {
	_, err := s.cli.ReadDir("/")
	return err
}

func (s *WebDAVStore) fullKey(key string) (string, error) {
	k, err := safeStorageKey(key)
	if err != nil {
		return "", err
	}
	if s.base == "" {
		return "/" + k, nil
	}
	return path.Join(s.base, k), nil
}

func (s *WebDAVStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return 0, err
	}
	if dir := path.Dir(target); dir != "/" && dir != "." {
		if err := s.cli.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
	}
	// gowebdav.WriteStream не возвращает число записанных байт — считаем через TeeReader.
	pr, pw := io.Pipe()
	var n int64
	done := make(chan error, 1)
	go func() {
		buf := make([]byte, 32*1024)
		var err error
		for {
			var nr int
			nr, err = r.Read(buf)
			if nr > 0 {
				if _, werr := pw.Write(buf[:nr]); werr != nil {
					err = werr
					break
				}
				n += int64(nr)
			}
			if err != nil {
				break
			}
			if ctx.Err() != nil {
				err = ctx.Err()
				break
			}
		}
		if err == io.EOF {
			err = nil
		}
		pw.CloseWithError(err)
		done <- err
	}()
	werr := s.cli.WriteStream(target, pr, 0o644)
	perr := <-done
	if werr != nil {
		return 0, werr
	}
	if perr != nil {
		return 0, perr
	}
	return n, nil
}

func (s *WebDAVStore) PutBucket(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) (int64, error) {
	return s.Put(ctx, key, r, size, contentType)
}

func (s *WebDAVStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return nil, err
	}
	return s.cli.ReadStream(target)
}

func (s *WebDAVStore) GetBucket(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.Get(ctx, key)
}

func (s *WebDAVStore) ListBucket(ctx context.Context, bucket, prefix string) ([]string, error) {
	prefix = strings.Trim(prefix, "/")
	base := s.base
	if base == "" {
		base = "/"
	}
	var out []string
	err := s.walk(base, func(p string) {
		rel := strings.TrimPrefix(p, s.base)
		rel = strings.TrimPrefix(rel, "/")
		if strings.HasPrefix(rel, prefix) {
			out = append(out, rel)
		}
	})
	return out, err
}

func (s *WebDAVStore) walk(dir string, fn func(string)) error {
	entries, err := s.cli.ReadDir(dir)
	if err != nil {
		return nil // недоступный каталог пропускаем
	}
	for _, e := range entries {
		p := path.Join(dir, e.Name())
		if e.IsDir() {
			if err := s.walk(p, fn); err != nil {
				return err
			}
			continue
		}
		fn(p)
	}
	return nil
}

func (s *WebDAVStore) RemoveBucket(ctx context.Context, bucket, key string) error {
	target, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return s.cli.Remove(target)
}

func (s *WebDAVStore) Remove(ctx context.Context, key string) error {
	return s.RemoveBucket(ctx, "", key)
}

func (s *WebDAVStore) Close() error { return nil }
