package main

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// memBackend is a small in-memory Backend: the tasks (and their history) of the --fixtures mode live in it, so that
// the same code serves the tasks with and without a Compose. The remarks of an object are those of the fixture case:
// Search of the module of remarks is answered by remarksOf.
type memBackend struct {
	mu        sync.Mutex
	recs      map[string]map[string]idcheck.Record // module → id → record
	next      int
	remarksOf func(objectID string) []idcheck.Record
}

func newMemBackend() *memBackend { return &memBackend{recs: map[string]map[string]idcheck.Record{}} }

var memEq = regexp.MustCompile(`^(\w+) = '([^']*)'$`)

func (m *memBackend) GetRecord(_, module, id string) (*idcheck.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.recs[module][id]
	if !ok {
		return nil, fmt.Errorf("no %s/%s", module, id)
	}
	cp := idcheck.Record{ID: r.ID, Values: map[string][]string{}}
	for k, v := range r.Values {
		cp.Values[k] = append([]string(nil), v...)
	}
	return &cp, nil
}

func (m *memBackend) Search(_, module, filter string) ([]idcheck.Record, error) {
	if module == cmp.ModExpertise && m.remarksOf != nil {
		if q := memEq.FindStringSubmatch(filter); q != nil && q[1] == "object" {
			return m.remarksOf(q[2]), nil
		}
		return nil, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	q := memEq.FindStringSubmatch(filter)
	var out []idcheck.Record
	for _, r := range m.recs[module] {
		if q != nil && r.Get(q[1]) != q[2] {
			continue
		}
		cp := idcheck.Record{ID: r.ID, Values: map[string][]string{}}
		for k, v := range r.Values {
			cp.Values[k] = append([]string(nil), v...)
		}
		out = append(out, cp)
	}
	return out, nil
}

func (m *memBackend) Create(_, module string, v idcheck.Values) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	id := module + "-" + strconv.Itoa(m.next)
	if m.recs[module] == nil {
		m.recs[module] = map[string]idcheck.Record{}
	}
	rec := idcheck.Record{ID: id, Values: map[string][]string{}}
	for k, x := range v {
		if len(x) > 0 && x[0] != "" {
			rec.Values[k] = append([]string(nil), x...)
		}
	}
	m.recs[module][id] = rec
	return id, nil
}

// Update overlays the values; an explicit empty value clears a field (as in Compose).
func (m *memBackend) Update(_, module, id string, v idcheck.Values) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.recs[module][id]
	if !ok {
		return fmt.Errorf("no %s/%s", module, id)
	}
	for k, x := range v {
		if len(x) == 0 || x[0] == "" {
			delete(r.Values, k)
			continue
		}
		r.Values[k] = append([]string(nil), x...)
	}
	return nil
}

func (m *memBackend) Delete(_, module, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recs[module], id)
	return nil
}

func (m *memBackend) Download(_, att string) ([]byte, string, error) {
	return nil, "", fmt.Errorf("no files in the in-memory backend: %s", att)
}
