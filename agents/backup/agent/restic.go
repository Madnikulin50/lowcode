package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func resticAvailable() bool {
	_, err := exec.LookPath("restic")
	return err == nil
}

// ResticRepo — готовая спека репозитория restic: URL + env + (опционально)
// временный SSH-ключ для sftp. Собирается из назначения (engineRepoFor) —
// restic умеет s3/local/sftp/rclone, поэтому спека зависит от типа.
type ResticRepo struct {
	Repo    string
	Env     []string
	KeyFile string
}

func (r *ResticRepo) close() {
	if r != nil && r.KeyFile != "" {
		_ = os.Remove(r.KeyFile)
	}
}

// resticPassword — пароль репозитория (restic допускает пустой).
func resticPassword(cfg Config, password string) string {
	return firstNonEmpty(password, cfg.ResticPassword, os.Getenv("RESTIC_PASSWORD"), os.Getenv("BACKUP_RESTIC_PASSWORD"))
}

// resticEnvBase — env для s3-репозитория (прежнее поведение).
func resticEnvBase(cfg Config, repo, password string) []string {
	env := append(os.Environ(),
		"RESTIC_REPOSITORY="+repo,
		"RESTIC_PASSWORD="+resticPassword(cfg, password),
		"AWS_ACCESS_KEY_ID="+cfg.Minio.Access,
		"AWS_SECRET_ACCESS_KEY="+cfg.Minio.Secret,
	)
	scheme := "http"
	if cfg.Minio.Secure {
		scheme = "https"
	}
	env = append(env, "AWS_S3_ENDPOINT="+scheme+"://"+cfg.Minio.Endpoint)
	return env
}

func s3ResticRepo(cfg Config, sourceHandle string) string {
	return fmt.Sprintf("s3:%s/%s/%s", cfg.Minio.Endpoint, cfg.Minio.Bucket, strings.TrimSuffix(ResticPrefix(sourceHandle), "/"))
}

// engineRepoFor собирает спеку restic-репозитория по назначению.
// dest == nil — дефолт агента (S3 из env). Ключ sftp сохраняется во временный
// файл и подставляется через -o sftp.args; подчистка — вызывающей стороной.
func (a *Agent) engineRepoFor(ctx context.Context, dest *Destination, handle string) (*ResticRepo, error) {
	cfg := a.cfg
	cfg.Minio = a.minioConfigFor(dest)
	pw := resticPassword(a.cfg, a.cfg.ResticPassword)
	switch StoreKind(dest) {
	case "minio":
		repo := s3ResticRepo(cfg, handle)
		return &ResticRepo{Repo: repo, Env: resticEnvBase(cfg, repo, pw)}, nil
	case "local", "nfs":
		root := strings.TrimSpace(dest.Path)
		if StoreKind(dest) == "nfs" {
			fs, err := nfsMount(ctx, dest.Host(), dest.Path, dest.Options)
			if err != nil {
				return nil, err
			}
			root = fs.root
		}
		repo := "local:" + filepath.Join(root, ResticPrefix(handle))
		return &ResticRepo{Repo: repo, Env: resticEnvBase(a.cfg, repo, pw)}, nil
	case "sftp":
		user := firstNonEmpty(dest.Access, os.Getenv("USER"), "backup")
		port := dest.Port
		if port <= 0 {
			port = 22
		}
		base := filepath.Join(dest.Path, ResticPrefix(handle))
		repo := fmt.Sprintf("sftp:%s@%s:%s", user, dest.Endpoint, base)
		spec := &ResticRepo{Repo: repo, Env: resticEnvBase(a.cfg, repo, pw)}
		secret := strings.TrimSpace(dest.Secret)
		if strings.HasPrefix(secret, "-----BEGIN") {
			keyFile := filepath.Join(os.TempDir(), "backup-dest-key-"+sanitizeKey(dest.Handle)+"-"+sanitizeKey(handle))
			if err := os.WriteFile(keyFile, []byte(secret), 0o600); err != nil {
				return nil, err
			}
			spec.KeyFile = keyFile
		} else if secret != "" {
			return nil, fmt.Errorf("restic sftp supports key-based auth only: put an SSH private key into BACKUP_SECRET_%s (password auth needs sshpass and is not supported)", dest.Handle)
		}
		return spec, nil
	case "rclone":
		if !rcloneAvailable() {
			return nil, fmt.Errorf("rclone destination requires rclone binary in agent PATH")
		}
		sub := filepath.Join(strings.Trim(dest.Path, "/"), ResticPrefix(handle))
		repo := fmt.Sprintf("rclone:%s:%s", dest.Remote, sub)
		return &ResticRepo{Repo: repo, Env: resticEnvBase(a.cfg, repo, pw)}, nil
	default:
		return nil, fmt.Errorf("engine backups (restic/kopia) to %s destinations are not supported: use engine=archive or a minio/local/nfs/sftp/rclone destination", StoreKind(dest))
	}
}

// args возвращает аргументы restic с учётом sftp-ключа.
func (r *ResticRepo) args(base ...string) []string {
	if r.KeyFile != "" {
		return append([]string{"-o", "sftp.args=-i " + r.KeyFile + " -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"}, base...)
	}
	return base
}

func resticInit(ctx context.Context, spec *ResticRepo, password string) error {
	cmd := exec.CommandContext(ctx, "restic", spec.args("init")...)
	cmd.Env = spec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.ToLower(stderr.String())
		if strings.Contains(msg, "already initialized") || strings.Contains(msg, "config file already") {
			return nil
		}
		return fmt.Errorf("restic init: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func resticBackupPath(ctx context.Context, spec *ResticRepo, localPath string, excludes []string, password string) (snapshotID string, err error) {
	if err := resticInit(ctx, spec, password); err != nil {
		return "", err
	}
	args := spec.args("backup", "--json")
	for _, ex := range excludes {
		args = append(args, "--exclude", ex)
	}
	args = append(args, localPath)
	cmd := exec.CommandContext(ctx, "restic", args...)
	cmd.Env = spec.Env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("restic backup: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	id := parseResticSnapshotID(stdout.Bytes())
	if id == "" {
		id = time.Now().UTC().Format("20060102T150405")
	}
	return id, nil
}

func resticRestore(ctx context.Context, spec *ResticRepo, snapshotID, dest, password string) error {
	if snapshotID == "" {
		snapshotID = "latest"
	}
	if dest == "" {
		return fmt.Errorf("restic restore dest is empty")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "restic", spec.args("restore", snapshotID, "--target", dest)...)
	cmd.Env = spec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic restore: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func resticForget(ctx context.Context, spec *ResticRepo, keepDays int, password string) error {
	if keepDays <= 0 {
		keepDays = 14
	}
	cmd := exec.CommandContext(ctx, "restic", spec.args("forget", "--prune", fmt.Sprintf("--keep-within=%dd", keepDays))...)
	cmd.Env = spec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic forget: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func parseResticSnapshotID(raw []byte) string {
	for _, line := range bytes.Split(raw, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var msg struct {
			MessageType string `json:"message_type"`
			SnapshotID  string `json:"snapshot_id"`
			ID          string `json:"id"`
		}
		if json.Unmarshal(line, &msg) != nil {
			continue
		}
		if msg.MessageType == "summary" || msg.SnapshotID != "" || msg.ID != "" {
			return firstNonEmpty(msg.SnapshotID, msg.ID)
		}
	}
	return ""
}

func resticKeyHint(sourceHandle, snapshotID string) string {
	return path.Join(strings.TrimSuffix(ResticPrefix(sourceHandle), "/"), snapshotID)
}

// resticCmd — exec.Cmd для read-only restic-команд в репозитории источника.
func resticCmd(ctx context.Context, spec *ResticRepo, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "restic", spec.args(args...)...)
	cmd.Env = spec.Env
	return cmd
}

// resticExtract восстанавливает поддерево: restic restore <id> --target --include.
func resticExtract(ctx context.Context, spec *ResticRepo, snapshotID, include, dest string) error {
	if snapshotID == "" {
		snapshotID = "latest"
	}
	cmd := exec.CommandContext(ctx, "restic", spec.args("restore", snapshotID, "--target", dest, "--include", include)...)
	cmd.Env = spec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic restore --include: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// resticCheck — read-only проверка репозитория (10% данных) для верификации L1.
func resticCheck(ctx context.Context, spec *ResticRepo) error {
	cmd := exec.CommandContext(ctx, "restic", spec.args("check", "--read-data-subset=10%")...)
	cmd.Env = spec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic check: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// --- Сейф (S3-only): репозитории внутри vault-бакета -------------------------

// resticVaultRepo — отдельный restic-репозиторий внутри vault-бакета.
// Блоки репозитория защищены default-retention бакета: restic пишет их сам,
// per-object retention агент выставить не может.
func resticVaultRepo(cfg Config, sourceHandle string) string {
	bucket := cfg.Vault.Bucket
	if bucket == "" {
		bucket = "vault"
	}
	return fmt.Sprintf("s3:%s/%s/%s", cfg.Minio.Endpoint, bucket, strings.TrimSuffix(ResticPrefix(sourceHandle), "/"))
}

// resticCopyVault копирует снапшот из рабочего репозитория в vault-репозиторий.
func resticCopyVault(ctx context.Context, cfg Config, sourceHandle, snapshotID, password string) error {
	to := resticVaultRepo(cfg, sourceHandle)
	toSpec := &ResticRepo{Repo: to, Env: resticEnvBase(cfg, to, password)}
	if err := resticInit(ctx, toSpec, password); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "restic", "copy",
		"--from-repo", s3ResticRepo(cfg, sourceHandle),
		"--from-password", resticPassword(cfg, password),
		snapshotID)
	cmd.Env = toSpec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic copy: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// resticRestoreFromVault — restore из vault-репозитория (чистая копия).
func resticRestoreFromVault(ctx context.Context, cfg Config, sourceHandle, snapshotID, dest, password string) error {
	repo := resticVaultRepo(cfg, sourceHandle)
	spec := &ResticRepo{Repo: repo, Env: resticEnvBase(cfg, repo, password)}
	if snapshotID == "" {
		snapshotID = "latest"
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "restic", spec.args("restore", snapshotID, "--target", dest)...)
	cmd.Env = spec.Env
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic restore from vault: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
