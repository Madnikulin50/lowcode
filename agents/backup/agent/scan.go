package agent

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strings"
	"time"
)

// Scanner — проверка содержимого снапшота перед отправкой в сейф.
// Status: clean — проверено, infected — найдена сигнатура,
// unknown — проверить не удалось (clamd недоступен, лимит выборки);
// unknown никогда не маскируется под clean.

type ScanResult struct {
	Status    ScanStatus
	Signature string
	Bytes     int64
	Detail    string
}

type Scanner interface {
	// Scan читает r до EOF и возвращает вердикт по потоку в целом.
	Scan(ctx context.Context, r io.Reader, name string) (*ScanResult, error)
	// MaxBytes — сколько байт максимум будет реально прочитано из потока.
	MaxBytes() int64
}

func (r *ScanResult) LogLine(name string) string {
	ts := time.Now().UTC().Format(time.RFC3339)
	switch r.Status {
	case ScanClean:
		return fmt.Sprintf("%s %s: OK (%d bytes)", ts, name, r.Bytes)
	case ScanInfected:
		return fmt.Sprintf("%s %s: INFECTED %s (%d bytes)", ts, name, r.Signature, r.Bytes)
	default:
		return fmt.Sprintf("%s %s: UNKNOWN (%s)", ts, name, r.Detail)
	}
}

// --- ClamAV (zINSTREAM) -------------------------------------------------

// ClamAVScanner шлёт поток в clamd по TCP (host:3310) или unix-сокету
// (unix:///путь). Протокол: "zINSTREAM" + чанки <длина:4 LE><data> + 0.
type ClamAVScanner struct {
	Addr    string
	MaxScan int64
	dial    func() (net.Conn, error)
}

func NewClamAVScanner(addr string, maxScan int64) *ClamAVScanner {
	c := &ClamAVScanner{Addr: strings.TrimSpace(addr), MaxScan: maxScan}
	c.dial = c.defaultDial
	return c
}

func (c *ClamAVScanner) MaxBytes() int64 {
	if c.MaxScan <= 0 {
		return math.MaxInt64
	}
	return c.MaxScan
}

func (c *ClamAVScanner) defaultDial() (net.Conn, error) {
	addr := c.Addr
	if strings.HasPrefix(addr, "unix://") {
		return net.DialTimeout("unix", strings.TrimPrefix(addr, "unix://"), 5*time.Second)
	}
	return net.DialTimeout("tcp", addr, 5*time.Second)
}

func (c *ClamAVScanner) Ping(ctx context.Context) error {
	conn, err := c.dial()
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}
	_ = conn.SetDeadline(deadline)
	if _, err := io.WriteString(conn, "zPING\000"); err != nil {
		return err
	}
	reply, err := readClamdReply(bufio.NewReader(conn))
	if err != nil {
		return err
	}
	if !strings.Contains(reply, "PONG") {
		return fmt.Errorf("clamd: unexpected ping reply %q", strings.TrimSpace(reply))
	}
	return nil
}

func (c *ClamAVScanner) Scan(ctx context.Context, r io.Reader, _ string) (*ScanResult, error) {
	conn, err := c.dial()
	if err != nil {
		return nil, fmt.Errorf("clamd dial %s: %w", c.Addr, err)
	}
	defer conn.Close()
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(30 * time.Minute)
	}
	_ = conn.SetDeadline(deadline)

	if _, err := io.WriteString(conn, "zINSTREAM\000"); err != nil {
		return nil, err
	}
	lr := &io.LimitedReader{R: r, N: c.MaxBytes()}
	chunk := make([]byte, 64*1024)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, readErr := lr.Read(chunk)
		if n > 0 {
			var lenBuf [4]byte
			binary.LittleEndian.PutUint32(lenBuf[:], uint32(n))
			if _, err := conn.Write(lenBuf[:]); err != nil {
				return nil, err
			}
			if _, err := conn.Write(chunk[:n]); err != nil {
				return nil, err
			}
			total += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}
	// Пустой чанк завершает поток.
	var zero [4]byte
	if _, err := conn.Write(zero[:]); err != nil {
		return nil, err
	}

	// Ответ clamd завершается NUL-байтом ("stream: OK\0").
	br := bufio.NewReader(conn)
	line, err := readClamdReply(br)
	if err != nil {
		return nil, fmt.Errorf("clamd reply: %w", err)
	}
	line = strings.TrimSpace(line)
	res := &ScanResult{Bytes: total}
	if strings.HasSuffix(line, ": OK") || line == "OK" {
		res.Status = ScanClean
		return res, nil
	}
	if strings.Contains(line, "FOUND") {
		res.Status = ScanInfected
		res.Signature = strings.TrimSpace(strings.TrimSuffix(strings.SplitN(line, ":", 2)[1], "FOUND"))
		return res, nil
	}
	res.Status = ScanUnknown
	res.Detail = "clamd: " + line
	return res, nil
}

// readClamdReply читает до NUL-терминатора (протокол clamd).
func readClamdReply(br *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		b, err := br.ReadByte()
		if err != nil {
			return sb.String(), err
		}
		if b == 0 {
			return sb.String(), nil
		}
		sb.WriteByte(b)
		if sb.Len() > 4096 {
			return sb.String(), fmt.Errorf("clamd reply too long")
		}
	}
}

// --- Эвристика (fallback без clamd) -------------------------------------

// HeuristicScanner оценивает энтропию скользящего окна первых MaxScan байт.
// Вердикт всегда unknown: без сигнатур честно сказать "чисто" нельзя.
type HeuristicScanner struct {
	MaxScan int64
}

func (h *HeuristicScanner) MaxBytes() int64 {
	if h.MaxScan <= 0 {
		return 4 << 20
	}
	return h.MaxScan
}

func (h *HeuristicScanner) Scan(_ context.Context, r io.Reader, _ string) (*ScanResult, error) {
	lr := &io.LimitedReader{R: r, N: h.MaxBytes()}
	sample := make([]byte, 0, 256*1024)
	buf := make([]byte, 32*1024)
	var total int64
	for {
		n, err := lr.Read(buf)
		total += int64(n)
		if len(sample) < cap(sample) {
			take := buf[:min(n, cap(sample)-len(sample))]
			sample = append(sample, take...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return &ScanResult{
		Status: ScanUnknown,
		Bytes:  total,
		Detail: fmt.Sprintf("clamd unavailable; sampled %d bytes, entropy=%.2f bits/byte", len(sample), entropy(sample)),
	}, nil
}

func entropy(b []byte) float64 {
	if len(b) == 0 {
		return 0
	}
	var freq [256]int
	for _, c := range b {
		freq[c]++
	}
	var e float64
	n := float64(len(b))
	for _, f := range freq {
		if f == 0 {
			continue
		}
		p := float64(f) / n
		e -= p * math.Log2(p)
	}
	return e
}

// --- Выбор сканера -------------------------------------------------------

// resolveScanner возвращает ClamAV, если он жив, иначе — эвристику.
func resolveScanner(addr string, maxMB int64) Scanner {
	maxScan := maxMB * (1 << 20)
	if maxMB <= 0 {
		maxScan = 2 << 30
	}
	if strings.TrimSpace(addr) != "" {
		cs := NewClamAVScanner(addr, maxScan)
		if err := cs.Ping(context.Background()); err != nil {
			log.Printf("scan: clamd %s unreachable (%v); heuristics only", addr, err)
		} else {
			return cs
		}
	}
	return &HeuristicScanner{MaxScan: 4 << 20}
}
