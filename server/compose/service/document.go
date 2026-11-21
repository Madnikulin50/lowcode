package service

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/handle"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms"
)

type (
	documentAccess interface {
		CanReadNamespace(context.Context, *types.Namespace) bool
		CanUpdateNamespace(context.Context, *types.Namespace) bool
	}

	document struct {
		ac documentAccess
	}

	DocumentWrite struct {
		Title   string `json:"title"`
		Handle  string `json:"handle"`
		Kind    string `json:"kind"`
		Body    string `json:"body"`
		Visible *bool  `json:"visible"`
		Weight  *int   `json:"weight"`
	}
)

func Document() *document {
	return &document{ac: DefaultAccessControl}
}

func (svc *document) store() (*rdbms.Store, error) {
	rs := rdbmsStore(DefaultStore)
	if rs == nil {
		return nil, errors.Internal("store type not supported")
	}
	return rs, nil
}

func (svc *document) namespace(ctx context.Context, namespaceID uint64) (*types.Namespace, error) {
	if namespaceID == 0 {
		return nil, errors.InvalidData("namespaceID is required")
	}
	return Namespace().FindByID(ctx, namespaceID)
}

func (svc *document) Find(ctx context.Context, namespaceID uint64, query string) (types.DocumentSet, error) {
	ns, err := svc.namespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	rs, err := svc.store()
	if err != nil {
		return nil, err
	}

	f := types.DocumentFilter{
		NamespaceID: namespaceID,
		Query:       query,
		VisibleOnly: !svc.ac.CanUpdateNamespace(ctx, ns),
	}
	f.Limit = 500
	set, _, err := rdbms.SearchDocuments(ctx, rs, f)
	if err != nil {
		return nil, err
	}
	for _, doc := range set {
		doc.Body = ""
	}
	return set, nil
}

func (svc *document) FindByID(ctx context.Context, namespaceID, documentID uint64) (*types.Document, error) {
	ns, err := svc.namespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	doc, err := svc.lookup(ctx, namespaceID, documentID)
	if err != nil {
		return nil, err
	}
	if !doc.Visible && !svc.ac.CanUpdateNamespace(ctx, ns) {
		return nil, errors.NotFound("document not found")
	}
	return doc, nil
}

func (svc *document) Create(ctx context.Context, namespaceID uint64, in DocumentWrite) (*types.Document, error) {
	ns, err := svc.namespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	if !svc.ac.CanUpdateNamespace(ctx, ns) {
		return nil, errors.Unauthorized("not allowed to update namespace")
	}

	doc := &types.Document{
		ID:          nextID(),
		NamespaceID: namespaceID,
		Title:       strings.TrimSpace(in.Title),
		Handle:      strings.TrimSpace(in.Handle),
		Kind:        strings.TrimSpace(in.Kind),
		Body:        in.Body,
		Visible:     true,
		CreatedAt:   time.Now().UTC(),
	}
	if in.Visible != nil {
		doc.Visible = *in.Visible
	}
	if doc.Kind == "" {
		doc.Kind = types.DocumentKindMarkdown
	}
	if err = validateDocument(doc); err != nil {
		return nil, err
	}
	if doc.Kind == types.DocumentKindPDF {
		doc.Body = ""
	}

	rs, err := svc.store()
	if err != nil {
		return nil, err
	}
	if err = svc.uniqueHandle(ctx, rs, doc); err != nil {
		return nil, err
	}

	set, _, err := rdbms.SearchDocuments(ctx, rs, types.DocumentFilter{NamespaceID: namespaceID})
	if err != nil {
		return nil, err
	}
	doc.Weight = len(set)
	if in.Weight != nil {
		doc.Weight = *in.Weight
	}

	if err = rdbms.CreateDocument(ctx, rs, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (svc *document) Update(ctx context.Context, namespaceID, documentID uint64, in DocumentWrite) (*types.Document, error) {
	doc, err := svc.writable(ctx, namespaceID, documentID)
	if err != nil {
		return nil, err
	}

	if title := strings.TrimSpace(in.Title); title != "" {
		doc.Title = title
	}
	doc.Handle = strings.TrimSpace(in.Handle)
	if doc.Kind == types.DocumentKindMarkdown {
		doc.Body = in.Body
	}
	if in.Visible != nil {
		doc.Visible = *in.Visible
	}
	if in.Weight != nil {
		doc.Weight = *in.Weight
	}
	if err = validateDocument(doc); err != nil {
		return nil, err
	}

	rs, err := svc.store()
	if err != nil {
		return nil, err
	}
	if err = svc.uniqueHandle(ctx, rs, doc); err != nil {
		return nil, err
	}
	if err = rdbms.UpdateDocument(ctx, rs, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (svc *document) Delete(ctx context.Context, namespaceID, documentID uint64) error {
	if _, err := svc.writable(ctx, namespaceID, documentID); err != nil {
		return err
	}
	rs, err := svc.store()
	if err != nil {
		return err
	}
	return rdbms.DeleteDocumentByID(ctx, rs, documentID)
}

func (svc *document) Reorder(ctx context.Context, namespaceID uint64, ids []uint64) error {
	ns, err := svc.namespace(ctx, namespaceID)
	if err != nil {
		return err
	}
	if !svc.ac.CanUpdateNamespace(ctx, ns) {
		return errors.Unauthorized("not allowed to update namespace")
	}
	rs, err := svc.store()
	if err != nil {
		return err
	}
	for i, id := range ids {
		doc, err := svc.lookup(ctx, namespaceID, id)
		if err != nil {
			return err
		}
		doc.Weight = i
		if err = rdbms.UpdateDocument(ctx, rs, doc); err != nil {
			return err
		}
	}
	return nil
}

func (svc *document) Upload(ctx context.Context, namespaceID, documentID uint64, name string, size int64, fh io.ReadSeeker) (*types.Document, error) {
	doc, err := svc.writable(ctx, namespaceID, documentID)
	if err != nil {
		return nil, err
	}
	if doc.Kind != types.DocumentKindPDF {
		return nil, errors.InvalidData("file upload is only for pdf documents")
	}
	if DefaultAttachment == nil {
		return nil, errors.Internal("attachment service not ready")
	}
	att, err := DefaultAttachment.CreateDocumentAttachment(ctx, namespaceID, name, size, fh)
	if err != nil {
		return nil, err
	}
	doc.AttachmentID = att.ID
	doc.FileName = att.Name
	rs, err := svc.store()
	if err != nil {
		return nil, err
	}
	if err = rdbms.UpdateDocument(ctx, rs, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (svc *document) OpenFile(ctx context.Context, namespaceID, documentID uint64) (*types.Document, *types.Attachment, io.ReadSeekCloser, error) {
	doc, err := svc.FindByID(ctx, namespaceID, documentID)
	if err != nil {
		return nil, nil, nil, err
	}
	if doc.Kind != types.DocumentKindPDF || doc.AttachmentID == 0 {
		return nil, nil, nil, errors.NotFound("document file not found")
	}
	if DefaultAttachment == nil {
		return nil, nil, nil, errors.Internal("attachment service not ready")
	}
	att, err := DefaultAttachment.FindByID(ctx, namespaceID, doc.AttachmentID)
	if err != nil {
		return nil, nil, nil, err
	}
	if att.NamespaceID != namespaceID || att.Kind != types.DocumentAttachment {
		return nil, nil, nil, errors.NotFound("document file not found")
	}
	fh, err := DefaultAttachment.OpenOriginal(att)
	if err != nil {
		return nil, nil, nil, err
	}
	return doc, att, fh, nil
}

func (svc *document) writable(ctx context.Context, namespaceID, documentID uint64) (*types.Document, error) {
	ns, err := svc.namespace(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	if !svc.ac.CanUpdateNamespace(ctx, ns) {
		return nil, errors.Unauthorized("not allowed to update namespace")
	}
	return svc.lookup(ctx, namespaceID, documentID)
}

func (svc *document) lookup(ctx context.Context, namespaceID, documentID uint64) (*types.Document, error) {
	if documentID == 0 {
		return nil, errors.InvalidData("documentID is required")
	}
	rs, err := svc.store()
	if err != nil {
		return nil, err
	}
	doc, err := rdbms.LookupDocumentByID(ctx, rs, documentID)
	if err != nil {
		return nil, err
	}
	if doc.NamespaceID != namespaceID {
		return nil, errors.NotFound("document not found")
	}
	return doc, nil
}

func (svc *document) uniqueHandle(ctx context.Context, rs *rdbms.Store, doc *types.Document) error {
	if doc.Handle == "" {
		return nil
	}
	existing, err := rdbms.LookupDocumentByHandle(ctx, rs, doc.NamespaceID, doc.Handle)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != doc.ID {
		return errors.DuplicateData("handle not unique")
	}
	return nil
}

func validateDocument(doc *types.Document) error {
	if doc.Title == "" {
		return errors.InvalidData("title is required")
	}
	if doc.Kind != types.DocumentKindMarkdown && doc.Kind != types.DocumentKindPDF {
		return errors.InvalidData("kind must be markdown or pdf")
	}
	if !handle.IsValid(doc.Handle) {
		return errors.InvalidData("invalid handle")
	}
	return nil
}
