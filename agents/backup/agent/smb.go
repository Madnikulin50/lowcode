package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

type smbFS struct {
	session *smb2.Session
	share   *smb2.Share
	root    string
}

type smbExtra struct {
	Domain string `json:"domain"`
	Port   int    `json:"port"`
}

func ParseUNC(unc, host, share, smbPath string) (string, string, string, error) {
	host = strings.TrimSpace(host)
	share = strings.TrimSpace(share)
	smbPath = strings.Trim(strings.ReplaceAll(smbPath, `\`, "/"), "/")
	if host != "" && share != "" {
		return host, share, smbPath, nil
	}
	raw := strings.TrimSpace(unc)
	if raw == "" {
		return "", "", "", fmt.Errorf("smb host and share are required")
	}
	if strings.HasPrefix(strings.ToLower(raw), "smb://") {
		raw = raw[6:]
	}
	raw = strings.ReplaceAll(raw, `/`, `\`)
	raw = strings.TrimPrefix(raw, `\\`)
	raw = strings.TrimPrefix(raw, `\`)
	parts := strings.SplitN(raw, `\`, 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", fmt.Errorf("invalid UNC %q", unc)
	}
	host = parts[0]
	share = parts[1]
	if len(parts) == 3 && smbPath == "" {
		smbPath = strings.ReplaceAll(parts[2], `\`, "/")
	}
	return host, share, strings.Trim(smbPath, "/"), nil
}

func DialSMB(host, username, password, extraJSON, share, root string) (*smbFS, error) {
	var extra smbExtra
	_ = json.Unmarshal([]byte(extraJSON), &extra)
	port := extra.Port
	if port == 0 {
		port = 445
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return nil, err
	}
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
			Domain:   extra.Domain,
		},
	}
	session, err := d.Dial(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	sh, err := session.Mount(share)
	if err != nil {
		session.Logoff()
		return nil, err
	}
	return &smbFS{session: session, share: sh, root: strings.Trim(strings.ReplaceAll(root, `\`, "/"), "/")}, nil
}

func (s *smbFS) Close() {
	if s.share != nil {
		_ = s.share.Umount()
	}
	if s.session != nil {
		_ = s.session.Logoff()
	}
}

func (s *smbFS) Walk(ctx context.Context, fn func(FileEntry, openFn) error) error {
	return s.walkDir(ctx, s.root, fn)
}

func (s *smbFS) walkDir(ctx context.Context, dir string, fn func(FileEntry, openFn) error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	entries, err := s.share.ReadDir(smbPath(dir))
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == "." || name == ".." {
			continue
		}
		child := name
		if dir != "" {
			child = path.Join(dir, name)
		}
		if e.IsDir() {
			if err := s.walkDir(ctx, child, fn); err != nil {
				return err
			}
			continue
		}
		rel := child
		if s.root != "" {
			rel = strings.TrimPrefix(child, s.root+"/")
		}
		entry := FileEntry{
			RelPath: rel,
			Size:    e.Size(),
			Mode:    uint32(e.Mode()),
			ModTime: e.ModTime(),
		}
		p := smbPath(child)
		if err := fn(entry, func() (io.ReadCloser, error) {
			return s.share.OpenFile(p, os.O_RDONLY, 0)
		}); err != nil {
			return err
		}
	}
	return nil
}

func smbPath(p string) string {
	p = strings.ReplaceAll(p, "/", `\`)
	if p == "" {
		return `.`
	}
	return p
}
