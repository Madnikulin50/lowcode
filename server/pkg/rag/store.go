package rag

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketDocuments    = []byte("documents")
	bucketChunks       = []byte("chunks")
	bucketVectors      = []byte("vectors")
	bucketMeta         = []byte("meta")
	bucketPagesMeta    = []byte("pages_meta")
	bucketPagesChunks  = []byte("pages_chunks")
	bucketPagesVectors = []byte("pages_vectors")
)

type Document struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt int64  `json:"createdAt"`
}

type Chunk struct {
	ID         string    `json:"id"`
	DocID      string    `json:"docID"`
	Text       string    `json:"text"`
	Embedding  []float32 `json:"-"`
	ChunkIndex int       `json:"chunkIndex"`
}

type Store struct {
	db   *bolt.DB
	mu   sync.RWMutex
	path string
}

func NewStore(path string) (*Store, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("rag store: mkdir %s: %w", dir, err)
	}
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("rag store: open %s: %w", path, err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{bucketDocuments, bucketChunks, bucketVectors, bucketMeta, bucketPagesMeta, bucketPagesChunks, bucketPagesVectors} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &Store{db: db, path: path}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) PutDocument(doc Document) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		nsb, err := tx.Bucket(bucketDocuments).CreateBucketIfNotExists([]byte(doc.Namespace))
		if err != nil {
			return err
		}
		return nsb.Put([]byte(doc.ID), data)
	})
}

func (s *Store) GetDocument(namespace, id string) (*Document, error) {
	var doc Document
	err := s.db.View(func(tx *bolt.Tx) error {
		nsb := tx.Bucket(bucketDocuments).Bucket([]byte(namespace))
		if nsb == nil {
			return fmt.Errorf("namespace %s not found", namespace)
		}
		data := nsb.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("document %s not found", id)
		}
		return json.Unmarshal(data, &doc)
	})
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (s *Store) ListDocuments(namespace string) ([]Document, error) {
	var docs []Document
	err := s.db.View(func(tx *bolt.Tx) error {
		nsb := tx.Bucket(bucketDocuments).Bucket([]byte(namespace))
		if nsb == nil {
			return nil
		}
		return nsb.ForEach(func(k, v []byte) error {
			var d Document
			if err := json.Unmarshal(v, &d); err != nil {
				return err
			}
			docs = append(docs, d)
			return nil
		})
	})
	return docs, err
}

func (s *Store) DeleteDocument(namespace, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		nsb := tx.Bucket(bucketDocuments).Bucket([]byte(namespace))
		if nsb != nil {
			nsb.Delete([]byte(id))
		}
		chb := tx.Bucket(bucketChunks).Bucket([]byte(namespace))
		if chb != nil {
			prefix := []byte(id + ":")
			c := chb.Cursor()
			for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
				chb.Delete(k)
			}
		}
		vecb := tx.Bucket(bucketVectors).Bucket([]byte(namespace))
		if vecb != nil {
			prefix := []byte(id + ":")
			c := vecb.Cursor()
			for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
				vecb.Delete(k)
			}
		}
		return nil
	})
}

func (s *Store) PutChunk(namespace string, chunk Chunk) error {
	embData := floatsToBytes(chunk.Embedding)
	chunkData, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		chb, err := tx.Bucket(bucketChunks).CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		key := []byte(chunk.DocID + ":" + chunk.ID)
		if err := chb.Put(key, chunkData); err != nil {
			return err
		}
		vecb, err := tx.Bucket(bucketVectors).CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return err
		}
		return vecb.Put(key, embData)
	})
}

func (s *Store) GetChunks(namespace, docID string) ([]Chunk, error) {
	var chunks []Chunk
	err := s.db.View(func(tx *bolt.Tx) error {
		chb := tx.Bucket(bucketChunks).Bucket([]byte(namespace))
		if chb == nil {
			return nil
		}
		vecb := tx.Bucket(bucketVectors).Bucket([]byte(namespace))
		prefix := []byte(docID + ":")
		c := chb.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var ch Chunk
			if err := json.Unmarshal(v, &ch); err != nil {
				return err
			}
			if vecb != nil {
				embData := vecb.Get(k)
				if embData != nil {
					ch.Embedding = bytesToFloats(embData)
				}
			}
			chunks = append(chunks, ch)
		}
		return nil
	})
	return chunks, err
}

type SearchResult struct {
	Chunk Chunk   `json:"chunk"`
	Score float64 `json:"score"`
}

func (s *Store) Search(namespace string, embedding []float32, topK int) ([]SearchResult, error) {
	var results []SearchResult
	err := s.db.View(func(tx *bolt.Tx) error {
		vecb := tx.Bucket(bucketVectors).Bucket([]byte(namespace))
		if vecb == nil {
			return nil
		}
		chb := tx.Bucket(bucketChunks).Bucket([]byte(namespace))
		if chb == nil {
			return nil
		}
		var scoredList []scoredChunk
		c := vecb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			vec := bytesToFloats(v)
			if len(vec) != len(embedding) {
				continue
			}
			sim := cosineSimilarity(embedding, vec)
			scoredList = append(scoredList, scoredChunk{key: k, score: sim})
		}
		// Keep top-K
		if len(scoredList) > topK {
			scoredList = topKResults(scoredList, topK)
		}
		for _, sc := range scoredList {
			chunkData := chb.Get(sc.key)
			if chunkData == nil {
				continue
			}
			var ch Chunk
			if err := json.Unmarshal(chunkData, &ch); err != nil {
				continue
			}
			results = append(results, SearchResult{Chunk: ch, Score: sc.score})
		}
		return nil
	})
	return results, err
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

type scoredChunk struct {
	key   []byte
	score float64
}

func topKResults(s []scoredChunk, k int) []scoredChunk {
	for i := 0; i < k && i < len(s); i++ {
		best := i
		for j := i + 1; j < len(s); j++ {
			if s[j].score > s[best].score {
				best = j
			}
		}
		s[i], s[best] = s[best], s[i]
	}
	if len(s) > k {
		s = s[:k]
	}
	return s
}

func floatsToBytes(f []float32) []byte {
	buf := new(bytes.Buffer)
	for _, v := range f {
		binary.Write(buf, binary.LittleEndian, math.Float32bits(v))
	}
	return buf.Bytes()
}

func bytesToFloats(data []byte) []float32 {
	f := make([]float32, len(data)/4)
	buf := bytes.NewReader(data)
	for i := range f {
		var bits uint32
		binary.Read(buf, binary.LittleEndian, &bits)
		f[i] = math.Float32frombits(bits)
	}
	return f
}

type PageChunk struct {
	ID          string `json:"id"`
	PageID      uint64 `json:"pageID"`
	NamespaceID uint64 `json:"namespaceID"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	ChunkIndex  int    `json:"chunkIndex"`
}

func (s *Store) PutPageChunk(pc PageChunk, embedding []float32) error {
	embData := floatsToBytes(embedding)
	chunkData, err := json.Marshal(pc)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		key := fmt.Sprintf("%d:%s", pc.PageID, pc.ID)
		if err := tx.Bucket(bucketPagesChunks).Put([]byte(key), chunkData); err != nil {
			return err
		}
		return tx.Bucket(bucketPagesVectors).Put([]byte(key), embData)
	})
}

func (s *Store) ClearPages() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(bucketPagesChunks); err != nil && err != bolt.ErrBucketNotFound {
			return err
		}
		if err := tx.DeleteBucket(bucketPagesVectors); err != nil && err != bolt.ErrBucketNotFound {
			return err
		}
		if _, err := tx.CreateBucket(bucketPagesChunks); err != nil {
			return err
		}
		if _, err := tx.CreateBucket(bucketPagesVectors); err != nil {
			return err
		}
		return tx.Bucket(bucketPagesMeta).Put([]byte("last_crawl"), []byte{})
	})
}

func (s *Store) ListPageChunks() ([]PageChunk, error) {
	var chunks []PageChunk
	err := s.db.View(func(tx *bolt.Tx) error {
		chb := tx.Bucket(bucketPagesChunks)
		if chb == nil {
			return nil
		}
		return chb.ForEach(func(k, v []byte) error {
			var pc PageChunk
			if err := json.Unmarshal(v, &pc); err != nil {
				return err
			}
			chunks = append(chunks, pc)
			return nil
		})
	})
	return chunks, err
}

func (s *Store) SearchPages(embedding []float32, topK int) ([]SearchResult, error) {
	var results []SearchResult
	err := s.db.View(func(tx *bolt.Tx) error {
		vecb := tx.Bucket(bucketPagesVectors)
		if vecb == nil {
			return nil
		}
		chb := tx.Bucket(bucketPagesChunks)
		if chb == nil {
			return nil
		}
		var scoredList []scoredChunk
		c := vecb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			vec := bytesToFloats(v)
			if len(vec) != len(embedding) {
				continue
			}
			sim := cosineSimilarity(embedding, vec)
			scoredList = append(scoredList, scoredChunk{key: k, score: sim})
		}
		if len(scoredList) > topK {
			scoredList = topKResults(scoredList, topK)
		}
		for _, sc := range scoredList {
			data := chb.Get(sc.key)
			if data == nil {
				continue
			}
			var pc PageChunk
			if err := json.Unmarshal(data, &pc); err != nil {
				continue
			}
			results = append(results, SearchResult{Chunk: Chunk{ID: pc.ID, DocID: fmt.Sprint(pc.PageID), Text: pc.Text}, Score: sc.score})
		}
		return nil
	})
	return results, err
}

func (s *Store) SetPagesCrawlTime(t int64) error {
	data := fmt.Sprint(t)
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPagesMeta).Put([]byte("last_crawl"), []byte(data))
	})
}

func (s *Store) GetPagesCrawlTime() (int64, error) {
	var t int64
	err := s.db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket(bucketPagesMeta).Get([]byte("last_crawl"))
		if data == nil {
			return nil
		}
		var i int
		if _, err := fmt.Sscanf(string(data), "%d", &i); err != nil {
			return err
		}
		t = int64(i)
		return nil
	})
	return t, err
}
