package service

import (
	"testing"

	"github.com/gabriel-vasile/mimetype"
)

func TestCheckMimeTypeFilenameFallback(t *testing.T) {
	var svc attachment
	detected := mimetype.Detect([]byte("this is not a cad file"))

	if !svc.checkMimeType(detected, "plan.pln", ".pln", "application/pdf") {
		t.Fatal("filename .pln should match allowlist .pln even when sniffed mime is unrelated")
	}
	if svc.checkMimeType(detected, "plan.pln", "application/pdf") {
		t.Fatal("filename .pln should not match a PDF-only allowlist")
	}
	if !svc.checkMimeType(detected, "drawing.DXF", ".dxf") {
		t.Fatal("filename extension match should be case-insensitive")
	}
}
