package rest

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type Document struct{}

func MountDocumentRoutes(r chi.Router) {
	ctrl := Document{}
	r.Get("/namespace/{namespaceID}/document/", ctrl.List)
	r.Post("/namespace/{namespaceID}/document/", ctrl.Create)
	r.Post("/namespace/{namespaceID}/document/reorder", ctrl.Reorder)
	r.Get("/namespace/{namespaceID}/document/{documentID}", ctrl.Read)
	r.Post("/namespace/{namespaceID}/document/{documentID}", ctrl.Update)
	r.Put("/namespace/{namespaceID}/document/{documentID}", ctrl.Update)
	r.Delete("/namespace/{namespaceID}/document/{documentID}", ctrl.Delete)
	r.Post("/namespace/{namespaceID}/document/{documentID}/file", ctrl.Upload)
	r.Get("/namespace/{namespaceID}/document/{documentID}/file", ctrl.File)
}

func (Document) List(w http.ResponseWriter, r *http.Request) {
	namespaceID, err := documentNamespaceID(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	set, err := service.Document().Find(r.Context(), namespaceID, r.URL.Query().Get("query"))
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"set": set})
}

func (Document) Read(w http.ResponseWriter, r *http.Request) {
	namespaceID, documentID, err := documentIDs(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	doc, err := service.Document().FindByID(r.Context(), namespaceID, documentID)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"document": doc})
}

func (Document) Create(w http.ResponseWriter, r *http.Request) {
	namespaceID, err := documentNamespaceID(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	in, err := decodeDocumentWrite(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	doc, err := service.Document().Create(r.Context(), namespaceID, in)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"document": doc})
}

func (Document) Update(w http.ResponseWriter, r *http.Request) {
	namespaceID, documentID, err := documentIDs(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	in, err := decodeDocumentWrite(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	doc, err := service.Document().Update(r.Context(), namespaceID, documentID, in)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"document": doc})
}

func (Document) Delete(w http.ResponseWriter, r *http.Request) {
	namespaceID, documentID, err := documentIDs(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if err = service.Document().Delete(r.Context(), namespaceID, documentID); err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"deleted": true})
}

func (Document) Reorder(w http.ResponseWriter, r *http.Request) {
	namespaceID, err := documentNamespaceID(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	defer r.Body.Close()
	var payload struct {
		DocumentIDs []string `json:"documentIDs"`
	}
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.Send(w, r, err)
		return
	}
	ids := make([]uint64, 0, len(payload.DocumentIDs))
	for _, raw := range payload.DocumentIDs {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			api.Send(w, r, err)
			return
		}
		ids = append(ids, id)
	}
	if err = service.Document().Reorder(r.Context(), namespaceID, ids); err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"reordered": true})
}

func (Document) Upload(w http.ResponseWriter, r *http.Request) {
	namespaceID, documentID, err := documentIDs(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	if err = r.ParseMultipartForm(32 << 20); err != nil {
		api.Send(w, r, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		api.Send(w, r, err)
		return
	}
	defer file.Close()
	doc, err := service.Document().Upload(r.Context(), namespaceID, documentID, header.Filename, header.Size, file)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	api.Send(w, r, map[string]interface{}{"document": doc})
}

func (Document) File(w http.ResponseWriter, r *http.Request) {
	namespaceID, documentID, err := documentIDs(r)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	_, att, fh, err := service.Document().OpenFile(r.Context(), namespaceID, documentID)
	if err != nil {
		api.Send(w, r, err)
		return
	}
	defer fh.Close()

	name := url.QueryEscape(att.Name)
	if mime := att.Meta.Original.Mimetype; mime != "" {
		w.Header().Set("Content-Type", mime)
	} else {
		w.Header().Set("Content-Type", "application/pdf")
	}
	w.Header().Set("Content-Disposition", "inline; filename="+name)
	http.ServeContent(w, r, att.Name, att.CreatedAt, fh)
}

func decodeDocumentWrite(r *http.Request) (service.DocumentWrite, error) {
	defer r.Body.Close()
	var in service.DocumentWrite
	err := json.NewDecoder(r.Body).Decode(&in)
	return in, err
}

func documentNamespaceID(r *http.Request) (uint64, error) {
	return strconv.ParseUint(chi.URLParam(r, "namespaceID"), 10, 64)
}

func documentIDs(r *http.Request) (uint64, uint64, error) {
	namespaceID, err := documentNamespaceID(r)
	if err != nil {
		return 0, 0, err
	}
	documentID, err := strconv.ParseUint(chi.URLParam(r, "documentID"), 10, 64)
	return namespaceID, documentID, err
}
