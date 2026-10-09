package agent

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type localFS struct {
	root string
}

func NewLocalFS(root string) (*localFS, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("empty filesystem path")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", abs)
	}
	return &localFS{root: abs}, nil
}

func (l *localFS) Walk(ctx context.Context, fn func(FileEntry, openFn) error) error {
	return filepath.WalkDir(l.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(l.root, path)
		if err != nil {
			return err
		}
		entry := FileEntry{
			RelPath: rel,
			Size:    info.Size(),
			Mode:    uint32(info.Mode()),
			ModTime: info.ModTime(),
		}
		if info.Mode()&os.ModeSymlink != 0 {
			// info.Size() here is the length of the link text (Lstat), not
			// the target's size, and os.Open below would follow the link
			// and read the (differently sized) target. Record it as a
			// symlink instead of copying content.
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			entry.LinkTarget = target
			entry.Size = 0
			return fn(entry, nil)
		}
		return fn(entry, func() (io.ReadCloser, error) {
			return os.Open(path)
		})
	})
}
