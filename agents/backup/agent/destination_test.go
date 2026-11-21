package agent

import (
	"testing"
)

func TestMinioConfigFor(t *testing.T) {
	a := &Agent{cfg: Config{Minio: MinioConfig{
		Endpoint: "minio:9000",
		Access:   "agent-key",
		Secret:   "agent-secret",
		Bucket:   "backups",
		Region:   "ru-1",
	}}}

	// nil — дефолт агента.
	if got := a.minioConfigFor(nil); got != a.cfg.Minio {
		t.Fatalf("nil dest must return agent cfg, got %+v", got)
	}

	// Полное назначение перекрывает всё.
	full := &Destination{
		Endpoint: "other:9000", Bucket: "b2", Region: "eu",
		Secure: true, Access: "k", Secret: "s",
	}
	got := a.minioConfigFor(full)
	if got.Endpoint != "other:9000" || got.Bucket != "b2" || got.Region != "eu" ||
		!got.Secure || got.Access != "k" || got.Secret != "s" {
		t.Fatalf("full dest: %+v", got)
	}

	// Bucket-only (legacy dest_bucket): endpoint/учётки наследуются.
	legacy := a.minioConfigFor(LegacyDest("tenant"))
	if legacy.Bucket != "tenant" || legacy.Endpoint != "minio:9000" ||
		legacy.Access != "agent-key" || legacy.Secret != "agent-secret" {
		t.Fatalf("legacy dest: %+v", legacy)
	}

	// Назначение без учёток наследует ключи агента.
	noCreds := &Destination{Bucket: "b3"}
	got = a.minioConfigFor(noCreds)
	if got.Access != "agent-key" || got.Secret != "agent-secret" || got.Bucket != "b3" {
		t.Fatalf("no-creds dest: %+v", got)
	}
}

func TestStoreCacheKey(t *testing.T) {
	base := MinioConfig{Endpoint: "e:9000", Access: "k", Bucket: "b", Region: "r"}
	if storeCacheKey(base) != storeCacheKey(base) {
		t.Fatal("same cfg — same key")
	}
	// Раньше ключом был только бакет: меняем endpoint и учётки при том же бакете.
	other := base
	other.Endpoint = "other:9000"
	if storeCacheKey(base) == storeCacheKey(other) {
		t.Fatal("endpoint must be part of the key")
	}
	other = base
	other.Access = "k2"
	if storeCacheKey(base) == storeCacheKey(other) {
		t.Fatal("access key must be part of the key")
	}
	other = base
	other.Bucket = "b2"
	if storeCacheKey(base) == storeCacheKey(other) {
		t.Fatal("bucket must be part of the key")
	}
}

func TestCachedStoreReturnsDefaultForAgentCfg(t *testing.T) {
	def, err := NewObjectStore(MinioConfig{Endpoint: "e:9000", Access: "k", Secret: "s", Bucket: "b"})
	if err != nil {
		t.Fatal(err)
	}
	a := &Agent{cfg: Config{Minio: def.Config()}, store: def}
	// Дефолтная конфигурация — вернётся основной стор без создания нового.
	got, err := a.cachedStore(t.Context(), a.cfg.Minio, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != def {
		t.Fatal("agent cfg must resolve to the default store")
	}
	// Другой бакет — новый клиент, кэшируется по полному ключу.
	other := a.cfg.Minio
	other.Bucket = "other"
	s1, err := a.cachedStore(t.Context(), other, false)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := a.cachedStore(t.Context(), other, false)
	if err != nil {
		t.Fatal(err)
	}
	if s1 != s2 {
		t.Fatal("second call must hit the cache")
	}
	if len(a.stores) != 1 {
		t.Fatalf("cache size %d, want 1", len(a.stores))
	}
}

func TestDestinationIsRecord(t *testing.T) {
	if LegacyDest("x").IsRecord() {
		t.Fatal("legacy dest is not a record")
	}
	if !(&Destination{ID: 7}).IsRecord() {
		t.Fatal("record-backed dest")
	}
	if (&Destination{}).IsRecord() {
		t.Fatal("zero dest")
	}
}
