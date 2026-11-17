package agent

import "time"

type Config struct {
	ListenAddr     string
	PublicURL      string
	CortezaAPI     string
	Token          string
	NamespaceID    uint64
	PollInterval   time.Duration
	Minio          MinioConfig
	ResticPassword string
	Concurrency    int
	Vault          VaultConfig
	// ClamAVAddr — TCP "host:3310" или unix:///путь/к/сокету; пусто — только эвристики.
	ClamAVAddr string
	// ScanMaxMB — суммарный лимит байт на сканирование одного архива.
	ScanMaxMB int64
	// DockerHost — endpoint dockerd для верификации уровня l2
	// (unix:///var/run/docker.sock или tcp://verify-worker:2375).
	DockerHost string
}

type VaultConfig struct {
	Bucket        string // "vault"
	Mode          string // "compliance" | "governance", default compliance
	RetentionDays int    // per-object retention для не-restic объектов
	Enabled       bool   // false если бакет без object lock
}

type MinioConfig struct {
	Endpoint string
	Access   string
	Secret   string
	Bucket   string
	Secure   bool
	Region   string
}
