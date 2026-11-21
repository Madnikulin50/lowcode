package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/rag"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

const maxRuleChainExtractBytes = 50 << 20
const defaultExtractMaxChars = rag.DefaultExtractMaxChars

// RuleChainAttachmentLoader reads compose record attachments for document.extract.
type RuleChainAttachmentLoader struct{}

func (RuleChainAttachmentLoader) Load(ctx context.Context, namespaceID, attachmentID uint64) (*rulesgo.AttachmentBytes, error) {
	if DefaultAttachment == nil {
		return nil, fmt.Errorf("attachment service not initialized")
	}
	if attachmentID == 0 {
		return nil, fmt.Errorf("attachmentID is required")
	}
	att, err := DefaultAttachment.FindByID(ctx, namespaceID, attachmentID)
	if err != nil {
		return nil, err
	}
	if att == nil {
		return nil, fmt.Errorf("attachment %d not found", attachmentID)
	}
	rc, err := DefaultAttachment.OpenOriginal(att)
	if err != nil {
		return nil, err
	}
	if rc == nil {
		return nil, fmt.Errorf("attachment %d has no original", attachmentID)
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, maxRuleChainExtractBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxRuleChainExtractBytes {
		return nil, fmt.Errorf("attachment too large")
	}
	return &rulesgo.AttachmentBytes{
		ID:   att.ID,
		Name: att.Name,
		MIME: att.Meta.Original.Mimetype,
		Data: data,
	}, nil
}

var _ rulesgo.AttachmentLoader = RuleChainAttachmentLoader{}

type extractConfig struct {
	AttachmentField string `json:"attachmentField"`
	MaxChars        int    `json:"maxChars"`
	OutText         string `json:"outText"`
}

type documentExtractExecutor struct {
	load rulesgo.AttachmentLoader
}

// NewDocumentExtractExecutor is the rulesgo document.extract node.
func NewDocumentExtractExecutor() rulesgo.NodeExecutor {
	return &documentExtractExecutor{load: RuleChainAttachmentLoader{}}
}

func (n *documentExtractExecutor) Execute(ctx context.Context, node rulesgo.ChainNode, ec *rulesgo.ExecutionContext) (map[string]interface{}, error) {
	cfg, err := rulesgo.ParseNodeConfig[extractConfig](node.Config)
	if err != nil {
		return nil, err
	}
	field := strings.TrimSpace(cfg.AttachmentField)
	if field == "" {
		field = "file"
	}
	outKey := strings.TrimSpace(cfg.OutText)
	if outKey == "" {
		outKey = "extracted_text"
	}
	maxChars := cfg.MaxChars
	if maxChars <= 0 {
		maxChars = defaultExtractMaxChars
	}

	fail := func(kind, msg string) map[string]interface{} {
		status := "failed"
		if kind == "unsupported" {
			status = "unsupported"
		}
		out := map[string]interface{}{
			"extract_ok":      false,
			"extract_status":  status,
			"extract_kind":    kind,
			"extract_error":   msg,
			"extract_partial": false,
			"needs_ocr":       false,
			outKey:            "",
		}
		promoteExtract(ec, out, outKey)
		return out
	}

	if n.load == nil {
		return fail("unsupported", "attachment loader not configured"), nil
	}

	nsID := extractUint64(ec.Get("namespaceID"))
	ids := rulesgo.AttachmentIDsFromValue(ec.Get(field))
	if len(ids) == 0 {
		ids = rulesgo.AttachmentIDsFromValue(ec.Get("attachmentID"))
	}
	if len(ids) == 0 {
		return fail("unsupported", "no file attachment on the record"), nil
	}

	loaded := make([]*rulesgo.AttachmentBytes, 0, len(ids))
	for _, id := range ids {
		att, err := n.load.Load(ctx, nsID, id)
		if err != nil {
			return fail("failed", err.Error()), nil
		}
		if att == nil || len(att.Data) == 0 {
			continue
		}
		if int64(len(att.Data)) > maxRuleChainExtractBytes {
			return fail("failed", fmt.Sprintf("file %s exceeds %d bytes", att.Name, maxRuleChainExtractBytes)), nil
		}
		loaded = append(loaded, att)
	}
	if len(loaded) == 0 {
		return fail("unsupported", "attachments empty"), nil
	}

	chosen := pickExtractAttachment(loaded)
	parsed, err := rag.ParseDocument(chosen.Data, chosen.Name, chosen.MIME)
	if err != nil {
		return fail("failed", err.Error()), nil
	}

	text := rag.TruncateRunes(parsed.Text, maxChars)
	sum := sha256.Sum256(chosen.Data)
	status := "ready"
	if parsed.Partial {
		status = "partial"
	}
	if strings.TrimSpace(text) == "" {
		if parsed.NeedsOCR {
			status = "failed"
		} else if parsed.Partial {
			status = "partial"
		} else {
			status = "failed"
		}
	}
	ok := status == "ready" || status == "partial"
	out := map[string]interface{}{
		"extract_ok":      ok,
		"extract_status":  status,
		"extract_kind":    parsed.Kind,
		"extract_error":   "",
		"extract_partial": parsed.Partial,
		"needs_ocr":       parsed.NeedsOCR,
		"file_hash":       hex.EncodeToString(sum[:]),
		"extracted_at":    time.Now().UTC().Format(time.RFC3339),
		"extract_name":    chosen.Name,
		outKey:            text,
	}
	if !ok && parsed.NeedsOCR {
		out["extract_error"] = "scanned PDF has no text layer"
		out["extract_status"] = "failed"
	}
	if !ok {
		if errMsg, _ := out["extract_error"].(string); errMsg == "" {
			out["extract_error"] = "no text extracted"
		}
	}
	promoteExtract(ec, out, outKey)
	return out, nil
}

func promoteExtract(ec *rulesgo.ExecutionContext, out map[string]interface{}, textKey string) {
	if ec == nil {
		return
	}
	for _, k := range []string{"extract_ok", "extract_status", "extract_kind", "extract_error", "extract_partial", "needs_ocr", "file_hash", "extracted_at", "extract_name", textKey} {
		if v, ok := out[k]; ok {
			ec.Set(k, v)
		}
	}
}

func pickExtractAttachment(files []*rulesgo.AttachmentBytes) *rulesgo.AttachmentBytes {
	best := files[0]
	bestRank := rag.ExtractRank(rag.DetectKind(best.Name, best.MIME))
	for _, f := range files[1:] {
		r := rag.ExtractRank(rag.DetectKind(f.Name, f.MIME))
		if r > bestRank {
			best, bestRank = f, r
		}
	}
	return best
}

func extractUint64(v interface{}) uint64 {
	switch t := v.(type) {
	case nil:
		return 0
	case uint64:
		return t
	case int:
		if t > 0 {
			return uint64(t)
		}
	case int64:
		if t > 0 {
			return uint64(t)
		}
	case float64:
		if t > 0 {
			return uint64(t)
		}
	case string:
		var n uint64
		fmt.Sscanf(strings.TrimSpace(t), "%d", &n)
		return n
	}
	return 0
}
