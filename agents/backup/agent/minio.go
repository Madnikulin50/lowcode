package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectStore struct {
	client *minio.Client
	cfg    MinioConfig
}

func NewObjectStore(cfg MinioConfig) (*ObjectStore, error) {
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, fmt.Errorf("minio endpoint is empty")
	}
	if cfg.Bucket == "" {
		cfg.Bucket = "backups"
	}
	cl, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Access, cfg.Secret, ""),
		Secure: cfg.Secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &ObjectStore{client: cl, cfg: cfg}, nil
}

func (s *ObjectStore) EnsureBucket(ctx context.Context) error {
	ok, err := s.client.BucketExists(ctx, s.cfg.Bucket)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return s.client.MakeBucket(ctx, s.cfg.Bucket, minio.MakeBucketOptions{Region: s.cfg.Region})
}

func (s *ObjectStore) Bucket() string { return s.cfg.Bucket }

func (s *ObjectStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	info, err := s.client.PutObject(ctx, s.cfg.Bucket, key, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (s *ObjectStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.cfg.Bucket, key, minio.GetObjectOptions{})
}

func (s *ObjectStore) Remove(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.cfg.Bucket, key, minio.RemoveObjectOptions{})
}

func (s *ObjectStore) List(ctx context.Context, prefix string) ([]string, error) {
	var out []string
	for obj := range s.client.ListObjects(ctx, s.cfg.Bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		out = append(out, obj.Key)
	}
	return out, nil
}

func (s *ObjectStore) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.cfg.Bucket)
	return err
}

func ObjectKey(sourceHandle string, jobID string, filename string) string {
	now := time.Now().UTC()
	src := sanitizeKey(sourceHandle)
	if src == "" {
		src = "unknown"
	}
	if jobID == "" {
		jobID = now.Format("150405")
	}
	if filename == "" {
		filename = "data.bin"
	}
	return path.Join(src, now.Format("2006"), now.Format("01"), jobID, filename)
}

func ResticPrefix(sourceHandle string) string {
	return path.Join("restic", sanitizeKey(sourceHandle)) + "/"
}

func sanitizeKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\\", "/")
	s = strings.Trim(s, "/")
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}

type countingHasher struct {
	w io.Writer
	n int64
	h hash.Hash
}

func newCountingHasher(w io.Writer) *countingHasher {
	return &countingHasher{w: w, h: sha256.New()}
}

func (c *countingHasher) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	if n > 0 {
		c.n += int64(n)
		c.h.Write(p[:n])
	}
	return n, err
}

func (c *countingHasher) SumHex() string { return hex.EncodeToString(c.h.Sum(nil)) }
func (c *countingHasher) Bytes() int64   { return c.n }

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "backup-agent"
	}
	return h
}
