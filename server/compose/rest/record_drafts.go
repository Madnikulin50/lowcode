package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/payload"
)

// RecordDrafts keeps a user's unsaved record edits on the server
// (compose/service/record_draft.go). Everything works on the drafts of the
// signed-in user.
//
//	GET    /record-drafts?namespaceID=&moduleID=&recordID=
//	PUT    /record-drafts/{changeID}   {namespaceID, moduleID, recordID, payload}
//	DELETE /record-drafts/{changeID}
//	DELETE /record-drafts
//
// Hand-written (not in rest.yaml); the client methods are hand-written too.
type RecordDrafts struct{}

type (
	recordDraftPayload struct {
		ChangeID    string          `json:"changeID"`
		NamespaceID uint64          `json:"namespaceID,string"`
		ModuleID    uint64          `json:"moduleID,string"`
		RecordID    uint64          `json:"recordID,string"`
		UpdatedAt   string          `json:"updatedAt,omitempty"`
		Payload     json.RawMessage `json:"payload"`
	}

	saveRecordDraftRequest struct {
		NamespaceID uint64          `json:"namespaceID,string"`
		ModuleID    uint64          `json:"moduleID,string"`
		RecordID    uint64          `json:"recordID,string"`
		Payload     json.RawMessage `json:"payload"`
	}
)

func makeRecordDraftPayload(d *types.RecordDraft) recordDraftPayload {
	out := recordDraftPayload{
		ChangeID:    d.ChangeID,
		NamespaceID: d.NamespaceID,
		ModuleID:    d.ModuleID,
		RecordID:    d.RecordID,
		Payload:     json.RawMessage(d.Payload),
	}
	if d.UpdatedAt != nil {
		out.UpdatedAt = d.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")
	} else {
		out.UpdatedAt = d.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return out
}

func (RecordDrafts) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	set, err := service.DefaultRecordDraft.List(r.Context(), types.RecordDraftFilter{
		NamespaceID: payload.ParseUint64(q.Get("namespaceID")),
		ModuleID:    payload.ParseUint64(q.Get("moduleID")),
		RecordID:    payload.ParseUint64(q.Get("recordID")),
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}

	out := make([]recordDraftPayload, 0, len(set))
	for _, d := range set {
		out = append(out, makeRecordDraftPayload(d))
	}

	api.Send(w, r, map[string]interface{}{"set": out})
}

func (RecordDrafts) Save(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, (1<<20)+4096))
	if err != nil {
		api.Send(w, r, err)
		return
	}

	var in saveRecordDraftRequest
	if err = json.Unmarshal(body, &in); err != nil {
		api.Send(w, r, errors.InvalidData("invalid JSON: %v", err))
		return
	}

	d, err := service.DefaultRecordDraft.Save(r.Context(), &types.RecordDraft{
		ChangeID:    chi.URLParam(r, "changeID"),
		NamespaceID: in.NamespaceID,
		ModuleID:    in.ModuleID,
		RecordID:    in.RecordID,
		Payload:     string(in.Payload),
	})
	if err != nil {
		api.Send(w, r, err)
		return
	}

	api.Send(w, r, makeRecordDraftPayload(d))
}

func (RecordDrafts) Delete(w http.ResponseWriter, r *http.Request) {
	if err := service.DefaultRecordDraft.Delete(r.Context(), chi.URLParam(r, "changeID")); err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, api.OK())
}

func (RecordDrafts) Clear(w http.ResponseWriter, r *http.Request) {
	if err := service.DefaultRecordDraft.Clear(r.Context()); err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, api.OK())
}

func MountRecordDraftRoutes(r chi.Router) {
	d := RecordDrafts{}
	r.Get("/record-drafts", d.List)
	r.Put("/record-drafts/{changeID}", d.Save)
	r.Delete("/record-drafts/{changeID}", d.Delete)
	r.Delete("/record-drafts", d.Clear)
}
