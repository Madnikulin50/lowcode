package cmp

import (
	"sort"
	"strings"
	"unicode"
)

type opType int

const (
	opEqual opType = iota
	opDel          // only in a (reference)
	opAdd          // only in b (target)
)

type op struct {
	t    opType
	a, b int // indices into the sequences (a for equal/del, b for equal/add)
}

// dpLimit caps the O(n*m) table of the fallback LCS.
const dpLimit = 4_000_000

// diffKeys is a patience-style line diff: lines that occur exactly once on
// both sides anchor the alignment (longest increasing subsequence), the gaps
// between anchors are diffed recursively, and gaps without anchors fall
// back to a plain LCS (or, when huge, to delete+add).
func diffKeys(a, b []string) []op {
	var ops []op
	diffRange(a, b, 0, len(a), 0, len(b), &ops)
	return ops
}

func diffRange(a, b []string, a0, a1, b0, b1 int, out *[]op) {
	// trim common prefix/suffix
	for a0 < a1 && b0 < b1 && a[a0] == b[b0] {
		*out = append(*out, op{opEqual, a0, b0})
		a0++
		b0++
	}
	var tail []op
	for a0 < a1 && b0 < b1 && a[a1-1] == b[b1-1] {
		a1--
		b1--
		tail = append(tail, op{opEqual, a1, b1})
	}
	flushTail := func() {
		for i := len(tail) - 1; i >= 0; i-- {
			*out = append(*out, tail[i])
		}
	}
	defer flushTail()

	if a0 == a1 {
		for j := b0; j < b1; j++ {
			*out = append(*out, op{opAdd, a0, j})
		}
		return
	}
	if b0 == b1 {
		for i := a0; i < a1; i++ {
			*out = append(*out, op{opDel, i, b0})
		}
		return
	}

	anchors := uniqueAnchors(a, b, a0, a1, b0, b1)
	if len(anchors) == 0 {
		lcsRange(a, b, a0, a1, b0, b1, out)
		return
	}
	pa, pb := a0, b0
	for _, an := range anchors {
		diffRange(a, b, pa, an[0], pb, an[1], out)
		*out = append(*out, op{opEqual, an[0], an[1]})
		pa, pb = an[0]+1, an[1]+1
	}
	diffRange(a, b, pa, a1, pb, b1, out)
}

// uniqueAnchors returns (ai, bi) pairs of lines unique on both sides, as a
// longest increasing subsequence in b-order.
func uniqueAnchors(a, b []string, a0, a1, b0, b1 int) [][2]int {
	ca := map[string]int{}
	ia := map[string]int{}
	for i := a0; i < a1; i++ {
		ca[a[i]]++
		ia[a[i]] = i
	}
	cb := map[string]int{}
	ib := map[string]int{}
	for j := b0; j < b1; j++ {
		cb[b[j]]++
		ib[b[j]] = j
	}
	var pairs [][2]int
	for k, n := range ca {
		if n == 1 && cb[k] == 1 {
			pairs = append(pairs, [2]int{ia[k], ib[k]})
		}
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i][0] < pairs[j][0] })
	// LIS on pairs[i][1]
	n := len(pairs)
	if n == 0 {
		return nil
	}
	tails := make([]int, 0, n)
	prev := make([]int, n)
	for i := range pairs {
		v := pairs[i][1]
		pos := sort.Search(len(tails), func(k int) bool { return pairs[tails[k]][1] >= v })
		if pos > 0 {
			prev[i] = tails[pos-1]
		} else {
			prev[i] = -1
		}
		if pos == len(tails) {
			tails = append(tails, i)
		} else {
			tails[pos] = i
		}
	}
	res := make([][2]int, len(tails))
	k := tails[len(tails)-1]
	for i := len(tails) - 1; i >= 0; i-- {
		res[i] = pairs[k]
		k = prev[k]
	}
	return res
}

func lcsRange(a, b []string, a0, a1, b0, b1 int, out *[]op) {
	n, m := a1-a0, b1-b0
	if n*m > dpLimit {
		for i := a0; i < a1; i++ {
			*out = append(*out, op{opDel, i, b0})
		}
		for j := b0; j < b1; j++ {
			*out = append(*out, op{opAdd, a1, j})
		}
		return
	}
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[a0+i] == b[b0+j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[a0+i] == b[b0+j]:
			*out = append(*out, op{opEqual, a0 + i, b0 + j})
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			*out = append(*out, op{opDel, a0 + i, b0 + j})
			i++
		default:
			*out = append(*out, op{opAdd, a0 + i, b0 + j})
			j++
		}
	}
	for ; i < n; i++ {
		*out = append(*out, op{opDel, a0 + i, b0 + m})
	}
	for ; j < m; j++ {
		*out = append(*out, op{opAdd, a0 + n, b0 + j})
	}
}

// change is a run of non-equal ops resolved into paired replacements and
// leftovers.
type change struct {
	pairs   [][2]int // (a index, b index): same place, different wording
	delOnly []int
	addOnly []int
}

// groupChanges walks the ops and, inside every run of deletes/adds, pairs
// lines that are similar enough to be a replacement (monotonic greedy).
func groupChanges(ops []op, a, b []string) (equal [][2]int, changes []change) {
	for i := 0; i < len(ops); {
		if ops[i].t == opEqual {
			equal = append(equal, [2]int{ops[i].a, ops[i].b})
			i++
			continue
		}
		var dels, adds []int
		for i < len(ops) && ops[i].t != opEqual {
			if ops[i].t == opDel {
				dels = append(dels, ops[i].a)
			} else {
				adds = append(adds, ops[i].b)
			}
			i++
		}
		var c change
		used := make([]bool, len(adds))
		j0 := 0
		for _, di := range dels {
			best, bj := 0.0, -1
			for j := j0; j < len(adds); j++ {
				if used[j] {
					continue
				}
				if s := similarity(a[di], b[adds[j]]); s > best {
					best, bj = s, j
				}
			}
			if bj >= 0 && best >= pairThreshold {
				c.pairs = append(c.pairs, [2]int{di, adds[bj]})
				used[bj] = true
				j0 = bj + 1
			} else {
				c.delOnly = append(c.delOnly, di)
			}
		}
		for j, ai := range adds {
			if !used[j] {
				c.addOnly = append(c.addOnly, ai)
			}
		}
		changes = append(changes, c)
	}
	return
}

// pairThreshold: minimum token similarity for two lines to count as one
// line edited rather than one deleted and another added.
const pairThreshold = 0.5

func tokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '/' || r == '-')
	})
}

// similarity is the Dice coefficient over token multisets.
func similarity(x, y string) float64 {
	tx, ty := tokens(x), tokens(y)
	if len(tx) == 0 || len(ty) == 0 {
		return 0
	}
	bag := map[string]int{}
	for _, t := range tx {
		bag[t]++
	}
	common := 0
	for _, t := range ty {
		if bag[t] > 0 {
			bag[t]--
			common++
		}
	}
	return 2 * float64(common) / float64(len(tx)+len(ty))
}

// numbersOf lists numeric tokens (with optional decimal part) of a key.
func numbersOf(s string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, strings.TrimRight(cur.String(), "."))
			cur.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsDigit(r) || (r == '.' && cur.Len() > 0) {
			cur.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
