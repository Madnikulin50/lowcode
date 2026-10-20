package cmp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Ручное сопоставление: специалист указывает, какой фрагмент (диапазон страниц) какого файла
// эталона сравнивать с каким фрагментом (фрагментами) проверяемой стороны. Нужно, когда автоподбор
// ошибся — например, том ПД с несколькими сооружениями и том РД только по одному из них.

// Slice is a range of pages of one document (1-based, inclusive; To == 0 means "to the end").
type Slice struct {
	Doc  string `json:"doc"`
	From int    `json:"from"`
	To   int    `json:"to"`
}

// ManualPair links a slice of the reference side with slices of the target side.
type ManualPair struct {
	Ref     Slice   `json:"ref"`
	Targets []Slice `json:"targets"`
}

// ParsePairing reads the stored JSON ("" is no manual pairing).
func ParsePairing(s string) ([]ManualPair, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []ManualPair
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, fmt.Errorf("ручное сопоставление: %w", err)
	}
	return out, nil
}

// PairingJSON is the stored form.
func PairingJSON(p []ManualPair) string {
	if len(p) == 0 {
		return ""
	}
	b, _ := json.Marshal(p)
	return string(b)
}

// PairsState is what the pairing panel shows: the volumes found in every document, which were
// compared with which and what was left without a counterpart.
type PairsState struct {
	Volumes    map[string][]VolumeInfoJSON `json:"volumes"`
	Pairs      []PairInfoJSON              `json:"pairs"`
	RefOnly    []VolumeInfoJSON            `json:"refOnly"`
	TargetOnly []VolumeInfoJSON            `json:"targetOnly"`
	Params     []ParamRow                  `json:"params,omitempty"` // key parameters per volume and element
}

// VolumeInfoJSON is VolumeInfo with the field names of the stored/wire form.
type VolumeInfoJSON struct {
	Doc   string `json:"doc"`
	Index int    `json:"index"`
	Code  string `json:"code"`
	Title string `json:"title"`
	From  int    `json:"from"`
	To    int    `json:"to"`
}

// PairInfoJSON is PairInfo in the wire form.
type PairInfoJSON struct {
	Ref      VolumeInfoJSON   `json:"ref"`
	Targets  []VolumeInfoJSON `json:"targets"`
	Fragment bool             `json:"fragment"`
	Manual   bool             `json:"manual"`
	Score    float64          `json:"score"`
}

func volJSON(v VolumeInfo) VolumeInfoJSON {
	return VolumeInfoJSON{Doc: v.DocID, Index: v.Index, Code: v.Code, Title: v.Title, From: v.From, To: v.To}
}

func volsJSON(vs []VolumeInfo) []VolumeInfoJSON {
	out := make([]VolumeInfoJSON, 0, len(vs))
	for _, v := range vs {
		out = append(out, volJSON(v))
	}
	return out
}

// State arranges the result for the pairing panel.
func (r *Result) State() PairsState {
	st := PairsState{Volumes: map[string][]VolumeInfoJSON{}, Pairs: []PairInfoJSON{}, RefOnly: volsJSON(r.RefOnly), TargetOnly: volsJSON(r.TargetOnly)}
	st.Params = r.Params
	for id, vs := range r.Volumes {
		st.Volumes[id] = volsJSON(vs)
	}
	for _, p := range r.Pairs {
		st.Pairs = append(st.Pairs, PairInfoJSON{Ref: volJSON(p.Ref), Targets: volsJSON(p.Targets), Fragment: p.Fragment, Manual: p.Manual, Score: p.Score})
	}
	return st
}

// StateJSON is the form stored in the comparison record.
func (r *Result) StateJSON() string {
	b, _ := json.Marshal(r.State())
	return string(b)
}

type pageRange struct{ from, to int }

// subtract removes the taken ranges from [from, to] and returns what is left.
func subtract(from, to int, taken []pageRange) []pageRange {
	left := []pageRange{{from, to}}
	for _, t := range taken {
		var next []pageRange
		for _, l := range left {
			if t.to < l.from || t.from > l.to {
				next = append(next, l)
				continue
			}
			if t.from > l.from {
				next = append(next, pageRange{l.from, t.from - 1})
			}
			if t.to < l.to {
				next = append(next, pageRange{t.to + 1, l.to})
			}
		}
		left = next
	}
	return left
}

// sliceVolume builds a volume from a range of pages; ok is false if the range does not fit the document.
func sliceVolume(d *Document, s Slice, idx int) (*Volume, string) {
	n := len(d.Pages)
	from, to := s.From, s.To
	if from < 1 {
		from = 1
	}
	if to <= 0 || to > n {
		to = n
	}
	if from > to || from > n {
		return nil, fmt.Sprintf("«%s»: страницы %d–%d вне документа (%d стр.)", d.Name, s.From, s.To, n)
	}
	base := strings.TrimSuffix(d.Name, ".pdf")
	title := pageTitle(d.Pages[from-1])
	if title == "" {
		title = base
	}
	title = fmt.Sprintf("%s, стр. %d–%d", title, from, to)
	return buildVolume(d, idx, "", title, from, to, base), ""
}

// applyPairing makes the manual pairs and leaves the rest of the volumes to automatic matching: a
// volume the manual slices overlap is cut down to the pages they do not take.
func applyPairing(manual []ManualPair, docs map[string]*Document, side map[string]string, refVols, tgtVols []*Volume) (forced []*Pair, refRest, tgtRest []*Volume, notes []string) {
	taken := map[string][]pageRange{}
	next := 1000
	mk := func(s Slice, want string) *Volume {
		d := docs[s.Doc]
		if d == nil {
			notes = append(notes, fmt.Sprintf("Ручное сопоставление: файл %q не найден среди документов — пара пропущена.", s.Doc))
			return nil
		}
		if side[s.Doc] != want {
			notes = append(notes, fmt.Sprintf("Ручное сопоставление: файл «%s» лежит не на той стороне — пара пропущена.", d.Name))
			return nil
		}
		next++
		v, why := sliceVolume(d, s, next)
		if v == nil {
			notes = append(notes, "Ручное сопоставление: "+why+" — пара пропущена.")
			return nil
		}
		return v
	}
	for _, mp := range manual {
		rv := mk(mp.Ref, SideRef)
		if rv == nil {
			continue
		}
		var ts []*Volume
		for _, t := range mp.Targets {
			if tv := mk(t, SideTarget); tv != nil {
				ts = append(ts, tv)
			}
		}
		if len(ts) == 0 {
			continue
		}
		forced = append(forced, &Pair{Ref: rv, Targets: ts, Manual: true, Score: 1})
		taken[rv.DocID] = append(taken[rv.DocID], pageRange{rv.FirstPage, rv.LastPage})
		for _, t := range ts {
			taken[t.DocID] = append(taken[t.DocID], pageRange{t.FirstPage, t.LastPage})
		}
	}
	trim := func(vs []*Volume) []*Volume {
		var out []*Volume
		for _, v := range vs {
			tk := taken[v.DocID]
			if len(tk) == 0 {
				out = append(out, v)
				continue
			}
			d := docs[v.DocID]
			for _, r := range subtract(v.FirstPage, v.LastPage, tk) {
				if r.from == v.FirstPage && r.to == v.LastPage {
					out = append(out, v)
					continue
				}
				out = append(out, buildVolume(d, v.Index, v.RawCode, v.Title, r.from, r.to, strings.TrimSuffix(d.Name, ".pdf")))
			}
		}
		sort.SliceStable(out, func(i, j int) bool {
			if out[i].DocID != out[j].DocID {
				return false
			}
			return out[i].FirstPage < out[j].FirstPage
		})
		return out
	}
	return forced, trim(refVols), trim(tgtVols), notes
}
