package agent

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// Назначения (destinations) — явное управление тем, куда сохраняется бэкап.
// Резолвинг назначения для прогона (первое совпавшее):
//  1. destinationID из запроса (ad-hoc параметр backup/run);
//  2. запись destinations, привязанная к политике;
//  3. legacy dest_bucket политики — тот же MinIO агента, другой бакет;
//  4. назначение с флагом default = true;
//  5. nil — дефолт агента из env (MINIO_*).
//
// Типы назначений (Destination.Type): minio (S3-совместимое), local, nfs,
// sftp, smb, rclone, webdav. Реализации — storage*.go, общий контракт —
// Storage в storage.go. Учётки: username из связанной записи credentials,
// секрет — BACKUP_SECRET_<handle> в env агента; без учёток назначение
// наследует ключи агента (для minio) или работает анонимно/через SSH-agent.

func (a *Agent) resolveDestination(ctx context.Context, cz *Corteza, jr JobRequest, pol *Policy) (*Destination, error) {
	if id := ParseUint(jr.DestinationID); id > 0 {
		return a.loadDestination(ctx, cz, id)
	}
	if pol != nil && pol.DestinationID > 0 {
		return a.loadDestination(ctx, cz, pol.DestinationID)
	}
	if pol != nil && strings.TrimSpace(pol.DestBucket) != "" {
		return LegacyDest(strings.TrimSpace(pol.DestBucket)), nil
	}
	if cz != nil {
		if def, err := cz.DefaultDestination(ctx); err == nil && def != nil {
			return def, nil
		}
	}
	return nil, nil
}

// loadDestination читает запись destinations и подтягивает учётки.
func (a *Agent) loadDestination(ctx context.Context, cz *Corteza, id uint64) (*Destination, error) {
	dest, err := cz.LoadDestination(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("destination %d: %w", id, err)
	}
	if !dest.Enabled {
		return nil, fmt.Errorf("destination %d (%s) is disabled", id, dest.Name)
	}
	if dest.CredID > 0 {
		cred, err := cz.LoadCredential(ctx, dest.CredID)
		if err != nil {
			return nil, fmt.Errorf("destination %d: credential %d: %w", id, dest.CredID, err)
		}
		dest.Access = cred.Username
		dest.Secret = ResolveSecret(cred.Handle)
	}
	// Логин/пароль, введённые прямо в форме записи, перекрывают credentials.
	// NB: пароль хранится в записи открытым текстом — для production предпочтительнее credentials.
	if v := strings.TrimSpace(dest.Username); v != "" {
		dest.Access = v
	}
	if v := strings.TrimSpace(dest.Password); v != "" {
		dest.Secret = v
	}
	return dest, nil
}

// minioConfigFor собирает MinioConfig назначения: пустые поля назначения
// наследуются от дефолта агента (env). Только для типа minio.
func (a *Agent) minioConfigFor(dest *Destination) MinioConfig {
	cfg := a.cfg.Minio
	if dest == nil {
		return cfg
	}
	if v := strings.TrimSpace(dest.Endpoint); v != "" {
		cfg.Endpoint = v
	}
	if v := strings.TrimSpace(dest.Bucket); v != "" {
		cfg.Bucket = v
	}
	cfg.Region = dest.Region
	cfg.Secure = dest.Secure
	if dest.Access != "" {
		cfg.Access = dest.Access
	}
	if dest.Secret != "" {
		cfg.Secret = dest.Secret
	}
	return cfg
}

// destCacheKey различает кэшированные сторы назначений. Для minio — как
// раньше (endpoint+ключи+бакет); для остальных — полный набор полей, чтобы
// смена пути/share/remote давала новый стор.
func destCacheKey(dest *Destination) string {
	if StoreKind(dest) == "minio" {
		cfg := MinioConfig{Endpoint: dest.Endpoint, Access: dest.Access, Bucket: dest.Bucket, Region: dest.Region}
		return "minio|" + storeCacheKey(cfg)
	}
	return strings.Join([]string{
		StoreKind(dest), dest.Endpoint, fmt.Sprint(dest.Port), dest.Bucket,
		dest.Path, dest.Share, dest.Remote, dest.Access,
	}, "|")
}

// buildStorage создаёт стор по записи назначения (не-S3 типы и minio).
func (a *Agent) buildStorage(ctx context.Context, dest *Destination) (Storage, error) {
	if dest == nil {
		return nil, fmt.Errorf("destination is nil")
	}
	switch StoreKind(dest) {
	case "local":
		return NewFileStore(dest.Path)
	case "nfs":
		return nfsMount(ctx, dest.Host(), dest.Path, dest.Options)
	case "sftp":
		user := firstNonEmpty(dest.Access, os.Getenv("USER"), "backup")
		return NewSFTPStore(dest.Endpoint, dest.Port, user, dest.Secret, dest.Path)
	case "smb":
		return NewSMBStore(dest.Endpoint, dest.Share, dest.Access, dest.Secret, dest.Path, dest.Options)
	case "rclone":
		if !rcloneAvailable() {
			return nil, fmt.Errorf("rclone destination %q requires rclone binary in agent PATH", dest.Name)
		}
		return NewRcloneStore(dest.Remote, dest.Path)
	case "webdav":
		return NewWebDAVStore(dest.Endpoint, dest.Access, dest.Secret, dest.Path)
	default: // minio
		return NewObjectStore(a.minioConfigFor(dest))
	}
}

// destinationForSnapshot возвращает назначение снапшота (nil = legacy S3
// дефолта агента). Для не-S3 снапшотов запись destinations обязательна —
// учётки в снапшот не денормализуются.
func (a *Agent) destinationForSnapshot(ctx context.Context, cz *Corteza, snap *Snapshot) (*Destination, error) {
	kind := snap.DestType
	if kind == "" {
		kind = "minio"
	}
	if kind == "minio" {
		return nil, nil
	}
	if snap.DestID == 0 || cz == nil {
		return nil, fmt.Errorf("snapshot %d: destination record is required for %s operations", snap.ID, kind)
	}
	dest, err := a.loadDestination(ctx, cz, snap.DestID)
	if err != nil {
		return nil, err
	}
	if StoreKind(dest) != kind {
		return nil, fmt.Errorf("snapshot %d: destination %d is now type %q, snapshot needs %q", snap.ID, snap.DestID, StoreKind(dest), kind)
	}
	return dest, nil
}

// Host — для nfs в Endpoint может лежать "server:/export"; поле Host нет,
// поэтому аксессор с честным именем.
func (d *Destination) Host() string { return strings.TrimSpace(d.Endpoint) }

// storeCacheKey различает S3-клиентов по endpoint+учёткам, а не только по бакету.
func storeCacheKey(cfg MinioConfig) string {
	return strings.Join([]string{cfg.Endpoint, cfg.Access, cfg.Bucket, cfg.Region}, "|")
}

// cachedStore возвращает кэшированный S3-стор для конфигурации.
// ensure=true создаёт бакет при первом использовании (путь записи);
// ensure=false только строит клиента (пути чтения/удаления).
func (a *Agent) cachedStore(ctx context.Context, cfg MinioConfig, ensure bool) (Storage, error) {
	if a.store != nil && storeCacheKey(cfg) == storeCacheKey(a.cfg.Minio) {
		return a.store, nil
	}
	key := "minio|" + storeCacheKey(cfg)
	a.storesM.Lock()
	defer a.storesM.Unlock()
	if a.stores == nil {
		a.stores = map[string]Storage{}
	}
	if s, ok := a.stores[key]; ok {
		return s, nil
	}
	s, err := NewObjectStore(cfg)
	if err != nil {
		return nil, err
	}
	if ensure {
		if err := s.EnsureBucket(ctx); err != nil {
			return nil, fmt.Errorf("dest bucket %s: %w", cfg.Bucket, err)
		}
	}
	a.stores[key] = s
	return s, nil
}

// StoreForDestination возвращает стор назначения: nil = дефолт агента (env).
func (a *Agent) StoreForDestination(ctx context.Context, dest *Destination) (Storage, error) {
	if dest == nil {
		if a.store == nil {
			return nil, fmt.Errorf("minio is not configured")
		}
		if err := a.store.EnsureBucket(ctx); err != nil {
			return nil, err
		}
		return a.store, nil
	}
	key := destCacheKey(dest)
	a.storesM.Lock()
	if s, ok := a.stores[key]; ok {
		a.storesM.Unlock()
		return s, nil
	}
	a.storesM.Unlock()
	s, err := a.buildStorage(ctx, dest)
	if err != nil {
		return nil, err
	}
	if err := s.EnsureBucket(ctx); err != nil {
		if c, ok := s.(interface{ Close() error }); ok {
			_ = c.Close()
		}
		return nil, fmt.Errorf("destination %q: %w", dest.Name, err)
	}
	a.storesM.Lock()
	if a.stores == nil {
		a.stores = map[string]Storage{}
	}
	a.stores[key] = s
	a.storesM.Unlock()
	return s, nil
}

// storeForSnapshot выбирает стор для чтения/удаления существующего снапшота.
//   - снапшоты до появления назначений (без endpoint) — дефолтный S3-стор;
//   - не-S3 назначения переподключаются через запись destinations
//     (snap.DestID): учётки живут там и в снапшот не денормализуются;
//   - S3 на другом endpoint — клиент по endpoint+бакет с учётками агента.
func (a *Agent) storeForSnapshot(ctx context.Context, cz *Corteza, snap *Snapshot) (Storage, error) {
	if a.store == nil {
		return nil, fmt.Errorf("minio is not configured")
	}
	kind := snap.DestType
	if kind == "" {
		kind = "minio"
	}
	if kind != "minio" {
		if snap.DestID == 0 || cz == nil {
			return nil, fmt.Errorf("snapshot %d: destination record is required for %s restore (snapshot has no destination link)", snap.ID, kind)
		}
		dest, err := a.loadDestination(ctx, cz, snap.DestID)
		if err != nil {
			return nil, err
		}
		if StoreKind(dest) != kind {
			return nil, fmt.Errorf("snapshot %d: destination %d is now type %q, snapshot needs %q", snap.ID, snap.DestID, StoreKind(dest), kind)
		}
		return a.StoreForDestination(ctx, dest)
	}
	endpoint := strings.TrimSpace(snap.S3Endpoint)
	if endpoint == "" || endpoint == a.cfg.Minio.Endpoint {
		return a.store, nil
	}
	cfg := a.cfg.Minio
	cfg.Endpoint = endpoint
	if snap.S3Bucket != "" {
		cfg.Bucket = snap.S3Bucket
	}
	return a.cachedStore(ctx, cfg, false)
}

// cfgForStore — копия конфигурации агента с MinioConfig конкретного стора:
// для S3-сторов restic/kopia строят репозиторий на тех же endpoint/bucket.
// Для не-S3 сторов cfg.Minio обнуляется — движки получают repo-спеку отдельно
// (engineRepo), а не из этой конфигурации.
func (a *Agent) cfgForStore(store Storage) Config {
	cfg := a.cfg
	cfg.Minio = store.Config()
	return cfg
}

// DestTest проверяет назначение: хранилище доступно, корень пишется (для
// читающих бэкендов — читается), результат пишется в last_check* записи.
func (a *Agent) DestTest(ctx context.Context, jr JobRequest) (any, error) {
	cz := a.Corteza(jr)
	if cz == nil {
		return nil, fmt.Errorf("corteza is not configured")
	}
	id := ParseUint(jr.DestinationID)
	if id == 0 {
		return nil, fmt.Errorf("destinationID is required")
	}
	start := time.Now()
	dest, err := a.loadDestination(ctx, cz, id)
	now := time.Now().UTC().Format(time.RFC3339)
	status, errMsg := "ok", ""
	var result map[string]any
	if err != nil {
		status, errMsg = "failed", err.Error()
		result = map[string]any{"destinationID": fmtUint(id), "ok": false, "error": errMsg}
	} else {
		store, serr := a.StoreForDestination(ctx, dest)
		if serr != nil {
			status, errMsg = "failed", serr.Error()
			result = map[string]any{"destinationID": fmtUint(id), "ok": false, "error": errMsg}
		} else if perr := store.Ping(ctx); perr != nil {
			status, errMsg = "failed", perr.Error()
			result = map[string]any{"destinationID": fmtUint(id), "ok": false, "error": errMsg}
		} else {
			result = map[string]any{
				"destinationID": fmtUint(id),
				"ok":            true,
				"type":          StoreKind(dest),
				"location":      store.Location(),
				"bucket":        store.Bucket(),
				"durationMs":    time.Since(start).Milliseconds(),
			}
		}
	}
	if cz.HasToken() {
		vals := map[string]string{
			"last_check":        now,
			"last_check_status": status,
		}
		if errMsg != "" {
			vals["last_check_error"] = errMsg
		}
		if uerr := cz.UpdateValues(ctx, "destinations", id, vals); uerr != nil {
			return result, fmt.Errorf("dest-test: update record: %w", uerr)
		}
	}
	return result, nil
}
