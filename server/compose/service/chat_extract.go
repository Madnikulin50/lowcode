package service

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/rag"
)

func chatExtractAttachmentToolDef() chat.ToolDef {
	return chat.ToolDef{
		Name:        "extract_attachment_text",
		Description: "Extract text from a compose attachment (docx, xlsx, pdf, dxf, dwg, ifc, ArchiCAD). Prefer IFC/DXF/PDF over native PLN/DWG.",
		Params: []chat.ParamDef{
			{Name: "attachmentID", Type: "string", Required: true, Description: "Attachment ID"},
			{Name: "namespaceID", Type: "string", Required: false, Description: "Namespace ID (defaults to current)"},
		},
		Handler: chatExtractAttachment,
	}
}

func chatExtractAttachment(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	attID := parseUint64(params["attachmentID"])
	if attID == 0 {
		return "attachmentID is required"
	}
	att, err := RuleChainAttachmentLoader{}.Load(ctx, ns, attID)
	if err != nil {
		return fmt.Sprintf("load attachment: %v", err)
	}
	doc, err := rag.ParseDocument(att.Data, att.Name, att.MIME)
	if err != nil {
		return fmt.Sprintf("extract: %v", err)
	}
	text := rag.TruncateRunes(doc.Text, rag.DefaultExtractMaxChars)
	if text == "" {
		if doc.NeedsOCR {
			return "No text layer (scanned PDF)."
		}
		if doc.Partial {
			return fmt.Sprintf("Partial extract (%s): no usable text. Upload IFC, DXF or PDF if this is CAD.", doc.Kind)
		}
		return "No text extracted."
	}
	prefix := doc.Kind
	if doc.Partial {
		prefix += ", partial"
	}
	return fmt.Sprintf("[%s]\n%s", prefix, text)
}
