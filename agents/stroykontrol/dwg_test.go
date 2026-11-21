package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIsDWG(t *testing.T) {
	for _, tc := range []struct {
		name, mt string
		data     []byte
		want     bool
	}{
		{"R2018 magic", "", []byte("AC1032\x00\x00\x00"), true},
		{"R2000 magic", "application/octet-stream", []byte("AC1015\x00"), true},
		{"mimetype only", "image/vnd.dwg", nil, true},
		{"pdf", "application/pdf", []byte("%PDF-1.7"), false},
		{"ascii dxf", "", []byte("  0\nSECTION\n"), false},
		{"AC10 without version", "", []byte("AC10xx"), false},
	} {
		if got := IsDWG(tc.mt, tc.data); got != tc.want {
			t.Errorf("%s: IsDWG = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func fakeConverter(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "fake-dwg2pdf.sh")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestNormalizeDocConvertsAndCaches(t *testing.T) {
	counter := filepath.Join(t.TempDir(), "calls")
	script := fakeConverter(t, `echo x >> "`+counter+`"; printf '%%PDF-1.4 converted' > "$2"`)
	old := dwgConv
	defer func() { dwgConv = old }()
	dwgConv = NewDWGConverter(script+" {in} {out}", t.TempDir())

	dwg := []byte("AC1032 drawing")
	for i := 0; i < 2; i++ {
		data, mt := normalizeDoc(dwg, "image/vnd.dwg")
		if mt != "application/pdf" || !bytes.HasPrefix(data, []byte("%PDF")) {
			t.Fatalf("call %d: got %q %q", i, mt, data)
		}
	}
	calls, _ := os.ReadFile(counter)
	if n := bytes.Count(calls, []byte("x")); n != 1 {
		t.Errorf("converter ran %d times, want 1 (second call must hit the cache)", n)
	}
}

func TestNormalizeDocFallsBackOnFailure(t *testing.T) {
	old := dwgConv
	defer func() { dwgConv = old }()
	dwgConv = NewDWGConverter(fakeConverter(t, "exit 1")+" {in} {out}", t.TempDir())

	dwg := []byte("AC1032 broken")
	data, mt := normalizeDoc(dwg, "image/vnd.dwg")
	if !bytes.Equal(data, dwg) || mt != "image/vnd.dwg" {
		t.Errorf("failed conversion must return the original, got %q %q", data, mt)
	}
}

func TestNormalizeDocPassThrough(t *testing.T) {
	old := dwgConv
	defer func() { dwgConv = old }()
	dwgConv = nil
	dwg := []byte("AC1032")
	if data, _ := normalizeDoc(dwg, ""); !bytes.Equal(data, dwg) {
		t.Error("disabled converter must pass DWG through")
	}
	pdf := []byte("%PDF-1.4")
	dwgConv = NewDWGConverter("false {in} {out}", t.TempDir())
	if data, mt := normalizeDoc(pdf, "application/pdf"); !bytes.Equal(data, pdf) || mt != "application/pdf" {
		t.Error("PDF must never reach the converter")
	}
}
