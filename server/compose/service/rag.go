package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/rag"
)

type RAGService struct {
	store    *rag.Store
	embedder *rag.Embedder
}

func NewRAGService(store *rag.Store, embedder *rag.Embedder) *RAGService {
	return &RAGService{store: store, embedder: embedder}
}

func (s *RAGService) Ingest(ctx context.Context, namespaceID string, filename string, data []byte, mimetype string) (*rag.Document, error) {
	parsed, err := rag.ParseDocument(data, filename, mimetype)
	if err != nil {
		return nil, fmt.Errorf("ingest parse: %w", err)
	}

	id := hashID(namespaceID + "/" + filename + "/" + fmt.Sprint(time.Now().UnixNano()))
	doc := rag.Document{
		ID:        id,
		Namespace: namespaceID,
		Name:      filename,
		Size:      int64(len(data)),
		CreatedAt: time.Now().Unix(),
	}

	if err := s.store.PutDocument(doc); err != nil {
		return nil, fmt.Errorf("ingest put doc: %w", err)
	}

	chunks := rag.ChunkText(parsed.Text, 512, 64)
	for i, chunk := range chunks {
		emb, err := s.embedder.Embed(chunk)
		if err != nil {
			return nil, fmt.Errorf("ingest embed chunk %d: %w", i, err)
		}
		c := rag.Chunk{
			ID:         fmt.Sprintf("%s:%d", id, i),
			DocID:      id,
			Text:       chunk,
			Embedding:  emb,
			ChunkIndex: i,
		}
		if err := s.store.PutChunk(namespaceID, c); err != nil {
			return nil, fmt.Errorf("ingest put chunk %d: %w", i, err)
		}
	}

	return &doc, nil
}

func (s *RAGService) Delete(ctx context.Context, namespaceID, docID string) error {
	return s.store.DeleteDocument(namespaceID, docID)
}

func (s *RAGService) ListDocuments(ctx context.Context, namespaceID string) ([]rag.Document, error) {
	return s.store.ListDocuments(namespaceID)
}

func (s *RAGService) Search(ctx context.Context, namespaceID, query string, topK int) ([]rag.SearchResult, error) {
	emb, err := s.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("search embed: %w", err)
	}
	return s.store.Search(namespaceID, emb, topK)
}

func (s *RAGService) BuildContext(ctx context.Context, namespaceID, query string, topK int) string {
	results, err := s.Search(ctx, namespaceID, query, topK)
	if err != nil || len(results) == 0 {
		return ""
	}
	ctx2 := ""
	for i, r := range results {
		ctx2 += fmt.Sprintf("\n[Document %d]: %s\n", i+1, r.Chunk.Text)
	}
	return ctx2
}

func hashID(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:16])
}
