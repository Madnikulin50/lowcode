package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeClamd принимает zINSTREAM/zPING-соединения и отвечает verdict.
func fakeClamd(t *testing.T, verdict string) (addr string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
				br := bufio.NewReader(conn)
				for {
					line, err := br.ReadString(0)
					if err != nil {
						return
					}
					switch {
					case strings.HasPrefix(line, "zPING"):
						io.WriteString(conn, "PONG\000")
					case strings.HasPrefix(line, "zINSTREAM"):
						// Чанки <len:4><data> до нулевой длины.
						for {
							var lenBuf [4]byte
							if _, err := io.ReadFull(br, lenBuf[:]); err != nil {
								return
							}
							n := binary.LittleEndian.Uint32(lenBuf[:])
							if n == 0 {
								break
							}
							if _, err := io.ReadFull(br, make([]byte, n)); err != nil {
								return
							}
						}
						io.WriteString(conn, verdict+"\000")
					default:
						return
					}
				}
			}(conn)
		}
	}()
	return ln.Addr().String()
}

func TestClamAVScanCleanAndInfected(t *testing.T) {
	cases := []struct {
		verdict string
		want    ScanStatus
		sig     string
	}{
		{"stream: OK", ScanClean, ""},
		{"stream: Eicar-Test-Signature FOUND", ScanInfected, "Eicar-Test-Signature"},
	}
	for _, tc := range cases {
		addr := fakeClamd(t, tc.verdict)
		sc := NewClamAVScanner(addr, 1<<20)
		if err := sc.Ping(context.Background()); err != nil {
			t.Fatalf("ping: %v", err)
		}
		res, err := sc.Scan(context.Background(), strings.NewReader("hello world"), "x")
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		if res.Status != tc.want {
			t.Fatalf("%s: got %s want %s", tc.verdict, res.Status, tc.want)
		}
		if tc.sig != "" && !strings.Contains(res.Signature, tc.sig) {
			t.Fatalf("signature %q", res.Signature)
		}
	}
}

func TestHeuristicScannerAlwaysUnknown(t *testing.T) {
	res, err := (&HeuristicScanner{MaxScan: 1 << 10}).Scan(context.Background(), strings.NewReader(strings.Repeat("a", 1<<16)), "x")
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != ScanUnknown {
		t.Fatalf("heuristic must never claim clean: %s", res.Status)
	}
	if res.Detail == "" {
		t.Fatal("detail should explain")
	}
}

func TestScanTarGzWorstVerdict(t *testing.T) {
	// Пакуем два файла, "сканер" помечает второй infected.
	src := t.TempDir()
	writeFile := func(name, body string) {
		t.Helper()
		if err := writeTestFile(src, name, body); err != nil {
			t.Fatal(err)
		}
	}
	writeFile("a.txt", "clean content")
	writeFile("b.bin", "evil content")
	walker, err := NewLocalFS(src)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, _, err := PackTarGz(context.Background(), &buf, walker, nil); err != nil {
		t.Fatal(err)
	}

	sc := &scriptedScanner{verdicts: map[string]ScanStatus{"b.bin": ScanInfected}}
	res, err := scanTarGz(context.Background(), sc, bytes.NewReader(buf.Bytes()), &JobStatus{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != ScanInfected {
		t.Fatalf("worst verdict: %s", res.Status)
	}
	if sc.seen["a.txt"] != 1 || sc.seen["b.bin"] != 1 {
		t.Fatalf("seen: %v", sc.seen)
	}
}

type scriptedScanner struct {
	verdicts map[string]ScanStatus
	seen     map[string]int
}

func (s *scriptedScanner) MaxBytes() int64 { return 1 << 20 }
func (s *scriptedScanner) Scan(_ context.Context, r io.Reader, name string) (*ScanResult, error) {
	if s.seen == nil {
		s.seen = map[string]int{}
	}
	s.seen[name]++
	if _, err := io.Copy(io.Discard, r); err != nil {
		return nil, err
	}
	st := s.verdicts[name]
	if st == "" {
		st = ScanClean
	}
	res := &ScanResult{Status: st}
	if st == ScanInfected {
		res.Signature = "test-sig"
	}
	return res, nil
}

func writeTestFile(root, name, body string) error {
	return os.WriteFile(filepath.Join(root, name), []byte(body), 0o644)
}
