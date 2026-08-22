package agent

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type openFn func() (io.ReadCloser, error)

type FileWalker interface {
	Walk(ctx context.Context, fn func(FileEntry, openFn) error) error
}

func PackTarGz(ctx context.Context, w io.Writer, walker FileWalker, onFile func(FileEntry, int64)) (files int, bytesRead int64, err error) {
	gz := gzip.NewWriter(w)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	err = walker.Walk(ctx, func(entry FileEntry, open openFn) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		hdr := &tar.Header{
			Name:    strings.TrimPrefix(filepath.ToSlash(entry.RelPath), "/"),
			Mode:    int64(entry.Mode),
			Size:    entry.Size,
			ModTime: entry.ModTime,
		}
		if hdr.Mode == 0 {
			hdr.Mode = 0644
		}
		if hdr.ModTime.IsZero() {
			hdr.ModTime = time.Now()
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		rc, err := open()
		if err != nil {
			return err
		}
		n, err := io.Copy(tw, rc)
		rc.Close()
		if err != nil {
			return err
		}
		files++
		bytesRead += n
		if onFile != nil {
			onFile(entry, n)
		}
		return nil
	})
	return files, bytesRead, err
}

func UnpackTarGz(ctx context.Context, r io.Reader, dest string) (files int, err error) {
	if dest == "" {
		return 0, fmt.Errorf("empty restore path")
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return 0, err
	}
	gz, err := gzip.NewReader(r)
	if err != nil {
		return 0, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		if ctx.Err() != nil {
			return files, ctx.Err()
		}
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return files, err
		}
		target, err := safeJoin(dest, hdr.Name)
		if err != nil {
			return files, err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return files, err
			}
		default:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return files, err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(hdr.Mode))
			if err != nil {
				return files, err
			}
			_, err = io.Copy(f, tr)
			f.Close()
			if err != nil {
				return files, err
			}
			files++
		}
	}
	return files, nil
}

func safeJoin(root, name string) (string, error) {
	root = filepath.Clean(root)
	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(name)))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("refusing path %q outside %s", name, root)
	}
	return target, nil
}
