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

// Kopia-движок: дедуплицированные инкрементальные снапшоты fs-источников.
// Репозиторий kopia (content-defined chunking, шифрование AES-256/ChaCha20)
// живёт в том же MinIO: s3://<bucket>/kopia/<handle>. Каждый снапшот —
// инкремент по содержимому; повторяющийся контент разных источников
// хранится один раз на весь репозиторий.
//
// Отличия от restic-движка, важные для остального агента:
//   - блобы зашифрованы и порезаны на чанки → ClamAV их не сканирует
//     (scan для kopia-снапшотов помечается unknown, копия в сейф не делается);
//   - retention движется через `kopia snapshot expire` + maintenance,
//     а не через удаление объектов (runPrune вызывает kopiaExpire);
//   - верификация L1 — `kopia snapshot verify`.

func kopiaAvailable() bool {
	_, err := exec.LookPath("kopia")
	return err == nil
}

// kopiaPassword — пароль репозитория. У kopia, в отличие от restic, пустого
// пароля не бывает: пустое значение приводит к интерактивному промпту,
// который вне TTY падает с "inappropriate ioctl for device".
func kopiaPassword(cfg Config, password string) (string, error) {
	pw := firstNonEmpty(password, cfg.ResticPassword, os.Getenv("KOPIA_PASSWORD"), os.Getenv("RESTIC_PASSWORD"), os.Getenv("BACKUP_RESTIC_PASSWORD"))
	if pw == "" {
		return "", fmt.Errorf("kopia repository password is empty: set --restic-password or RESTIC_PASSWORD env (kopia does not support empty passwords, unlike restic)")
	}
	return pw, nil
}

// kopiaEnv — окружение для всех вызовов kopia.
func kopiaEnv(cfg Config, password string) []string {
	pw, err := kopiaPassword(cfg, password)
	if err != nil {
		pw = ""
	}
	return append(os.Environ(),
		"KOPIA_PASSWORD="+pw,
		"KOPIA_CHECK_FOR_UPDATES=false",
		"KOPIA_UPGRADE_LOCK_ENABLED=false",
	)
}

// kopiaGlobalArgs — глобальные флаги: config-файл и пароль (явный флаг
// надёжнее env: пустой env kopia молча игнорирует).
func kopiaGlobalArgs(cfg Config, cfgPath, password string) ([]string, error) {
	pw, err := kopiaPassword(cfg, password)
	if err != nil {
		return nil, err
	}
	return []string{"--config-file=" + cfgPath, "--password=" + pw}, nil
}

// kopiaConfigPath — per-handle файл подключения. Содержимое не секретно
// (эндпоинт и префикс), поэтому жить может в tmp; ensureConnect пересоздаёт
// его при каждом запуске, так что потеря безвредна.
func kopiaConfigPath(handle string) string {
	return path.Join(os.TempDir(), "backup-kopia-"+sanitizeKey(handle)+".config")
}

// kopiaConnectArgs — общие аргументы подключения к репозиторию по типу
// назначения. kopia нативно умеет s3/filesystem/sftp; для остальных типов —
// честная ошибка (archive engine).
func kopiaConnectArgs(ctx context.Context, cfg Config, dest *Destination, handle string, create bool) ([]string, error) {
	switch StoreKind(dest) {
	case "local", "nfs":
		root := strings.TrimSpace(dest.Path)
		if StoreKind(dest) == "nfs" {
			fs, err := nfsMount(ctx, dest.Host(), dest.Path, dest.Options)
			if err != nil {
				return nil, err
			}
			root = fs.root
		}
		return []string{"filesystem", "--path=" + filepath.Join(root, kopiaPrefix(handle))}, nil
	case "sftp":
		user := firstNonEmpty(dest.Access, os.Getenv("USER"), "backup")
		port := dest.Port
		if port <= 0 {
			port = 22
		}
		base := filepath.Join(dest.Path, kopiaPrefix(handle))
		args := []string{"sftp",
			"--host=" + dest.Endpoint,
			fmt.Sprintf("--port=%d", port),
			"--username=" + user,
			"--path=" + base,
		}
		secret := strings.TrimSpace(dest.Secret)
		if strings.HasPrefix(secret, "-----BEGIN") {
			keyFile := filepath.Join(os.TempDir(), "backup-kopia-key-"+sanitizeKey(dest.Handle)+"-"+sanitizeKey(handle))
			if err := os.WriteFile(keyFile, []byte(secret), 0o600); err != nil {
				return nil, err
			}
			args = append(args, "--key-file="+keyFile)
		} else if secret != "" {
			return nil, fmt.Errorf("kopia sftp supports key-based auth only: put an SSH private key into BACKUP_SECRET_%s", dest.Handle)
		}
		return args, nil
	default: // minio
		scheme := "https"
		if !cfg.Minio.Secure {
			scheme = "http"
		}
		endpoint := cfg.Minio.Endpoint
		// kopia ждёт endpoint без схемы для флага --endpoint.
		endpoint = strings.TrimPrefix(endpoint, scheme+"://")
		args := []string{"s3",
			"--bucket=" + cfg.Minio.Bucket,
			"--prefix=" + kopiaPrefix(handle),
			"--endpoint=" + endpoint,
			"--access-key=" + cfg.Minio.Access,
			"--secret-access-key=" + cfg.Minio.Secret,
		}
		if cfg.Minio.Region != "" {
			args = append(args, "--region="+cfg.Minio.Region)
		}
		if !cfg.Minio.Secure {
			args = append(args, "--disable-tls")
		}
		return args, nil
	}
}

func kopiaPrefix(handle string) string {
	return strings.TrimSuffix(path.Join("kopia", sanitizeKey(handle)), "/") + "/"
}

// kopiaConnect создаёт репозиторий, если его ещё нет, и пишет config-файл.
func kopiaConnect(ctx context.Context, cfg Config, dest *Destination, handle, password string) error {
	cfgPath := kopiaConfigPath(handle)
	if _, err := os.Stat(cfgPath); err == nil {
		return nil
	}
	gargs, err := kopiaGlobalArgs(cfg, cfgPath, password)
	if err != nil {
		return err
	}
	connArgs, err := kopiaConnectArgs(ctx, cfg, dest, handle, false)
	if err != nil {
		return err
	}
	connect := append([]string{"repository", "connect"}, connArgs...)
	cmd := exec.CommandContext(ctx, "kopia", append(gargs, connect...)...)
	cmd.Env = kopiaEnv(cfg, password)
	if out, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(out), "repository not initialized") &&
			!strings.Contains(string(out), "not found") {
			// connect провалился по неизвестной причине — пробуем create.
			_ = out
		}
		create := append([]string{"repository", "create"}, connArgs...)
		ccmd := exec.CommandContext(ctx, "kopia", append(gargs, create...)...)
		ccmd.Env = kopiaEnv(cfg, password)
		cout, cerr := ccmd.CombinedOutput()
		if cerr != nil {
			return fmt.Errorf("kopia repository create: %w: %s", cerr, strings.TrimSpace(string(cout)))
		}
		return nil
	}
	return nil
}

// kopiaSnapshotCreate делает снапшот path и возвращает manifest ID.
func kopiaSnapshotCreate(ctx context.Context, cfg Config, dest *Destination, handle, snapshotPath string, excludes []string, password string) (string, error) {
	if err := kopiaConnect(ctx, cfg, dest, handle, password); err != nil {
		return "", err
	}
	cfgPath := kopiaConfigPath(handle)
	gargs, err := kopiaGlobalArgs(cfg, cfgPath, password)
	if err != nil {
		return "", err
	}
	args := append(gargs, "snapshot", "create", snapshotPath, "--no-progress")
	for _, ex := range excludes {
		args = append(args, "--exclude="+ex)
	}
	cmd := exec.CommandContext(ctx, "kopia", args...)
	cmd.Env = kopiaEnv(cfg, password)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if out, err := cmd.Output(); err != nil {
		return "", fmt.Errorf("kopia snapshot create: %w: %s", err, strings.TrimSpace(stderr.String()))
	} else {
		_ = out
	}
	return kopiaLatestSnapshotID(ctx, cfg, handle, password)
}

// kopiaLatestSnapshotID — manifest ID последнего снапшота (один источник
// на репозиторий, поэтому «последний» — это наш). Флага --latest у
// `kopia snapshot list` нет (v0.19) — берём максимум по времени из JSON.
func kopiaLatestSnapshotID(ctx context.Context, cfg Config, handle, password string) (string, error) {
	cfgPath := kopiaConfigPath(handle)
	gargs, err := kopiaGlobalArgs(cfg, cfgPath, password)
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "kopia", append(gargs, "snapshot", "list", "--json")...)
	cmd.Env = kopiaEnv(cfg, password)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("kopia snapshot list: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return parseKopiaSnapshotID(out), nil
}

func parseKopiaSnapshotID(raw []byte) string {
	var entries []struct {
		ID               string    `json:"id"`
		StartTime        time.Time `json:"startTime"`
		ModificationTime time.Time `json:"modTime"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(raw), &entries); err != nil {
		return ""
	}
	latest := ""
	var latestAt time.Time
	for _, e := range entries {
		if e.ID == "" {
			continue
		}
		at := e.StartTime
		if at.IsZero() {
			at = e.ModificationTime
		}
		if latest == "" || at.After(latestAt) {
			latest, latestAt = e.ID, at
		}
	}
	return latest
}

// kopiaSnapshotRestore распаковывает снапшот в dest; include (опционально)
// ограничивает восстановление поддеревом внутри снапшота.
func kopiaSnapshotRestore(ctx context.Context, cfg Config, d *Destination, handle, snapshotID, dest, include, password string) error {
	if snapshotID == "" {
		snapshotID = "latest"
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	if err := kopiaConnect(ctx, cfg, d, handle, password); err != nil {
		return err
	}
	cfgPath := kopiaConfigPath(handle)
	gargs, err := kopiaGlobalArgs(cfg, cfgPath, password)
	if err != nil {
		return err
	}
	args := append(gargs, "snapshot", "restore", snapshotID, dest)
	if include != "" {
		args = append(args, "--include="+include)
	}
	cmd := exec.CommandContext(ctx, "kopia", args...)
	cmd.Env = kopiaEnv(cfg, password)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kopia restore: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// kopiaSnapshotVerify — read-only проверка репозитория (верификация L1).
func kopiaSnapshotVerify(ctx context.Context, cfg Config, d *Destination, handle, password string) error {
	if err := kopiaConnect(ctx, cfg, d, handle, password); err != nil {
		return err
	}
	cfgPath := kopiaConfigPath(handle)
	gargs, err := kopiaGlobalArgs(cfg, cfgPath, password)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "kopia", append(gargs, "snapshot", "verify")...)
	cmd.Env = kopiaEnv(cfg, password)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kopia snapshot verify: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// kopiaExpire удаляет снапшоты старше keepDays и запускает maintenance,
// чтобы освободить место (неиспользуемые чанки удаляет full maintenance).
func kopiaExpire(ctx context.Context, cfg Config, d *Destination, handle string, keepDays int, password string) error {
	if keepDays <= 0 {
		keepDays = 14
	}
	cutoff := time.Now().Add(-time.Duration(keepDays) * 24 * time.Hour).Format("2006-01-02")
	if err := kopiaConnect(ctx, cfg, d, handle, password); err != nil {
		return err
	}
	cfgPath := kopiaConfigPath(handle)
	gargs, err := kopiaGlobalArgs(cfg, cfgPath, password)
	if err != nil {
		return err
	}
	expire := exec.CommandContext(ctx, "kopia",
		append(gargs, "snapshot", "expire", "--expire-before="+cutoff)...)
	expire.Env = kopiaEnv(cfg, password)
	if out, err := expire.CombinedOutput(); err != nil {
		return fmt.Errorf("kopia snapshot expire: %w: %s", err, strings.TrimSpace(string(out)))
	}
	maint := exec.CommandContext(ctx, "kopia",
		append(gargs, "maintenance", "run", "--full")...)
	maint.Env = kopiaEnv(cfg, password)
	if out, err := maint.CombinedOutput(); err != nil {
		// Maintenance не критичен: место освободится при следующем запуске.
		_ = out
	}
	return nil
}

func kopiaKeyHint(handle, snapshotID string) string {
	return path.Join(kopiaPrefix(handle), "manifests", snapshotID)
}

// kopiaCLI — exec.Cmd для команд в уже подключённом репозитории источника.
// Ошибка пароля здесь не проверяется: browse-обёртки разбирают вывод сами.
func kopiaCLI(ctx context.Context, cfg Config, cfgPath string, args ...string) *exec.Cmd {
	full := append([]string{"--config-file=" + cfgPath}, args...)
	cmd := exec.CommandContext(ctx, "kopia", full...)
	cmd.Env = kopiaEnv(cfg, cfg.ResticPassword)
	return cmd
}
