package agent

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/hirochachacha/go-smb2"
)

// SMBStore — сетевая шара Windows/Samba как назначение. Учётка из
// credentials (username + пароль в BACKUP_SECRET_<handle>); порт/домен —
// через поле options назначения (JSON {"port":445,"domain":"X"}).

type SMBStore struct {
	session *smb2.Session
	share   *smb2.Share
	root    string
	host    string
	shareNm string
}

func NewSMBStore(host, share, username, password, root, optionsJSON string) (*SMBStore, error) {
	host = strings.TrimSpace(host)
	share = strings.TrimSpace(share)
	if host == "" || share == "" {
		return nil, fmt.Errorf("smb destination requires host and share")
	}
	fs, err := DialSMB(host, username, password, optionsJSON, share, root)
	if err != nil {
		return nil, err
	}
	return &SMBStore{
		session: fs.session,
		share:   fs.share,
		root:    fs.root,
		host:    host,
		shareNm: share,
	}, nil
}

func (s *SMBStore) EnsureBucket(ctx context.Context) error {
	if s.root == "" {
		return nil
	}
	return s.share.MkdirAll(s.root, 0o755)
}

func (s *SMBStore) Bucket() string {
	if s.root == "" {
		return s.shareNm
	}
	return path.Join(s.shareNm, s.root)
}
func (s *SMBStore) Location() string {
	return fmt.Sprintf("smb://%s/%s", s.host, s.Bucket())
}
func (s *SMBStore) Config() MinioConfig { return MinioConfig{} }

func (s *SMBStore) Ping(ctx context.Context) error {
	if _, err := s.share.ReadDir(s.rootOr(".")); err != nil {
		return err
	}
	// Проверка записи: временный файл.
	probe := s.rootOr(".backup-dest-probe")
	f, err := s.share.Create(probe)
	if err != nil {
		return fmt.Errorf("share is not writable: %w", err)
	}
	f.Close()
	return s.share.Remove(probe)
}

func (s *SMBStore) rootOr(fallback string) string {
	if s.root == "" {
		return fallback
	}
	return s.root
}

func (s *SMBStore) fullKey(key string) (string, error) {
	k, err := safeStorageKey(key)
	if err != nil {
		return "", err
	}
	if s.root == "" {
		return k, nil
	}
	return path.Join(s.root, k), nil
}

func (s *SMBStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return 0, err
	}
	if dir := path.Dir(target); dir != "." && dir != "/" && dir != "" {
		if err := s.share.MkdirAll(dir, 0o755); err != nil {
			return 0, err
		}
	}
	f, err := s.share.Create(target)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	return io.Copy(f, r)
}

func (s *SMBStore) PutBucket(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) (int64, error) {
	return s.Put(ctx, key, r, size, contentType)
}

func (s *SMBStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return nil, err
	}
	return s.share.Open(target)
}

func (s *SMBStore) GetBucket(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.Get(ctx, key)
}

func (s *SMBStore) ListBucket(ctx context.Context, bucket, prefix string) ([]string, error) {
	prefix = strings.Trim(prefix, "/")
	root := s.rootOr(".")
	var out []string
	err := s.walk(root, func(p string) {
		rel := strings.TrimPrefix(strings.TrimPrefix(p, root), "/")
		if strings.HasPrefix(rel, prefix) {
			out = append(out, rel)
		}
	})
	return out, err
}

func (s *SMBStore) walk(dir string, fn func(string)) error {
	entries, err := s.share.ReadDir(dir)
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

func (s *SMBStore) RemoveBucket(ctx context.Context, bucket, key string) error {
	target, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return s.share.Remove(target)
}

func (s *SMBStore) Remove(ctx context.Context, key string) error {
	return s.RemoveBucket(ctx, "", key)
}

func (s *SMBStore) Close() error {
	if s.share != nil {
		_ = s.share.Umount()
	}
	if s.session != nil {
		return s.session.Logoff()
	}
	return nil
}
