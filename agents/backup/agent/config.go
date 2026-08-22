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
}

type MinioConfig struct {
	Endpoint string
	Access   string
	Secret   string
	Bucket   string
	Secure   bool
	Region   string
}
