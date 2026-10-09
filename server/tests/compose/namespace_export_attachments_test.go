package compose

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/dalutils"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/tests/helpers"
	"github.com/spf13/cast"
)

func Test_namespace_export_attachments(t *testing.T) {
	png, err := os.ReadFile("./testdata/test.png")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("round trip record page and logo", func(t *testing.T) {
		ctx, h, s := setup(t)
		h.clearRecords()
		h.clearPages()
		h.clearNamespaces()
		grantImportExport(h)
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "export")
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "modules.export")
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "charts.export")
		helpers.AllowMeRecordCRUD(h)
		helpers.AllowMe(h, types.PageRbacResource(0, 0), "update")

		src := seedNamespaceWithAttachments(t, h, ctx, png)
		assertSeededRecordFile(t, h, ctx, src)

		arch := namespaceExportSafe(t, h, src.ns.ID)
		assertZipHasAttachments(t, h, arch)
		assertZipRecordFileJSONL(t, h, arch, src.recordAttID)

		sessionID := namespaceImportInitSafe(t, h, arch)
		ns, mm, pp, _ := namespaceImportRunData(ctx, s, t, h, sessionID, "imported att", "imported_att", true)

		h.a.NotEqual(src.ns.ID, ns.ID)
		h.a.NotZero(ns.Meta.LogoID)
		h.a.NotEqual(src.logoID, ns.Meta.LogoID)
		h.a.Contains(ns.Meta.Logo, fmt.Sprintf("/namespace/%d/attachment/namespace/%d/", ns.ID, ns.Meta.LogoID))
		h.a.NotZero(ns.Meta.IconID)
		h.a.NotEqual(src.iconID, ns.Meta.IconID)
		h.a.Contains(ns.Meta.Icon, fmt.Sprintf("/namespace/%d/attachment/namespace/%d/", ns.ID, ns.Meta.IconID))

		logo, err := store.LookupComposeAttachmentByID(ctx, s, ns.Meta.LogoID)
		h.a.NoError(err)
		assertAttachmentBytes(t, h, logo, png)
		iconAtt, err := store.LookupComposeAttachmentByID(ctx, s, ns.Meta.IconID)
		h.a.NoError(err)
		assertAttachmentBytes(t, h, iconAtt, png)

		mod := findModuleByName(mm, "docs")
		h.a.NotNil(mod)
		recs, _, err := dalutils.ComposeRecordsList(ctx, defDal, mod, types.RecordFilter{
			ModuleID:    mod.ID,
			NamespaceID: ns.ID,
		})
		h.a.NoError(err)
		h.a.Len(recs, 1)

		fileVals := recs[0].Values.FilterByName("file")
		h.a.Len(fileVals, 1)
		newFileID := cast.ToUint64(fileVals[0].Value)
		if newFileID == 0 {
			newFileID = fileVals[0].Ref
		}
		h.a.NotZero(newFileID)
		h.a.NotEqual(src.recordAttID, newFileID)

		fileAtt, err := store.LookupComposeAttachmentByID(ctx, s, newFileID)
		h.a.NoError(err)
		h.a.Equal("doc.png", fileAtt.Name)
		assertAttachmentBytes(t, h, fileAtt, png)

		pg := pp.FindByHandle("pg_files")
		h.a.NotNil(pg)
		pageAttIDs := pageFileBlockIDs(pg)
		h.a.Len(pageAttIDs, 1)
		h.a.NotEqual(src.pageAttID, pageAttIDs[0])
		pageAtt, err := store.LookupComposeAttachmentByID(ctx, s, pageAttIDs[0])
		h.a.NoError(err)
		assertAttachmentBytes(t, h, pageAtt, png)

		cleanup(t)
	})

	t.Run("importData false keeps schema files skips records", func(t *testing.T) {
		ctx, h, s := setup(t)
		h.clearRecords()
		h.clearPages()
		h.clearNamespaces()
		grantImportExport(h)
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "export")
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "modules.export")
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "charts.export")
		helpers.AllowMeRecordCRUD(h)
		helpers.AllowMe(h, types.PageRbacResource(0, 0), "update")

		src := seedNamespaceWithAttachments(t, h, ctx, png)
		arch := namespaceExportSafe(t, h, src.ns.ID)
		sessionID := namespaceImportInitSafe(t, h, arch)
		ns, mm, pp, _ := namespaceImportRunData(ctx, s, t, h, sessionID, "imported nodata", "imported_nodata", false)

		h.a.NotZero(ns.Meta.LogoID)
		h.a.NotZero(ns.Meta.IconID)
		logo, err := store.LookupComposeAttachmentByID(ctx, s, ns.Meta.LogoID)
		h.a.NoError(err)
		assertAttachmentBytes(t, h, logo, png)

		pg := pp.FindByHandle("pg_files")
		h.a.NotNil(pg)
		h.a.Len(pageFileBlockIDs(pg), 1)

		mod := findModuleByName(mm, "docs")
		h.a.NotNil(mod)
		recs, _, err := dalutils.ComposeRecordsList(ctx, defDal, mod, types.RecordFilter{
			ModuleID:    mod.ID,
			NamespaceID: ns.ID,
		})
		h.a.NoError(err)
		h.a.Len(recs, 0)

		recAtts, _, err := store.SearchComposeAttachments(ctx, s, types.AttachmentFilter{
			NamespaceID: ns.ID,
			Kind:        types.RecordAttachment,
		})
		h.a.NoError(err)
		h.a.Len(recAtts, 0)

		cleanup(t)
	})

	t.Run("archive without attachments folder imports schema", func(t *testing.T) {
		ctx, h, s := setup(t)
		h.clearRecords()
		h.clearPages()
		h.clearNamespaces()
		grantImportExport(h)
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "export")
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "modules.export")
		helpers.AllowMe(h, types.NamespaceRbacResource(0), "charts.export")
		helpers.AllowMeRecordCRUD(h)
		helpers.AllowMe(h, types.PageRbacResource(0, 0), "update")

		src := seedNamespaceWithAttachments(t, h, ctx, png)
		arch := stripZipAttachments(t, h, namespaceExportSafe(t, h, src.ns.ID))
		sessionID := namespaceImportInitSafe(t, h, arch)
		ns, mm, pp, _ := namespaceImportRunData(ctx, s, t, h, sessionID, "imported noatt", "imported_noatt", true)

		h.a.NotEqual(src.ns.ID, ns.ID)
		h.a.NotNil(findModuleByName(mm, "docs"))
		h.a.NotNil(pp.FindByHandle("pg_files"))

		cleanup(t)
	})
}

type seededNamespace struct {
	ns          *types.Namespace
	logoID      uint64
	iconID      uint64
	recordAttID uint64
	pageAttID   uint64
	mod         *types.Module
	rec         *types.Record
}

func seedNamespaceWithAttachments(t *testing.T, h helper, ctx context.Context, png []byte) seededNamespace {
	t.Helper()

	ns := h.makeNamespace("att_ns_" + rs())
	mod := h.makeModule(ns, "docs",
		&types.ModuleField{Name: "title", Kind: "String"},
		&types.ModuleField{Name: "file", Kind: "File"},
	)

	recAtt, err := service.DefaultAttachment.CreateRecordAttachment(ctx, ns.ID, "doc.png", int64(len(png)), bytes.NewReader(png), mod.ID, 0, "file")
	h.a.NoError(err)

	rec := h.makeRecord(mod,
		&types.RecordValue{Name: "title", Value: "hello"},
		&types.RecordValue{Name: "file", Value: strconv.FormatUint(recAtt.ID, 10), Ref: recAtt.ID},
	)

	page := h.repoMakePage(ns, "files page")
	page.Handle = "pg_files"
	pageAtt, err := service.DefaultAttachment.CreatePageAttachment(ctx, ns.ID, "page.png", int64(len(png)), bytes.NewReader(png), page.ID)
	h.a.NoError(err)
	page.Blocks = types.PageBlocks{
		{
			BlockID: 1,
			Kind:    "File",
			Options: map[string]interface{}{
				"attachments": []interface{}{strconv.FormatUint(pageAtt.ID, 10)},
			},
		},
	}
	h.a.NoError(store.UpdateComposePage(ctx, service.DefaultStore, page))

	logo, err := service.DefaultAttachment.CreateNamespaceAttachment(ctx, "logo.png", int64(len(png)), bytes.NewReader(png))
	h.a.NoError(err)
	icon, err := service.DefaultAttachment.CreateNamespaceAttachment(ctx, "icon.png", int64(len(png)), bytes.NewReader(png))
	h.a.NoError(err)
	ns.Meta.LogoID = logo.ID
	ns.Meta.LogoEnabled = true
	ns.Meta.Logo = fmt.Sprintf("/namespace/%d/attachment/namespace/%d/original/logo.png", ns.ID, logo.ID)
	ns.Meta.IconID = icon.ID
	ns.Meta.Icon = fmt.Sprintf("/namespace/%d/attachment/namespace/%d/original/icon.png", ns.ID, icon.ID)
	h.a.NoError(store.UpdateComposeNamespace(ctx, service.DefaultStore, ns))

	return seededNamespace{
		ns:          ns,
		logoID:      logo.ID,
		iconID:      icon.ID,
		recordAttID: recAtt.ID,
		pageAttID:   pageAtt.ID,
		mod:         mod,
		rec:         rec,
	}
}

func assertSeededRecordFile(t *testing.T, h helper, ctx context.Context, src seededNamespace) {
	t.Helper()
	got := h.lookupRecordByID(src.mod, src.rec.ID)
	fileVals := got.Values.FilterByName("file")
	h.a.Len(fileVals, 1, "seeded record must keep File field before export")
	h.a.Equal(src.recordAttID, fileVals[0].Ref)
}

func assertZipRecordFileJSONL(t *testing.T, h helper, arch []byte, recordAttID uint64) {
	t.Helper()
	raw := zipFileBytes(t, h, arch, "data/docs.json")
	h.a.Contains(string(raw), strconv.FormatUint(recordAttID, 10), "export JSONL must contain source File attachment ID")
}

func zipFileBytes(t *testing.T, h helper, arch []byte, name string) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(arch), int64(len(arch)))
	h.a.NoError(err)
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		h.a.NoError(err)
		defer rc.Close()
		raw, err := io.ReadAll(rc)
		h.a.NoError(err)
		return raw
	}
	h.a.FailNow("zip entry not found: " + name)
	return nil
}

func stripZipAttachments(t *testing.T, h helper, arch []byte) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(arch), int64(len(arch)))
	h.a.NoError(err)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "attachments/") {
			continue
		}
		w, err := zw.Create(f.Name)
		h.a.NoError(err)
		rc, err := f.Open()
		h.a.NoError(err)
		_, err = io.Copy(w, rc)
		h.a.NoError(err)
		h.a.NoError(rc.Close())
	}
	h.a.NoError(zw.Close())
	return buf.Bytes()
}

func assertZipHasAttachments(t *testing.T, h helper, arch []byte) {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(arch), int64(len(arch)))
	h.a.NoError(err)
	found := false
	for _, f := range zr.File {
		if f.Name == "attachments/manifest.json" {
			found = true
			break
		}
	}
	h.a.True(found, "export zip must contain attachments/manifest.json")
}

func assertAttachmentBytes(t *testing.T, h helper, att *types.Attachment, want []byte) {
	t.Helper()
	r, err := service.DefaultAttachment.OpenOriginal(att)
	h.a.NoError(err)
	h.a.NotNil(r)
	defer r.Close()
	got, err := io.ReadAll(r)
	h.a.NoError(err)
	h.a.Equal(want, got)
}

func findModuleByName(mm types.ModuleSet, name string) *types.Module {
	for _, m := range mm {
		if m.Name == name || m.Handle == name {
			return m
		}
	}
	return nil
}

func pageFileBlockIDs(p *types.Page) []uint64 {
	var ids []uint64
	for _, b := range p.Blocks {
		if b.Kind != "File" {
			continue
		}
		raw, _ := b.Options["attachments"]
		switch x := raw.(type) {
		case []interface{}:
			for _, item := range x {
				if id := cast.ToUint64(item); id > 0 {
					ids = append(ids, id)
				}
			}
		case []string:
			for _, item := range x {
				if id := cast.ToUint64(item); id > 0 {
					ids = append(ids, id)
				}
			}
		default:
			if id := cast.ToUint64(raw); id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}
