package agent

import "time"

type SourceType string

const (
	SourceFS       SourceType = "fs"
	SourceSMB      SourceType = "smb"
	SourceDatabase SourceType = "database"
	SourceS3       SourceType = "s3"
)

type JobKind string

const (
	KindFull        JobKind = "full"
	KindIncremental JobKind = "incremental"
	KindRestore     JobKind = "restore"
	KindPrune       JobKind = "prune"
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
}

type JobStatus struct {
	ID           string        `json:"id"`
	Kind         string        `json:"kind"`
	Status       JobStatusName `json:"status"`
	Progress     float64       `json:"progress"`
	BytesRead    int64         `json:"bytesRead"`
	BytesWritten int64         `json:"bytesWritten"`
	Files        int           `json:"files"`
	Error        string        `json:"error,omitempty"`
	Message      string        `json:"message,omitempty"`
	SourceID     string        `json:"sourceID,omitempty"`
	PolicyID     string        `json:"policyID,omitempty"`
	JobRecordID  string        `json:"jobRecordID,omitempty"`
	SnapshotID   string        `json:"snapshotID,omitempty"`
	S3Bucket     string        `json:"s3Bucket,omitempty"`
	S3Key        string        `json:"s3Key,omitempty"`
	Checksum     string        `json:"checksum,omitempty"`
	Engine       string        `json:"engine,omitempty"`
	ResticID     string        `json:"resticID,omitempty"`
	StartedAt    time.Time     `json:"startedAt"`
	FinishedAt   *time.Time    `json:"finishedAt,omitempty"`
	CallbackURL  string        `json:"-"`
	Token        string        `json:"-"`
	NamespaceID  uint64        `json:"-"`
}

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
	Enabled    bool
	Notes      string
}

type Policy struct {
	ID            uint64
	Name          string
	SourceID      uint64
	Cron          string
	RetentionDays int
	Incremental   bool
	DestPrefix    string
	Enabled       bool
	LastRun       time.Time
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
