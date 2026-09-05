package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/dalutils"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/dal"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

const namespaceAttachmentManifestVersion = 1

var (
	reAttachmentURL = regexp.MustCompile(`/namespace/(\d+)/attachment/(record|page|icon|namespace)/(\d+)/`)
	reAttachmentID  = regexp.MustCompile(`/attachment/(record|page|icon|namespace)/(\d+)/`)
)

type (
	namespaceAttachmentManifest struct {
		Version     int                        `json:"version"`
		Attachments []namespaceAttachmentEntry `json:"attachments"`
	}

	namespaceAttachmentEntry struct {
		ID       uint64               `json:"id,string"`
		Kind     string               `json:"kind"`
		Name     string               `json:"name"`
		Meta     types.AttachmentMeta `json:"meta"`
		Original string               `json:"original"`
		Preview  string               `json:"preview,omitempty"`
	}
)

func PackNamespaceAttachments(ctx context.Context, zw *zip.Writer, ns *types.Namespace) error {
	if ns == nil {
		return nil
	}

	atts, err := collectNamespaceAttachments(ctx, ns)
	if err != nil {
		return err
	}
	if len(atts) == 0 {
		return nil
	}

	manifest := namespaceAttachmentManifest{
		Version:     namespaceAttachmentManifestVersion,
		Attachments: make([]namespaceAttachmentEntry, 0, len(atts)),
	}

	for _, att := range atts {
		entry, ok, err := packOneAttachment(zw, att)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		manifest.Attachments = append(manifest.Attachments, entry)
	}

	if len(manifest.Attachments) == 0 {
		return nil
	}

	raw, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	w, err := zw.Create("attachments/manifest.json")
	if err != nil {
		return err
	}
	_, err = w.Write(raw)
	return err
}

func packOneAttachment(zw *zip.Writer, att *types.Attachment) (namespaceAttachmentEntry, bool, error) {
	var empty namespaceAttachmentEntry
	if DefaultAttachment == nil {
		return empty, false, fmt.Errorf("attachment service is not initialized")
	}

	orig, err := DefaultAttachment.OpenOriginal(att)
	if err != nil || orig == nil {
		logger.Default().Warn("skipping attachment without original blob",
			zap.Uint64("attachmentID", att.ID),
			zap.Error(err),
		)
		return empty, false, nil
	}
	defer orig.Close()

	ext := att.Meta.Original.Extension
	if ext == "" {
		ext = strings.Trim(path.Ext(strings.Trim(att.Name, ".")), ".")
	}
	if ext == "" {
		ext = "bin"
	}
	origPath := fmt.Sprintf("attachments/%d/original.%s", att.ID, ext)
	if err = copyToZip(zw, origPath, orig); err != nil {
		return empty, false, err
	}

	entry := namespaceAttachmentEntry{
		ID:       att.ID,
		Kind:     att.Kind,
		Name:     att.Name,
		Meta:     att.Meta,
		Original: origPath,
	}

	if att.PreviewUrl != "" {
		prev, pErr := DefaultAttachment.OpenPreview(att)
		if pErr != nil || prev == nil {
			logger.Default().Warn("skipping missing attachment preview",
				zap.Uint64("attachmentID", att.ID),
				zap.Error(pErr),
			)
		} else {
			defer prev.Close()
			pExt := "jpg"
			if att.Meta.Preview != nil && att.Meta.Preview.Extension != "" {
				pExt = att.Meta.Preview.Extension
			}
			prevPath := fmt.Sprintf("attachments/%d/preview.%s", att.ID, pExt)
			if err = copyToZip(zw, prevPath, prev); err != nil {
				return empty, false, err
			}
			entry.Preview = prevPath
		}
	}

	return entry, true, nil
}

func copyToZip(zw *zip.Writer, name string, r io.Reader) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)
	return err
}

func collectNamespaceAttachments(ctx context.Context, ns *types.Namespace) (types.AttachmentSet, error) {
	ids := map[uint64]struct{}{}
	add := func(id uint64) {
		if id > 0 {
			ids[id] = struct{}{}
		}
	}

	add(ns.Meta.LogoID)
	add(ns.Meta.IconID)

	if set, _, err := store.SearchComposeAttachments(ctx, DefaultStore, types.AttachmentFilter{
		NamespaceID: ns.ID,
	}); err != nil {
		logger.Default().Warn("could not list namespace attachments", zap.Error(err))
	} else {
		for _, a := range set {
			add(a.ID)
		}
	}

	pages, _, err := store.SearchComposePages(ctx, DefaultStore, types.PageFilter{NamespaceID: ns.ID})
	if err != nil {
		return nil, err
	}
	for _, p := range pages {
		collectAttachmentIDsFromPage(p, add)
	}

	modules, _, err := store.SearchComposeModules(ctx, DefaultStore, types.ModuleFilter{NamespaceID: ns.ID})
	if err != nil {
		return nil, err
	}
	for _, m := range modules {
		m.Fields, _, err = store.SearchComposeModuleFields(ctx, DefaultStore, types.ModuleFieldFilter{ModuleID: []uint64{m.ID}})
		if err != nil {
			return nil, err
		}
		if !moduleHasFileField(m) {
			continue
		}
		recs, _, err := dalutils.ComposeRecordsList(ctx, dal.Service(), m, types.RecordFilter{
			ModuleID:    m.ID,
			NamespaceID: ns.ID,
		})
		if err != nil {
			logger.Default().Warn("could not list records while collecting attachments",
				zap.Uint64("moduleID", m.ID),
				zap.Error(err),
			)
			continue
		}
		for _, rec := range recs {
			for _, f := range m.Fields {
				if f.Kind != "File" {
					continue
				}
				for _, v := range rec.Values.FilterByName(f.Name) {
					add(v.Ref)
					add(cast.ToUint64(v.Value))
				}
			}
		}
	}

	out := make(types.AttachmentSet, 0, len(ids))
	for id := range ids {
		att, err := store.LookupComposeAttachmentByID(ctx, DefaultStore, id)
		if err != nil {
			if err == store.ErrNotFound {
				logger.Default().Warn("attachment referenced but not found", zap.Uint64("attachmentID", id))
				continue
			}
			return nil, err
		}
		out = append(out, att)
	}
	return out, nil
}

func moduleHasFileField(m *types.Module) bool {
	for _, f := range m.Fields {
		if f.Kind == "File" {
			return true
		}
	}
	return false
}

func collectAttachmentIDsFromPage(p *types.Page, add func(uint64)) {
	if p == nil {
		return
	}
	if p.Config.NavItem.Icon != nil && (p.Config.NavItem.Icon.Type == types.IconTypeAttachment || string(p.Config.NavItem.Icon.Type) == "attachment") {
		add(attachmentIDFromURL(p.Config.NavItem.Icon.Src))
	}
	for _, b := range p.Blocks {
		if strings.EqualFold(b.Kind, "File") {
			for _, id := range attachmentIDsFromAny(b.Options["attachments"]) {
				add(id)
			}
		}
		walkAnyForAttachmentURLs(b.Options, add)
		walkAnyForAttachmentURLs(b.Meta, add)
	}
}

func walkAnyForAttachmentURLs(v any, add func(uint64)) {
	switch x := v.(type) {
	case string:
		add(attachmentIDFromURL(x))
	case []any:
		for _, item := range x {
			walkAnyForAttachmentURLs(item, add)
		}
	case map[string]any:
		for _, item := range x {
			walkAnyForAttachmentURLs(item, add)
		}
	}
}

func attachmentIDsFromAny(v any) []uint64 {
	var out []uint64
	switch x := v.(type) {
	case nil:
		return nil
	case []any:
		for _, item := range x {
			out = append(out, attachmentIDsFromAny(item)...)
		}
	case []string:
		for _, item := range x {
			if id := cast.ToUint64(item); id > 0 {
				out = append(out, id)
			}
		}
	default:
		if id := cast.ToUint64(v); id > 0 {
			out = append(out, id)
		}
	}
	return out
}

func attachmentIDFromURL(s string) uint64 {
	m := reAttachmentID.FindStringSubmatch(s)
	if len(m) < 3 {
		return 0
	}
	id, _ := strconv.ParseUint(m[2], 10, 64)
	return id
}

func UnpackNamespaceAttachments(ctx context.Context, zr *zip.Reader, ns *types.Namespace, includeRecordFiles bool) (map[uint64]uint64, error) {
	idMap := map[uint64]uint64{}
	if zr == nil || ns == nil {
		return idMap, nil
	}

	manifestFile := zipFileByName(zr, "attachments/manifest.json")
	if manifestFile == nil {
		return idMap, nil
	}

	rc, err := manifestFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var manifest namespaceAttachmentManifest
	if err = json.NewDecoder(rc).Decode(&manifest); err != nil {
		return nil, err
	}

	for _, entry := range manifest.Attachments {
		if !includeRecordFiles && entry.Kind == types.RecordAttachment {
			continue
		}

		orig, origSize, err := openZipReadSeeker(zr, entry.Original)
		if err != nil {
			logger.Default().Warn("skipping attachment with missing original in archive",
				zap.Uint64("attachmentID", entry.ID),
				zap.String("path", entry.Original),
				zap.Error(err),
			)
			continue
		}

		var preview io.ReadSeeker
		if entry.Preview != "" {
			preview, _, err = openZipReadSeeker(zr, entry.Preview)
			if err != nil {
				logger.Default().Warn("attachment preview missing in archive",
					zap.Uint64("attachmentID", entry.ID),
					zap.Error(err),
				)
				preview = nil
			}
		}

		nsID := ns.ID
		if entry.Kind == types.IconAttachment {
			nsID = 0
		}

		created, err := DefaultAttachment.CreateImported(ctx, nsID, entry.Kind, entry.Name, entry.Meta, orig, origSize, preview)
		if err != nil {
			return nil, err
		}
		idMap[entry.ID] = created.ID
	}

	return idMap, nil
}

func zipFileByName(zr *zip.Reader, name string) *zip.File {
	name = path.Clean(name)
	for _, f := range zr.File {
		if path.Clean(f.Name) == name {
			return f
		}
	}
	return nil
}

func openZipReadSeeker(zr *zip.Reader, name string) (*bytes.Reader, int64, error) {
	f := zipFileByName(zr, name)
	if f == nil {
		return nil, 0, fmt.Errorf("archive entry %q not found", name)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, 0, err
	}
	defer rc.Close()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(raw), int64(len(raw)), nil
}

func RemapImportedAttachmentRefs(ctx context.Context, ns *types.Namespace, idMap map[uint64]uint64) error {
	if ns == nil || len(idMap) == 0 {
		return nil
	}

	changed := remapNamespaceMeta(ctx, ns, idMap)
	if changed {
		ns.UpdatedAt = now()
		if err := store.UpdateComposeNamespace(ctx, DefaultStore, ns); err != nil {
			return err
		}
	}

	pages, _, err := store.SearchComposePages(ctx, DefaultStore, types.PageFilter{NamespaceID: ns.ID})
	if err != nil {
		return err
	}
	for _, p := range pages {
		if !remapPageAttachmentRefs(p, ns.ID, idMap) {
			continue
		}
		p.UpdatedAt = now()
		if err = store.UpdateComposePage(ctx, DefaultStore, p); err != nil {
			return err
		}
	}
	return nil
}

func remapNamespaceMeta(ctx context.Context, ns *types.Namespace, idMap map[uint64]uint64) bool {
	changed := false
	if ns.Meta.LogoID > 0 {
		if nid, ok := idMap[ns.Meta.LogoID]; ok {
			if att, err := store.LookupComposeAttachmentByID(ctx, DefaultStore, nid); err == nil && att != nil {
				ns.Meta.LogoID = nid
				ns.Meta.Logo = attachmentPublicURL(ns.ID, att)
				changed = true
			}
		} else {
			ns.Meta.LogoID = 0
			ns.Meta.Logo = ""
			changed = true
		}
	}
	if ns.Meta.IconID > 0 {
		if nid, ok := idMap[ns.Meta.IconID]; ok {
			if att, err := store.LookupComposeAttachmentByID(ctx, DefaultStore, nid); err == nil && att != nil {
				ns.Meta.IconID = nid
				ns.Meta.Icon = attachmentPublicURL(ns.ID, att)
				changed = true
			}
		} else {
			ns.Meta.IconID = 0
			ns.Meta.Icon = ""
			changed = true
		}
	}
	return changed
}

func remapPageAttachmentRefs(p *types.Page, newNsID uint64, idMap map[uint64]uint64) bool {
	changed := false
	if p.Config.NavItem.Icon != nil {
		src := remapAttachmentURLs(p.Config.NavItem.Icon.Src, newNsID, idMap)
		if src != p.Config.NavItem.Icon.Src {
			p.Config.NavItem.Icon.Src = src
			changed = true
		}
	}
	for i, b := range p.Blocks {
		if strings.EqualFold(b.Kind, "File") && b.Options != nil {
			if raw, ok := b.Options["attachments"]; ok {
				remapped := remapFileAttachmentsList(raw, idMap)
				if !anyEqual(remapped, raw) {
					p.Blocks[i].Options["attachments"] = remapped
					changed = true
				}
			}
		}
		opts := remapAttachmentURLsInAny(b.Options, newNsID, idMap)
		meta := remapAttachmentURLsInAny(b.Meta, newNsID, idMap)
		if !anyEqual(opts, b.Options) || !anyEqual(meta, b.Meta) {
			p.Blocks[i].Options, _ = opts.(map[string]interface{})
			if meta != nil {
				p.Blocks[i].Meta, _ = meta.(map[string]any)
			}
			changed = true
		}
	}
	return changed
}

func remapFileAttachmentsList(v any, idMap map[uint64]uint64) any {
	switch x := v.(type) {
	case []any:
		out := make([]any, 0, len(x))
		for _, item := range x {
			mapped := remapIDString(strconv.FormatUint(cast.ToUint64(item), 10), idMap)
			if mapped != "" {
				out = append(out, mapped)
			}
		}
		return out
	case []string:
		out := make([]string, 0, len(x))
		for _, item := range x {
			mapped := remapIDString(item, idMap)
			if mapped != "" {
				out = append(out, mapped)
			}
		}
		return out
	default:
		mapped := remapIDString(strconv.FormatUint(cast.ToUint64(v), 10), idMap)
		if mapped == "" {
			return v
		}
		return mapped
	}
}

func anyEqual(a, b any) bool {
	ra, err1 := json.Marshal(a)
	rb, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return bytes.Equal(ra, rb)
}

func remapAttachmentURLsInAny(v any, newNsID uint64, idMap map[uint64]uint64) any {
	switch x := v.(type) {
	case string:
		return remapAttachmentURLs(x, newNsID, idMap)
	case []any:
		out := make([]any, len(x))
		for i, item := range x {
			out[i] = remapAttachmentURLsInAny(item, newNsID, idMap)
		}
		return out
	case []string:
		out := make([]string, len(x))
		for i, item := range x {
			out[i] = remapAttachmentURLs(item, newNsID, idMap)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, item := range x {
			out[k] = remapAttachmentURLsInAny(item, newNsID, idMap)
		}
		return out
	default:
		return v
	}
}

func remapAttachmentURLs(s string, newNsID uint64, idMap map[uint64]uint64) string {
	if s == "" || !strings.Contains(s, "/attachment/") {
		return s
	}

	out := reAttachmentURL.ReplaceAllStringFunc(s, func(m string) string {
		parts := reAttachmentURL.FindStringSubmatch(m)
		if len(parts) < 4 {
			return m
		}
		kind := parts[2]
		oldID, _ := strconv.ParseUint(parts[3], 10, 64)
		nid, ok := idMap[oldID]
		if !ok {
			return m
		}
		nsID := newNsID
		if kind == types.IconAttachment {
			nsID = 0
		}
		return fmt.Sprintf("/namespace/%d/attachment/%s/%d/", nsID, kind, nid)
	})
	return out
}

func attachmentPublicURL(nsID uint64, att *types.Attachment) string {
	if att == nil {
		return ""
	}
	id := att.NamespaceID
	if id == 0 && att.Kind != types.IconAttachment {
		id = nsID
	}
	return fmt.Sprintf("/namespace/%d/attachment/%s/%d/original/%s", id, att.Kind, att.ID, url.PathEscape(att.Name))
}

func RemapRecordFileJSONL(src io.Reader, mod *types.Module, idMap map[uint64]uint64) (*bytes.Reader, error) {
	if src == nil {
		return bytes.NewReader(nil), nil
	}
	if len(idMap) == 0 || mod == nil {
		raw, err := io.ReadAll(src)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(raw), nil
	}

	fileFields := map[string]bool{}
	for _, f := range mod.Fields {
		if f.Kind == "File" {
			fileFields[f.Name] = true
		}
	}
	if len(fileFields) == 0 {
		raw, err := io.ReadAll(src)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(raw), nil
	}

	dec := json.NewDecoder(src)
	dec.UseNumber()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for {
		var row map[string]any
		if err := dec.Decode(&row); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		for name := range fileFields {
			if v, ok := row[name]; ok {
				mapped := remapFileFieldValue(v, idMap)
				if mapped == nil || mapped == "" {
					delete(row, name)
				} else {
					row[name] = mapped
				}
			}
		}
		if err := enc.Encode(row); err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(buf.Bytes()), nil
}

func remapFileFieldValue(v any, idMap map[uint64]uint64) any {
	switch x := v.(type) {
	case nil:
		return nil
	case json.Number:
		return remapIDString(x.String(), idMap)
	case string:
		if strings.Contains(x, ";") {
			parts := strings.Split(x, ";")
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p == "" {
					continue
				}
				mapped := remapIDString(p, idMap)
				if mapped != "" {
					out = append(out, mapped)
				}
			}
			return strings.Join(out, ";")
		}
		return remapIDString(x, idMap)
	case []any:
		out := make([]any, 0, len(x))
		for _, item := range x {
			mapped := remapFileFieldValue(item, idMap)
			if mapped == nil || mapped == "" {
				continue
			}
			out = append(out, mapped)
		}
		return out
	default:
		return remapIDString(anyToIDString(v), idMap)
	}
}

func anyToIDString(v any) string {
	switch x := v.(type) {
	case json.Number:
		return x.String()
	case string:
		return x
	case uint64:
		return strconv.FormatUint(x, 10)
	case int64:
		if x < 0 {
			return ""
		}
		return strconv.FormatUint(uint64(x), 10)
	default:
		return cast.ToString(v)
	}
}

func remapIDString(s string, idMap map[uint64]uint64) string {
	id := cast.ToUint64(s)
	if id == 0 {
		return s
	}
	if nid, ok := idMap[id]; ok {
		return strconv.FormatUint(nid, 10)
	}
	return ""
}
