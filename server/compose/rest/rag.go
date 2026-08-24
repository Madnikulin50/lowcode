package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/rag"
)

type RAG struct {
	svc *service.RAGService
}

func NewRAG(svc *service.RAGService) *RAG {
	return &RAG{svc: svc}
}

func (RAG) New() *RAG {
	return &RAG{svc: service.DefaultRAG}
}

func (ctrl *RAG) MountRoutes(r chi.Router) {
	r.Route("/namespace/{namespaceID}/rag", func(r chi.Router) {
		r.Get("/documents", ctrl.List)
		r.Post("/documents", ctrl.Upload)
		r.Delete("/documents/{docID}", ctrl.Delete)
		r.Post("/search", ctrl.Search)
	})
	r.Get("/pages-rag", ctrl.PagesList)
	r.Post("/pages-rag/reindex", ctrl.PagesReindex)
}

type ragDocResponse struct {
	rag.Document
}

func (ctrl *RAG) List(w http.ResponseWriter, r *http.Request) {
	nsID := chi.URLParam(r, "namespaceID")
	docs, err := ctrl.svc.ListDocuments(r.Context(), nsID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if docs == nil {
		docs = []rag.Document{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"set": docs}})
}

func (ctrl *RAG) Upload(w http.ResponseWriter, r *http.Request) {
	nsID := chi.URLParam(r, "namespaceID")
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mimetype := header.Header.Get("Content-Type")
	doc, err := ctrl.svc.Ingest(r.Context(), nsID, header.Filename, data, mimetype)
	if err != nil {
		http.Error(w, "ingest failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"response": doc})
}

func (ctrl *RAG) Delete(w http.ResponseWriter, r *http.Request) {
	nsID := chi.URLParam(r, "namespaceID")
	docID := chi.URLParam(r, "docID")
	if err := ctrl.svc.Delete(r.Context(), nsID, docID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type searchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"topK"`
}

func (ctrl *RAG) Search(w http.ResponseWriter, r *http.Request) {
	nsID := chi.URLParam(r, "namespaceID")
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.TopK <= 0 {
		req.TopK = 3
	}
	results, err := ctrl.svc.Search(r.Context(), nsID, req.Query, req.TopK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"results": results}})
}

func (ctrl *RAG) PagesList(w http.ResponseWriter, r *http.Request) {
	if service.DefaultPagesRAG == nil {
		http.Error(w, "pages RAG not available", http.StatusNotFound)
		return
	}
	chunks, err := service.DefaultPagesRAG.ListPages(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if chunks == nil {
		chunks = []rag.PageChunk{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"response": map[string]any{"set": chunks}})
}

func (ctrl *RAG) PagesReindex(w http.ResponseWriter, r *http.Request) {
	if service.DefaultPagesRAG == nil {
		http.Error(w, "pages RAG not available", http.StatusNotFound)
		return
	}
	if err := service.DefaultPagesRAG.Reindex(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"response": true})
}
