package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func resticAvailable() bool {
	_, err := exec.LookPath("restic")
	return err == nil
}

func resticEnv(cfg Config, repo string, password string) []string {
	pw := firstNonEmpty(password, cfg.ResticPassword, os.Getenv("RESTIC_PASSWORD"), os.Getenv("BACKUP_RESTIC_PASSWORD"))
	env := append(os.Environ(),
		"RESTIC_REPOSITORY="+repo,
		"RESTIC_PASSWORD="+pw,
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

func resticRepo(cfg Config, sourceHandle string) string {
	return fmt.Sprintf("s3:%s/%s/%s", cfg.Minio.Endpoint, cfg.Minio.Bucket, strings.TrimSuffix(ResticPrefix(sourceHandle), "/"))
}

func resticInit(ctx context.Context, cfg Config, repo, password string) error {
	cmd := exec.CommandContext(ctx, "restic", "init")
	cmd.Env = resticEnv(cfg, repo, password)
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

func resticBackupPath(ctx context.Context, cfg Config, sourceHandle, localPath, password string) (snapshotID string, err error) {
	repo := resticRepo(cfg, sourceHandle)
	if err := resticInit(ctx, cfg, repo, password); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "restic", "backup", "--json", localPath)
	cmd.Env = resticEnv(cfg, repo, password)
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

func resticRestore(ctx context.Context, cfg Config, sourceHandle, snapshotID, dest, password string) error {
	repo := resticRepo(cfg, sourceHandle)
	if snapshotID == "" {
		snapshotID = "latest"
	}
	if dest == "" {
		return fmt.Errorf("restic restore dest is empty")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "restic", "restore", snapshotID, "--target", dest)
	cmd.Env = resticEnv(cfg, repo, password)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restic restore: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func resticForget(ctx context.Context, cfg Config, sourceHandle string, keepDays int, password string) error {
	repo := resticRepo(cfg, sourceHandle)
	if keepDays <= 0 {
		keepDays = 14
	}
	cmd := exec.CommandContext(ctx, "restic", "forget", "--prune", fmt.Sprintf("--keep-within=%dd", keepDays))
	cmd.Env = resticEnv(cfg, repo, password)
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
