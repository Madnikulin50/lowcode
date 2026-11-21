package rulesgo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type FilePersistence struct {
	path   string
	mu     sync.Mutex
	chains map[string]*Chain
}

func NewFilePersistence(path string) *FilePersistence {
	f := &FilePersistence{path: path, chains: make(map[string]*Chain)}
	_ = f.readLocked()
	return f
}

func (f *FilePersistence) readLocked() error {
	raw, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var list []*Chain
	if err := json.Unmarshal(raw, &list); err != nil {
		var wrap struct {
			Chains []*Chain `json:"chains"`
		}
		if err2 := json.Unmarshal(raw, &wrap); err2 != nil {
			return err
		}
		list = wrap.Chains
	}
	f.chains = make(map[string]*Chain, len(list))
	for _, c := range list {
		if c != nil && c.ID != "" {
			f.chains[c.ID] = c
		}
	}
	return nil
}

func (f *FilePersistence) flushLocked() error {
	if err := os.MkdirAll(filepath.Dir(f.path), 0o755); err != nil {
		return err
	}
	list := make([]*Chain, 0, len(f.chains))
	for _, c := range f.chains {
		list = append(list, c)
	}
	raw, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}

func (f *FilePersistence) LoadChains(ctx context.Context) ([]*Chain, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.readLocked(); err != nil {
		return nil, err
	}
	out := make([]*Chain, 0, len(f.chains))
	for _, c := range f.chains {
		out = append(out, c)
	}
	return out, nil
}

func (f *FilePersistence) SaveChain(ctx context.Context, chain *Chain) error {
	if chain == nil || chain.ID == "" {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.chains[chain.ID] = chain
	return f.flushLocked()
}

func (f *FilePersistence) UpdateChain(ctx context.Context, chain *Chain) error {
	return f.SaveChain(ctx, chain)
}

func (f *FilePersistence) DeleteChain(ctx context.Context, chainID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.chains, chainID)
	return f.flushLocked()
}
