package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// VaultStore — WORM-хранилище (object lock + versioning) для неизменяемых
// копий снапшотов. Отдельный бакет, недоступный prune и любым другим
// операциям агента: записывать можно, перезаписать/удалить — нет, пока не
// истечёт retention (Compliance) или не снят legal hold (Governance).
// Object lock в MinIO включается только при создании бакета — существующий
// бакет без lock не «долечивается», vault просто деградирует в warning.

type VaultStore struct {
	client   *minio.Client
	bucket   string
	mode     minio.RetentionMode
	cfg      VaultConfig
	region   string
	endpoint string
	// enabled=false: бакета нет или он создан без object lock.
	enabled bool
}

func NewVaultStore(cfg VaultConfig) *VaultStore {
	v := &VaultStore{bucket: cfg.Bucket, cfg: cfg}
	v.mode = minio.Governance
	if strings.EqualFold(strings.TrimSpace(cfg.Mode), "compliance") {
		v.mode = minio.Compliance
	}
	if v.bucket == "" {
		v.bucket = "vault"
	}
	return v
}

func (v *VaultStore) Enabled() bool { return v != nil && v.enabled }
func (v *VaultStore) Bucket() string {
	if v == nil {
		return ""
	}
	return v.bucket
}

// Bind привязывает vault к тому же S3-endpoint, что и рабочее хранилище.
// Отдельные учётные данные (MINIO_VAULT_*) позволяют отдать сейфу сервисный
// аккаунт с правами только на vault-бакет: компрометация основного ключа
// не даёт стереть сейф до истечения retention.
func (v *VaultStore) Bind(cfg MinioConfig) error {
	if v == nil {
		return fmt.Errorf("vault store is nil")
	}
	access := firstNonEmpty(os.Getenv("MINIO_VAULT_ACCESS_KEY"), cfg.Access)
	secret := firstNonEmpty(os.Getenv("MINIO_VAULT_SECRET_KEY"), cfg.Secret)
	cl, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: cfg.Secure,
		Region: cfg.Region,
	})
	if err != nil {
		return err
	}
	v.client = cl
	v.region = cfg.Region
	v.endpoint = cfg.Endpoint
	return nil
}

// SameEndpoint — true если стор снапшота смотрит на тот же S3-endpoint,
// что и сейф: тогда копия делается серверным CopyObject, иначе — потоком
// через агента (CopyRetainedFrom).
func (v *VaultStore) SameEndpoint(cfg MinioConfig) bool {
	return v != nil && strings.TrimSpace(v.endpoint) != "" &&
		strings.TrimSpace(v.endpoint) == strings.TrimSpace(cfg.Endpoint)
}

// Ensure создаёт vault-бакет с object lock и версионированием, либо
// проверяет, что существующий бакет годится. Не fatal: сейф — надстройка.
func (v *VaultStore) Ensure(ctx context.Context) error {
	if v == nil || v.client == nil {
		return fmt.Errorf("vault store is not bound")
	}
	ok, err := v.client.BucketExists(ctx, v.bucket)
	if err != nil {
		return fmt.Errorf("vault bucket %s: %w", v.bucket, err)
	}
	if !ok {
		if err := v.client.MakeBucket(ctx, v.bucket, minio.MakeBucketOptions{
			Region:        v.region,
			ObjectLocking: true,
		}); err != nil {
			return fmt.Errorf("vault MakeBucket: %w", err)
		}
		log.Printf("vault: bucket %s created with object lock (%s)", v.bucket, v.mode)
	}
	if err := v.client.EnableVersioning(ctx, v.bucket); err != nil {
		log.Printf("vault: enable versioning: %v", err)
	}
	// DefaultRetention защищает объекты, для которых retention не задан явно
	// (например блоки restic-репозитория, которые restic пишет сам).
	if v.cfg.RetentionDays > 0 {
		days := uint(v.cfg.RetentionDays)
		unit := minio.Days
		if err := v.client.SetObjectLockConfig(ctx, v.bucket, &v.mode, &days, &unit); err != nil {
			log.Printf("vault: default retention: %v", err)
		}
	}
	// Проверяем, что lock реально работает: конфиг должен читаться без ошибки.
	if _, _, _, _, err := v.client.GetObjectLockConfig(ctx, v.bucket); err != nil {
		v.enabled = false
		return fmt.Errorf("vault bucket %s exists without object lock (%v); recreate it with object locking", v.bucket, err)
	}
	v.enabled = true
	return nil
}

// CopyRetained копирует объект из рабочего бакета в сейф с retention-датой.
// Для объектов >5GiB minio-go сам перекладывает через multipart-copy.
// Работает только в пределах endpoint'а сейфа (server-side copy).
func (v *VaultStore) CopyRetained(ctx context.Context, srcBucket, srcKey, dstKey string, until time.Time) error {
	if !v.Enabled() {
		return fmt.Errorf("vault is not enabled")
	}
	dst, err := v.client.CopyObject(ctx, minio.CopyDestOptions{
		Bucket:          v.bucket,
		Object:          dstKey,
		Mode:            v.mode,
		RetainUntilDate: until,
	}, minio.CopySrcOptions{Bucket: srcBucket, Object: srcKey})
	if err != nil {
		return fmt.Errorf("vault copy %s: %w", srcKey, err)
	}
	log.Printf("vault: %s -> %s (%d bytes, locked until %s)",
		srcKey, dstKey, dst.Size, until.UTC().Format(time.RFC3339))
	return nil
}

// CopyRetainedFrom — потоковый вариант для снапшотов на другом S3-endpoint:
// server-side copy между серверами невозможен, объект качается из srcStore
// и заливается в сейф с retention-заголовками.
func (v *VaultStore) CopyRetainedFrom(ctx context.Context, srcStore Storage, srcBucket, srcKey, dstKey string, until time.Time) error {
	if !v.Enabled() {
		return fmt.Errorf("vault is not enabled")
	}
	rc, err := srcStore.GetBucket(ctx, srcBucket, srcKey)
	if err != nil {
		return fmt.Errorf("vault copy %s: read source: %w", srcKey, err)
	}
	defer rc.Close()
	info, err := v.client.PutObject(ctx, v.bucket, dstKey, rc, -1, minio.PutObjectOptions{
		Mode:            v.mode,
		RetainUntilDate: until,
	})
	if err != nil {
		return fmt.Errorf("vault upload %s: %w", dstKey, err)
	}
	log.Printf("vault: %s -> %s (%d bytes via agent, locked until %s)",
		srcKey, dstKey, info.Size, until.UTC().Format(time.RFC3339))
	return nil
}

// Get читает объект из сейфа (объекты под retention/legal hold доступны
// на чтение — именно это делает сейф восстановимым после атаки).
func (v *VaultStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if !v.Enabled() {
		return nil, fmt.Errorf("vault is not enabled")
	}
	return v.client.GetObject(ctx, v.bucket, key, minio.GetObjectOptions{})
}

// SetLegalHold блокирует/разблокирует объект независимо от retention.
// Применяется к «чистым» точкам восстановления: Governance + legal hold
// неудаляемы даже для владельца бакета.
func (v *VaultStore) SetLegalHold(ctx context.Context, key string, on bool) error {
	if !v.Enabled() {
		return fmt.Errorf("vault is not enabled")
	}
	status := minio.LegalHoldEnabled
	if !on {
		status = minio.LegalHoldDisabled
	}
	return v.client.PutObjectLegalHold(ctx, v.bucket, key, minio.PutObjectLegalHoldOptions{Status: &status})
}
