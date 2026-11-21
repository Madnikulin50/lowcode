package rag

import "strings"

// DetectKind maps a filename + MIME type to an extractor kind.
func DetectKind(filename, mimetype string) string {
	name := strings.ToLower(strings.TrimSpace(filename))
	mime := strings.ToLower(strings.TrimSpace(mimetype))
	switch {
	case strings.HasSuffix(name, ".txt") || mime == "text/plain":
		return "txt"
	case strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm") || mime == "text/html":
		return "html"
	case strings.HasSuffix(name, ".docx") || mime == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx"
	case strings.HasSuffix(name, ".pdf") || mime == "application/pdf":
		return "pdf"
	case strings.HasSuffix(name, ".xlsx") || mime == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return "xlsx"
	case strings.HasSuffix(name, ".dxf") || strings.Contains(mime, "dxf"):
		return "dxf"
	case strings.HasSuffix(name, ".ifczip") || mime == "application/ifczip":
		return "ifczip"
	case strings.HasSuffix(name, ".ifc") || strings.Contains(mime, "ifc"):
		return "ifc"
	case strings.HasSuffix(name, ".bimx"):
		return "bimx"
	case strings.HasSuffix(name, ".dwg") || strings.Contains(mime, "dwg") || mime == "application/acad":
		return "dwg"
	case strings.HasSuffix(name, ".pln"):
		return "pln"
	case strings.HasSuffix(name, ".pla"):
		return "pla"
	default:
		return ""
	}
}

// ExtractRank prefers machine-readable CAD/Office over proprietary binaries
// when a File field holds several attachments.
func ExtractRank(kind string) int {
	switch kind {
	case "ifc":
		return 100
	case "ifczip":
		return 95
	case "dxf":
		return 90
	case "pdf":
		return 80
	case "docx":
		return 75
	case "xlsx":
		return 70
	case "html":
		return 65
	case "txt":
		return 60
	case "bimx":
		return 40
	case "dwg":
		return 20
	case "pln", "pla":
		return 10
	default:
		return 1
	}
}
