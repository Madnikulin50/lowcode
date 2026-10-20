package cmp

// Pair links one reference volume to the target volume(s) that cover it. A
// reference volume with several targets (one ПД том → several РД томов) is
// compared against the concatenation of its targets.
type Pair struct {
	Ref      *Volume
	Targets  []*Volume
	Fragment bool    // the targets cover only part of the reference volume
	Manual   bool    // made by hand, not by matching
	Score    float64 // best match score
}

// Matching is the outcome of Match.
type Matching struct {
	Pairs      []*Pair
	RefOnly    []*Volume // reference volumes with no counterpart (maybe not released yet)
	TargetOnly []*Volume // target volumes with no reference counterpart
}

const (
	matchThreshold = 0.30 // min containment to consider two volumes related
	fragmentBelow  = 0.60 // reference coverage below this ⇒ fragment pair
)

func keySet(lines []Line) map[string]struct{} {
	s := make(map[string]struct{}, len(lines))
	for _, l := range lines {
		s[l.Key] = struct{}{}
	}
	return s
}

// containment is |A∩B| / min(|A|,|B|) over distinct line keys.
func containment(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	small, big := a, b
	if len(b) < len(a) {
		small, big = b, a
	}
	n := 0
	for k := range small {
		if _, ok := big[k]; ok {
			n++
		}
	}
	return float64(n) / float64(len(small))
}

func volScore(r, t *Volume, rs, ts map[string]struct{}) float64 {
	s := containment(rs, ts)
	if r.Code != "" && r.Code == t.Code {
		s += 0.5
	} else if r.Code != "" && t.Code != "" && Series(r.Code) != "" && Series(r.Code) == Series(t.Code) {
		s += 0.1
	}
	return s
}

// Match pairs reference and target volumes. Every target goes to its best
// reference volume (if related enough); a reference volume may collect
// several targets. Whatever is left unpaired is reported separately — it is
// the caller's job to treat that as "not released / out of scope" rather
// than as an error.
func Match(refs, targets []*Volume) Matching {
	// One volume against one volume is the user's own pairing (a file against
	// its next edition, a ПД volume against its РД): it needs no content
	// overlap to be accepted. Little overlap simply shows as a low similarity.
	if len(refs) == 1 && len(targets) == 1 {
		r, t := refs[0], targets[0]
		rs, ts := keySet(r.Lines), keySet(t.Lines)
		cov := 1.0
		if len(rs) > 0 {
			cov = float64(len(rs)-missing(rs, ts)) / float64(len(rs))
		}
		return Matching{Pairs: []*Pair{{Ref: r, Targets: []*Volume{t}, Fragment: cov < fragmentBelow, Score: volScore(r, t, rs, ts)}}}
	}
	rsets := make([]map[string]struct{}, len(refs))
	for i, r := range refs {
		rsets[i] = keySet(r.Lines)
	}
	byRef := map[int][]*Volume{}
	score := map[int]float64{}
	var m Matching
	for _, t := range targets {
		ts := keySet(t.Lines)
		best, bi := 0.0, -1
		for i, r := range refs {
			if s := volScore(r, t, rsets[i], ts); s > best {
				best, bi = s, i
			}
		}
		// a bare code/series hit without shared content is not a match
		bare := bi >= 0 && len(rsets[bi]) == 0 && len(ts) == 0 && refs[bi].Code != "" && refs[bi].Code == t.Code
		if bi < 0 || (!bare && containment(rsets[bi], ts) < matchThreshold) {
			m.TargetOnly = append(m.TargetOnly, t)
			continue
		}
		byRef[bi] = append(byRef[bi], t)
		if best > score[bi] {
			score[bi] = best
		}
	}
	for i, r := range refs {
		ts, ok := byRef[i]
		if !ok {
			m.RefOnly = append(m.RefOnly, r)
			continue
		}
		union := map[string]struct{}{}
		for _, t := range ts {
			for k := range keySet(t.Lines) {
				union[k] = struct{}{}
			}
		}
		covered := 0
		for k := range rsets[i] {
			if _, ok := union[k]; ok {
				covered++
			}
		}
		cov := 1.0
		if len(rsets[i]) > 0 {
			cov = float64(covered) / float64(len(rsets[i]))
		}
		m.Pairs = append(m.Pairs, &Pair{Ref: r, Targets: ts, Fragment: cov < fragmentBelow, Score: score[i]})
	}
	return m
}

// missing counts keys of a that are absent from b.
func missing(a, b map[string]struct{}) int {
	n := 0
	for k := range a {
		if _, ok := b[k]; !ok {
			n++
		}
	}
	return n
}
