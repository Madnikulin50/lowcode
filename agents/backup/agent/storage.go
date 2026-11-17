package agent

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// Storage — единый контракт хранилища бэкапов. Появился, когда назначения
// перестали быть только S3: реализации — S3 (ObjectStore), FileStore
// (local/nfs), SFTPStore, SMBStore, RcloneStore, WebDAVStore.
//
// Семантика ключей — как у S3: "a/b/c.tar.gz" без ведущего слэша. Бэкенды,
// где нет понятия бакета (всё кроме S3), игнорируют параметр bucket в
// *Bucket-методах и работают со своим корнем.
type Storage interface {
	// EnsureBucket создаёт корень/бакет, если его нет (путь, базовый каталог,
	// remote). Для read-only бэкендов может быть no-op.
	EnsureBucket(ctx context.Context) error
	// Bucket — логическое имя хранилища (бакет S3, корень file/nfs,
	// base-dir sftp/webdav, share+subdir smb, remote-path rclone).
	Bucket() string
	// Location — самодостаточный локатор для записи в снапшот и последующего
	// переподключения: endpoint S3, file:///path, sftp://…, rclone:…, smb://…
	Location() string
	// Config — MinioConfig для S3-хранилищ; для остальных — нулевое значение.
	// Нужен restic/kopia (s3-репозитории) и сейфу (server-side copy).
	Config() MinioConfig
	// Ping — лёгкая проверка доступности (dest-test).
	Ping(ctx context.Context) error

	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error)
	PutBucket(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) (int64, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	GetBucket(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	ListBucket(ctx context.Context, bucket, prefix string) ([]string, error)
	RemoveBucket(ctx context.Context, bucket, key string) error
	Remove(ctx context.Context, key string) error
	// Close освобождает соединения (sftp/smb). Для stateless — no-op.
	Close() error
}

// Location реализован ObjectStore-ом в minio.go (endpoint).

// safeStorageKey отсекает key traversal для файловых бэкендов: ключи у нас
// генерируются sanitizeKey + path.Join, но от пришедшего извне "../../../../"
// защищать всё равно нужно.
func safeStorageKey(key string) (string, error) {
	key = strings.TrimSpace(strings.ReplaceAll(key, "\\", "/"))
	key = strings.TrimPrefix(key, "/")
	if key == "" || key == "." || strings.HasPrefix(key, "..") || strings.Contains(key, "/../") {
		return "", fmt.Errorf("invalid storage key %q", key)
	}
	return key, nil
}

// StoreKind определяет тип назначения по Destination.
func StoreKind(dest *Destination) string {
	if dest == nil || strings.TrimSpace(dest.Type) == "" {
		return "minio"
	}
	return strings.ToLower(strings.TrimSpace(dest.Type))
}
