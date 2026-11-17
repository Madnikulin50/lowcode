package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FileStore — локальная файловая система агента как назначение (local)
// или как пишущая сторона смонтированного NFS. Ключи — относительные пути
// под корнем; корень создаётся при первом использовании.

type FileStore struct {
	root string
}

func NewFileStore(root string) (*FileStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("local destination path is empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &FileStore{root: abs}, nil
}

func (s *FileStore) EnsureBucket(ctx context.Context) error {
	return os.MkdirAll(s.root, 0o755)
}

func (s *FileStore) Bucket() string   { return s.root }
func (s *FileStore) Location() string { return "file://" + s.root }
func (s *FileStore) Config() MinioConfig {
	return MinioConfig{}
}
func (s *FileStore) Ping(ctx context.Context) error {
	if _, err := os.Stat(s.root); err != nil {
		return err
	}
	// Проверяем, что в корень реально можно писать (read-only mount — частая
	// причина «dest-test ok, backup failed»).
	probe := filepath.Join(s.root, ".backup-dest-probe")
	f, err := os.Create(probe)
	if err != nil {
		return fmt.Errorf("root is not writable: %w", err)
	}
	f.Close()
	os.Remove(probe)
	return nil
}

func (s *FileStore) fullKey(key string) (string, error) {
	k, err := safeStorageKey(key)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.root, filepath.FromSlash(k)), nil
}

func (s *FileStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return 0, err
	}
	// Read-only существующий файл нельзя перезаписать даже владельцу.
	if fi, statErr := os.Lstat(target); statErr == nil && fi.Mode().IsRegular() && fi.Mode().Perm()&0o200 == 0 {
		if rmErr := os.Remove(target); rmErr != nil {
			return 0, fmt.Errorf("remove read-only %s: %w", target, rmErr)
		}
	}
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	return io.Copy(f, r)
}

func (s *FileStore) PutBucket(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) (int64, error) {
	return s.Put(ctx, key, r, size, contentType)
}

func (s *FileStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return nil, err
	}
	return os.Open(target)
}

func (s *FileStore) GetBucket(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.Get(ctx, key)
}

func (s *FileStore) ListBucket(ctx context.Context, bucket, prefix string) ([]string, error) {
	base := s.root
	rootDepth := len(strings.Split(filepath.Clean(base), string(os.PathSeparator)))
	var out []string
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // недоступный файл/каталог пропускаем
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		parts := strings.Split(filepath.Clean(p), string(os.PathSeparator))
		rel := strings.Join(parts[rootDepth:], "/")
		if strings.HasPrefix(rel, strings.Trim(prefix, "/")) {
			out = append(out, rel)
		}
		return nil
	})
	return out, err
}

func (s *FileStore) RemoveBucket(ctx context.Context, bucket, key string) error {
	target, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return os.Remove(target)
}

func (s *FileStore) Remove(ctx context.Context, key string) error {
	return s.RemoveBucket(ctx, "", key)
}

func (s *FileStore) Close() error { return nil }

// --- NFS -------------------------------------------------------------------

// nfsMount гарантирует, что server:/export смонтирован в mountpoint, и
// возвращает FileStore поверх него. У агента обычно нет прав на mount —
// тогда честная ошибка с инструкцией (fstab + sudo или systemd mount).
func nfsMount(ctx context.Context, server, mountpoint, options string) (*FileStore, error) {
	server = strings.TrimSpace(server)
	mountpoint = strings.TrimSpace(mountpoint)
	if server == "" || mountpoint == "" {
		return nil, fmt.Errorf("nfs destination requires host (server:/export) and path (mountpoint)")
	}
	if mounted, err := isMounted(mountpoint); err != nil {
		return nil, err
	} else if !mounted {
		args := []string{"-t", "nfs", server, mountpoint}
		if strings.TrimSpace(options) != "" {
			args = append(args, "-o", strings.TrimSpace(options))
		}
		cmd := exec.CommandContext(ctx, "mount", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("nfs %s is not mounted at %s and auto-mount failed: %v: %s — mount it manually (fstab/systemd) and retry", server, mountpoint, err, strings.TrimSpace(string(out)))
		}
	}
	return NewFileStore(mountpoint)
}

// isMounted проверяет /proc/self/mounts на предмет смонтированного каталога.
func isMounted(mountpoint string) (bool, error) {
	raw, err := os.ReadFile("/proc/self/mounts")
	if err != nil {
		return false, fmt.Errorf("read /proc/self/mounts: %w", err)
	}
	abs, err := filepath.Abs(mountpoint)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == abs {
			return true, nil
		}
	}
	return false, nil
}
