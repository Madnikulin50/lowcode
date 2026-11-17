package agent

import "time"

type SourceType string

const (
	SourceFS       SourceType = "fs"
	SourceSMB      SourceType = "smb"
	SourceDatabase SourceType = "database"
	SourceS3       SourceType = "s3"
	SourceSaaS     SourceType = "saas"
	SourceK8s      SourceType = "kubernetes"
)

type JobKind string

const (
	KindFull        JobKind = "full"
	KindIncremental JobKind = "incremental"
	KindRestore     JobKind = "restore"
	KindPrune       JobKind = "prune"
	KindVault       JobKind = "vault"
	KindScan        JobKind = "scan"
	KindVerify      JobKind = "verify"
	KindExtract     JobKind = "extract"
	KindManifest    JobKind = "manifest"
	KindDestroy     JobKind = "destroy"
)

type JobStatusName string

const (
	StatusPending   JobStatusName = "pending"
	StatusRunning   JobStatusName = "running"
	StatusCompleted JobStatusName = "completed"
	StatusFailed    JobStatusName = "failed"
)

type JobRequest struct {
	SourceID      string `json:"sourceID"`
	Source        string `json:"source"`
	PolicyID      string `json:"policyID"`
	JobRecordID   string `json:"jobID"`
	RestoreID     string `json:"restoreID"`
	SnapshotID    string `json:"snapshotID"`
	NamespaceID   string `json:"namespaceID"`
	Token         string `json:"token"`
	CallbackURL   string `json:"callbackUrl"`
	DestPath      string `json:"destPath"`
	DestType      string `json:"destType"`
	Kind          string `json:"kind"`
	RetentionDays int    `json:"retentionDays"`
	VerifyLevel   string `json:"verifyLevel"`
	Namespaces    string `json:"namespaces"`
	// ExtractTo — куда писать при backup/extract (пусто = вернуть содержимое).
	ExtractTo string `json:"extractTo"`
	// Basis — основание для деструктивных операций (уничтожение архива).
	Basis string `json:"basis"`
	// DestinationID — явное назначение (запись destinations); перекрывает
	// назначение политики. Пусто = резолвить из политики/дефолта.
	DestinationID string `json:"destinationID"`
	// ScopeID — аудит только этой области (audit_scopes). Пусто = все
	// включённые области. Явный запрос по ID выполняется даже для
	// выключенной области.
	ScopeID string `json:"scopeID"`
}

type JobStatus struct {
	ID            string        `json:"id"`
	Kind          string        `json:"kind"`
	Status        JobStatusName `json:"status"`
	Progress      float64       `json:"progress"`
	BytesRead     int64         `json:"bytesRead"`
	BytesWritten  int64         `json:"bytesWritten"`
	Files         int           `json:"files"`
	Error         string        `json:"error,omitempty"`
	Message       string        `json:"message,omitempty"`
	SourceID      string        `json:"sourceID,omitempty"`
	PolicyID      string        `json:"policyID,omitempty"`
	JobRecordID   string        `json:"jobRecordID,omitempty"`
	SnapshotID    string        `json:"snapshotID,omitempty"`
	S3Endpoint    string        `json:"s3Endpoint,omitempty"`
	S3Bucket      string        `json:"s3Bucket,omitempty"`
	S3Key         string        `json:"s3Key,omitempty"`
	Checksum      string        `json:"checksum,omitempty"`
	Engine        string        `json:"engine,omitempty"`
	ResticID      string        `json:"resticID,omitempty"`
	StartedAt     time.Time     `json:"startedAt"`
	FinishedAt    *time.Time    `json:"finishedAt,omitempty"`
	ExtractInline string        `json:"extractInline,omitempty"`
	CallbackURL   string        `json:"-"`
	Token         string        `json:"-"`
	NamespaceID   uint64        `json:"-"`
	// Summary — произвольная сводка результата (аудит, compliance); ключи
	// попадают в Envelope.Result без перекрытия стандартных полей.
	Summary map[string]any `json:"-"`
}

// Destination — явное назначение бэкапа (запись модуля destinations).
// Type определяет, какие поля используются:
//   - minio (default): Endpoint/Bucket/Region/Secure + credential (access/secret);
//   - local: Path — корневой каталог на агенте;
//   - nfs: Host (server:/export) + Path (mountpoint) + Options (mount opts);
//   - sftp: Host (+Port), Path (base dir), credential (user + пароль/PEM-ключ);
//   - smb: Host, Share, Path (subdir), credential, Options (JSON port/domain);
//   - rclone: Remote (имя из rclone-конфига агента) + Path (подкаталог);
//   - webdav: Endpoint (URL), Path (subdir), credential.
//
// Учётки: username из связанной записи credentials = access key / логин,
// секрет — BACKUP_SECRET_<handle> на агенте. Пустые Access/Secret означают
// «наследовать учётки агента» — так работают дефолтное назначение из env.
type Destination struct {
	ID       uint64
	Name     string
	Handle   string
	Type     string // minio | local | nfs | sftp | smb | rclone | webdav
	Endpoint string
	Port     int
	Bucket   string
	Prefix   string // префикс ключей по умолчанию (перекрывается dest_prefix политики)
	Path     string // корень local / mountpoint nfs / base dir sftp,webdav / subdir smb,rclone
	Share    string // smb
	Remote   string // rclone
	Options  string // nfs mount opts / smb JSON {"port","domain"}
	Region   string
	Secure   bool
	CredID   uint64
	Access   string // resolved: username из поля записи или credentials (пусто = анонимно/учётки агента)
	Secret   string // resolved: пароль/ключ из поля записи или BACKUP_SECRET_<handle>
	Username string // сырое поле записи destinations: логин прямо в форме
	Password string // сырое поле записи destinations: пароль/PEM-ключ прямо в форме (открытым текстом!)
	Default  bool   // назначение по умолчанию для политик без явного destination
	Enabled  bool
	Notes    string
}

// LegacyDest — bucket-only назначение для политик со старым полем
// dest_bucket: тот же MinIO агента, другой бакет.
func LegacyDest(bucket string) *Destination {
	return &Destination{Name: "legacy:" + bucket, Type: "minio", Bucket: bucket}
}

// IsRecord — true если назначение пришло из записи destinations
// (у legacy-синтезированных записей ID == 0).
func (d *Destination) IsRecord() bool { return d != nil && d.ID > 0 }

type Source struct {
	ID         uint64
	Name       string
	Type       SourceType
	AgentID    uint64
	CredID     uint64
	CredHandle string
	Path       string
	Host       string
	Share      string
	SMBPath    string
	DBEngine   string
	DBName     string
	DBPort     int
	S3Bucket   string
	S3Prefix   string
	S3Region   string
	S3Secure   bool
	// SaaS-источники: провайдер (bitrix24/imap/caldav), endpoint и аккаунт.
	Provider string
	BaseURL  string
	Account  string
	// K8s-источники: список namespace через запятую (пусто = все),
	// include_pv — выгружать манифесты PV/PVC.
	Namespaces string
	IncludePV  bool
	// DataClass — класс конфиденциальности (public/internal/confidential/
	// personal_special). Влияет на требования к движку и сейфу.
	DataClass string
	// LastSync — курсор инкрементальной SaaS-синхронизации.
	LastSync time.Time
	Enabled  bool
	Notes    string
}

// ScanStatus маркирует результат сканирования снапшота перед отправкой в сейф.
type ScanStatus string

const (
	ScanQueued   ScanStatus = "queued"
	ScanClean    ScanStatus = "clean"
	ScanInfected ScanStatus = "infected"
	ScanUnknown  ScanStatus = "unknown"
)

// DuePolicy is the read-only shape returned by Agent.DuePolicies — just
// enough for a caller (the "backup-run-due" rule chain) to start a backup
// via the normal record+backup/run path.
type DuePolicy struct {
	ID       string `json:"id"`
	SourceID string `json:"sourceID"`
	Name     string `json:"name"`
}

type Policy struct {
	ID            uint64
	Name          string
	SourceID      uint64
	Cron          string
	RetentionDays int
	Incremental   bool
	DestPrefix    string
	// DestinationID — запись destinations; 0 = legacy dest_bucket / дефолт / env.
	DestinationID uint64
	// Сейф: копировать завершённые бэкапы в WORM-хранилище со сканированием.
	Vault              bool
	VaultRetentionDays int
	VaultLegalHold     bool
	// Верификация: регулярно поднимать снапшоты в песочнице и проверять восстановление.
	Verify      bool
	VerifyCron  string
	VerifyLevel string // l1 (целостность) | l2 (+ подъём в docker)
	// Engine — движок бэкапа fs-источников: "" (auto) | archive | restic | kopia.
	// kopia: дедупликация + content-defined chunking + шифрование (fs only).
	Engine string
	// Exclude — паттерны исключений (glob, по строке/через запятую).
	Exclude string
	// DestBucket — переопределение бакета назначения (пусто = бакет агента).
	DestBucket string
	// Archive — юрзначимый архив: keep_years хранения, legal hold обязателен.
	Archive   bool
	KeepYears int
	Enabled   bool
	LastRun   time.Time
}

type Credential struct {
	ID       uint64
	Name     string
	Handle   string
	Kind     string
	Username string
	Extra    string
}

type Snapshot struct {
	ID         uint64
	JobID      uint64
	SourceID   uint64
	PolicyID   uint64
	DestID     uint64 // назначение (запись destinations), 0 = legacy/env
	DestType   string // minio | local | nfs | sftp | smb | rclone | webdav
	S3Endpoint string
	S3Bucket   string
	S3Key      string
	SizeBytes  int64
	Checksum   string
	Files      int
	Kind       string
	Engine     string
	ResticID   string
	ExpiresAt  time.Time
	Restorable bool
	// Сейф и сканирование.
	VaultKey         string
	VaultLockedUntil time.Time
	ScanStatus       ScanStatus
	ScannedAt        time.Time
	ScanLog          string
	// Верификация восстановления.
	VerifiedAt        time.Time
	VerifyDurationSec int64
	VerifyError       string
	// Юрзначимый архив: манифест с sha256 файлов.
	ManifestKey string
	ManifestSHA string
}

type BackupResult struct {
	Bucket       string
	Key          string
	BytesRead    int64
	BytesWritten int64
	Files        int
	Checksum     string
	Engine       string
	Kind         JobKind
	ResticID     string
}

type FileEntry struct {
	RelPath string
	Size    int64
	Mode    uint32
	ModTime time.Time
	// LinkTarget, when non-empty, marks this entry as a symlink: it is
	// written as a tar symlink record (no body) instead of having its
	// content copied. Without this, WalkDir's Lstat-based Size (the length
	// of the link text) never matches what os.Open (which follows the
	// link) actually reads, and archive/tar errors "write too long".
	LinkTarget string
}

type ingestEnvelope struct {
	JobID           string     `json:"jobID"`
	Kind            string     `json:"kind"`
	Status          string     `json:"status"`
	Progress        float64    `json:"progress"`
	BytesRead       int64      `json:"bytesRead"`
	BytesWritten    int64      `json:"bytesWritten"`
	Files           int        `json:"files"`
	Error           string     `json:"error,omitempty"`
	Message         string     `json:"message,omitempty"`
	SourceID        string     `json:"sourceID,omitempty"`
	PolicyID        string     `json:"policyID,omitempty"`
	S3Bucket        string     `json:"s3Bucket,omitempty"`
	S3Key           string     `json:"s3Key,omitempty"`
	Checksum        string     `json:"checksum,omitempty"`
	Engine          string     `json:"engine,omitempty"`
	ResticID        string     `json:"resticID,omitempty"`
	SnapshotID      string     `json:"snapshotID,omitempty"`
	NamespaceID     string     `json:"namespaceID,omitempty"`
	CreatedRecordID string     `json:"createdRecordID,omitempty"`
	StartedAt       time.Time  `json:"startedAt,omitempty"`
	FinishedAt      *time.Time `json:"finishedAt,omitempty"`
}
