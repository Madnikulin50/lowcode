package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func DumpDatabase(ctx context.Context, src Source, username, password string, w io.Writer) (engine string, err error) {
	eng := strings.ToLower(strings.TrimSpace(src.DBEngine))
	if eng == "" {
		eng = "postgres"
	}
	switch eng {
	case "postgres", "postgresql", "pg":
		return "pg_dump", runPgDump(ctx, src, username, password, w)
	case "mysql", "mariadb":
		return "mysqldump", runMySQLDump(ctx, src, username, password, w)
	default:
		return "", fmt.Errorf("unsupported db engine %q", src.DBEngine)
	}
}

func runPgDump(ctx context.Context, src Source, username, password string, w io.Writer) error {
	bin, err := exec.LookPath("pg_dump")
	if err != nil {
		return fmt.Errorf("pg_dump not found in PATH")
	}
	host := firstNonEmpty(src.Host, "127.0.0.1")
	port := src.DBPort
	if port == 0 {
		port = 5432
	}
	db := firstNonEmpty(src.DBName, "postgres")
	user := firstNonEmpty(username, os.Getenv("PGUSER"), "postgres")
	args := []string{
		"-h", host,
		"-p", strconv.Itoa(port),
		"-U", user,
		"-d", db,
		"-Fc",
		"--no-password",
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
	cmd.Stdout = w
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func runMySQLDump(ctx context.Context, src Source, username, password string, w io.Writer) error {
	bin, err := exec.LookPath("mysqldump")
	if err != nil {
		return fmt.Errorf("mysqldump not found in PATH")
	}
	host := firstNonEmpty(src.Host, "127.0.0.1")
	port := src.DBPort
	if port == 0 {
		port = 3306
	}
	db := src.DBName
	if db == "" {
		return fmt.Errorf("database name is required")
	}
	user := firstNonEmpty(username, "root")
	args := []string{
		"-h", host,
		"-P", strconv.Itoa(port),
		"-u", user,
		"--single-transaction",
		"--routines",
		"--skip-comments",
		db,
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	if password != "" {
		cmd.Env = append(os.Environ(), "MYSQL_PWD="+password)
	}
	cmd.Stdout = w
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func RestoreDatabase(ctx context.Context, src Source, username, password string, r io.Reader, engine string) error {
	eng := strings.ToLower(engine)
	if strings.Contains(eng, "pg") || src.DBEngine == "postgres" || src.DBEngine == "postgresql" {
		return restorePg(ctx, src, username, password, r)
	}
	return restoreMySQL(ctx, src, username, password, r)
}

func restorePg(ctx context.Context, src Source, username, password string, r io.Reader) error {
	bin, err := exec.LookPath("pg_restore")
	if err != nil {
		return fmt.Errorf("pg_restore not found in PATH")
	}
	host := firstNonEmpty(src.Host, "127.0.0.1")
	port := src.DBPort
	if port == 0 {
		port = 5432
	}
	db := firstNonEmpty(src.DBName, "postgres")
	user := firstNonEmpty(username, os.Getenv("PGUSER"), "postgres")
	cmd := exec.CommandContext(ctx, bin,
		"-h", host, "-p", strconv.Itoa(port), "-U", user, "-d", db, "--no-password", "--clean", "--if-exists",
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
	cmd.Stdin = r
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_restore: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func restoreMySQL(ctx context.Context, src Source, username, password string, r io.Reader) error {
	bin, err := exec.LookPath("mysql")
	if err != nil {
		return fmt.Errorf("mysql client not found in PATH")
	}
	host := firstNonEmpty(src.Host, "127.0.0.1")
	port := src.DBPort
	if port == 0 {
		port = 3306
	}
	db := src.DBName
	if db == "" {
		return fmt.Errorf("database name is required")
	}
	user := firstNonEmpty(username, "root")
	cmd := exec.CommandContext(ctx, bin, "-h", host, "-P", strconv.Itoa(port), "-u", user, db)
	if password != "" {
		cmd.Env = append(os.Environ(), "MYSQL_PWD="+password)
	}
	cmd.Stdin = r
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}
