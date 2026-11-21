package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RcloneStore — назначение через rclone: один бинарь покрывает 70+ бэкендов
// (Google Drive, Dropbox, GCS, Azure, Яндекс.Диск, WebDAV-разновидности...).
// В записи назначения: remote — имя remote из rclone-конфига агента
// (rclone config), path — подкаталог внутри remote.

type RcloneStore struct {
	remoteSpec string // "remoteName:path" как единый аргумент для rclone
	loc        string
}

func rcloneAvailable() bool {
	_, err := exec.LookPath("rclone")
	return err == nil
}

func NewRcloneStore(remote, basePath string) (*RcloneStore, error) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return nil, fmt.Errorf("rclone destination requires remote name (from agent's rclone config)")
	}
	basePath = strings.Trim(strings.TrimSpace(basePath), "/")
	spec := remote
	if basePath != "" {
		spec = remote + ":" + basePath
	}
	return &RcloneStore{remoteSpec: spec, loc: "rclone:" + spec}, nil
}

// rcloneCtx — таймаут на одну rclone-команду (большие архивы — долго).
const rcloneTimeout = 6 * time.Hour

func (s *RcloneStore) run(ctx context.Context, stdin io.Reader, args ...string) ([]byte, error) {
	full := append([]string{"--config=" + rcloneConfigPath(), "--low-level-retries", "10", "--retries", "3"}, args...)
	cmd := exec.CommandContext(ctx, "rclone", full...)
	cmd.Stdin = stdin
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 400 {
			msg = msg[:400]
		}
		return out, fmt.Errorf("rclone %s: %w: %s", args[0], err, msg)
	}
	return out, nil
}

// rcloneConfigPath — конфиг rclone агента; по умолчанию ~/.config/rclone/rclone.conf.
func rcloneConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("RCLONE_CONFIG")); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "rclone", "rclone.conf")
}

func (s *RcloneStore) EnsureBucket(ctx context.Context) error {
	_, err := s.run(ctx, nil, "mkdir", s.remoteSpec)
	return err
}

func (s *RcloneStore) Bucket() string   { return s.remoteSpec }
func (s *RcloneStore) Location() string { return s.loc }
func (s *RcloneStore) Config() MinioConfig {
	return MinioConfig{}
}
func (s *RcloneStore) Ping(ctx context.Context) error {
	_, err := s.run(ctx, nil, "lsd", s.remoteSpec)
	if err != nil {
		return err
	}
	probeKey := ".backup-dest-probe"
	if _, err := s.Put(ctx, probeKey, strings.NewReader("ok"), 2, ""); err != nil {
		return fmt.Errorf("remote is not writable: %w", err)
	}
	return s.Remove(ctx, probeKey)
}

func (s *RcloneStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	k, err := safeStorageKey(key)
	if err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(ctx, rcloneTimeout)
	defer cancel()
	// rcat пишет stdin на remote; число байт считаем TeeReader-ом.
	var n int64
	counter := &countingReader{r: r, n: &n}
	_, err = s.run(ctx, counter, "rcat", s.remoteSpec+"/"+k)
	return n, err
}

func (s *RcloneStore) PutBucket(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) (int64, error) {
	return s.Put(ctx, key, r, size, contentType)
}

func (s *RcloneStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	k, err := safeStorageKey(key)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, rcloneTimeout)
	out, err := s.run(ctx, nil, "cat", s.remoteSpec+"/"+k)
	if err != nil {
		cancel()
		return nil, err
	}
	return &readCloserWithCancel{Reader: bytes.NewReader(out), cancel: cancel}, nil
}

func (s *RcloneStore) GetBucket(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.Get(ctx, key)
}

func (s *RcloneStore) ListBucket(ctx context.Context, bucket, prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, rcloneTimeout)
	defer cancel()
	out, err := s.run(ctx, nil, "lsf", "--recursive", "--files-only", s.remoteSpec)
	if err != nil {
		return nil, err
	}
	prefix = strings.Trim(prefix, "/")
	var res []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, prefix) {
			res = append(res, line)
		}
	}
	return res, nil
}

func (s *RcloneStore) RemoveBucket(ctx context.Context, bucket, key string) error {
	k, err := safeStorageKey(key)
	if err != nil {
		return err
	}
	_, err = s.run(ctx, nil, "deletefile", s.remoteSpec+"/"+k)
	return err
}

func (s *RcloneStore) Remove(ctx context.Context, key string) error {
	return s.RemoveBucket(ctx, "", key)
}

func (s *RcloneStore) Close() error { return nil }

type countingReader struct {
	r io.Reader
	n *int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	nr, err := c.r.Read(p)
	*c.n += int64(nr)
	return nr, err
}

type readCloserWithCancel struct {
	io.Reader
	cancel context.CancelFunc
}

func (r *readCloserWithCancel) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
