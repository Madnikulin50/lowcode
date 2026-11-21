package agent

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// sshDialTimeout — таймаут установки SSH-соединения к назначению.
const sshDialTimeout = 15 * time.Second

// sshAgentSigners — ключи из SSH-agent пользователя, запустившего агента
// (fallback-аутентификация, когда в credentials нет ни пароля, ни ключа).
func sshAgentSigners() ([]ssh.Signer, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return agent.NewClient(conn).Signers()
}

// SFTPStore — удалённый каталог по SSH/SFTP как назначение. Учётка — из
// связанной записи credentials: username + BACKUP_SECRET_<handle>. Секрет
// может быть паролем или PEM-ключом (начинается с "-----BEGIN").

type SFTPStore struct {
	client *sftp.Client
	ssh    *ssh.Client
	base   string
	loc    string
}

func NewSFTPStore(host string, port int, username, secret, baseDir string) (*SFTPStore, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("sftp host is empty")
	}
	if port <= 0 {
		port = 22
	}
	var auth []ssh.AuthMethod
	secret = strings.TrimSpace(secret)
	if strings.HasPrefix(secret, "-----BEGIN") {
		signer, err := ssh.ParsePrivateKey([]byte(secret))
		if err != nil {
			return nil, fmt.Errorf("sftp private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else if secret != "" {
		auth = append(auth, ssh.Password(secret))
	} else {
		// Ключ агента пользователя, запустившего агента.
		if signers, err := sshAgentSigners(); err == nil && len(signers) > 0 {
			auth = append(auth, ssh.PublicKeys(signers...))
		}
	}
	cfg := &ssh.ClientConfig{
		User:            username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // бэкап-назначение; строгий ключ задаётся позже при необходимости
		Timeout:         sshDialTimeout,
	}
	sshCli, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), cfg)
	if err != nil {
		return nil, fmt.Errorf("sftp dial %s:%d: %w", host, port, err)
	}
	cli, err := sftp.NewClient(sshCli)
	if err != nil {
		sshCli.Close()
		return nil, fmt.Errorf("sftp client: %w", err)
	}
	base := "/" + strings.Trim(strings.TrimSpace(baseDir), "/")
	if base == "/" {
		base = "."
	}
	loc := fmt.Sprintf("sftp://%s@%s:%d%s", username, host, port, base)
	return &SFTPStore{client: cli, ssh: sshCli, base: base, loc: loc}, nil
}

func (s *SFTPStore) EnsureBucket(ctx context.Context) error {
	if s.base == "." {
		return nil
	}
	return s.client.MkdirAll(s.base)
}

func (s *SFTPStore) Bucket() string   { return s.base }
func (s *SFTPStore) Location() string { return s.loc }
func (s *SFTPStore) Config() MinioConfig {
	return MinioConfig{}
}
func (s *SFTPStore) Ping(ctx context.Context) error {
	_, err := s.client.ReadDir(".")
	return err
}

func (s *SFTPStore) fullKey(key string) (string, error) {
	k, err := safeStorageKey(key)
	if err != nil {
		return "", err
	}
	if s.base == "." {
		return k, nil
	}
	return path.Join(s.base, k), nil
}

func (s *SFTPStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return 0, err
	}
	if dir := path.Dir(target); dir != "." && dir != "/" {
		if err := s.client.MkdirAll(dir); err != nil {
			return 0, err
		}
	}
	f, err := s.client.Create(target)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	return io.Copy(f, r)
}

func (s *SFTPStore) PutBucket(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) (int64, error) {
	return s.Put(ctx, key, r, size, contentType)
}

func (s *SFTPStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	target, err := s.fullKey(key)
	if err != nil {
		return nil, err
	}
	return s.client.Open(target)
}

func (s *SFTPStore) GetBucket(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.Get(ctx, key)
}

func (s *SFTPStore) ListBucket(ctx context.Context, bucket, prefix string) ([]string, error) {
	prefix = strings.Trim(prefix, "/")
	base := s.base
	if base == "." {
		base = ""
	}
	var out []string
	walker := s.client.Walk(base)
	for walker.Step() {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if walker.Err() != nil {
			continue
		}
		p := walker.Path()
		if walker.Stat().IsDir() {
			continue
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, base), "/")
		if strings.HasPrefix(rel, prefix) {
			out = append(out, rel)
		}
	}
	return out, nil
}

func (s *SFTPStore) RemoveBucket(ctx context.Context, bucket, key string) error {
	target, err := s.fullKey(key)
	if err != nil {
		return err
	}
	return s.client.Remove(target)
}

func (s *SFTPStore) Remove(ctx context.Context, key string) error {
	return s.RemoveBucket(ctx, "", key)
}

func (s *SFTPStore) Close() error {
	if s.client != nil {
		s.client.Close()
	}
	if s.ssh != nil {
		return s.ssh.Close()
	}
	return nil
}
