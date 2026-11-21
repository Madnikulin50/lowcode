package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIdentFromPlural(t *testing.T) {
	cases := map[string]string{
		"Nodes":               "node",
		"ComposeModules":      "composeModule",
		"ComposeModuleFields": "composeModuleField",
		"Ns":                  "n",
	}
	for in, want := range cases {
		if got := identFromPlural(in); got != want {
			t.Errorf("identFromPlural(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestIdentFromPluralSortOrderDiffersFromPluralSortOrder is a regression
// test for the bug found this session: mergeInterfacesFile originally
// sorted blocks by the exported PLURAL name, but the templates actually
// sort by the singular ident (res.store.ident). These two orderings are
// not the same in general - this pair is the concrete case that exposed it.
func TestIdentFromPluralSortOrderDiffersFromPluralSortOrder(t *testing.T) {
	a, b := "ComposeModuleFields", "ComposeModules"
	if !(a < b) {
		t.Fatalf("test fixture assumption broken: %q is not < %q as plain strings", a, b)
	}

	identA, identB := identFromPlural(a), identFromPlural(b)
	if !(identB < identA) {
		t.Fatalf(
			"expected the ident ordering to be the REVERSE of the plural-string ordering "+
				"(that's the whole point of the bug) - got identFromPlural(%q)=%q, identFromPlural(%q)=%q",
			a, identA, b, identB,
		)
	}
}

func TestGroupAdjacent(t *testing.T) {
	in := []keyedBlock{
		{key: "a", text: "a1"},
		{key: "a", text: "a2"},
		{key: "b", text: "b1"},
		{key: "a", text: "a3"}, // non-adjacent repeat of "a" - stays separate
	}
	got := groupAdjacent(in)
	want := []keyedBlock{
		{key: "a", text: "a1\n\na2"},
		{key: "b", text: "b1"},
		{key: "a", text: "a3"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("groupAdjacent = %#v, want %#v", got, want)
	}
}

func TestMergeKeyed(t *testing.T) {
	old := []keyedBlock{{key: "b", text: "old-b"}, {key: "a", text: "old-a"}}
	fresh := []keyedBlock{{key: "c", text: "fresh-c"}, {key: "a", text: "fresh-a"}}

	got := mergeKeyed(old, fresh)
	// sorted by key, fresh wins on conflict (key "a")
	want := []keyedBlock{
		{key: "a", text: "fresh-a"},
		{key: "b", text: "old-b"},
		{key: "c", text: "fresh-c"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeKeyed = %#v, want %#v", got, want)
	}
}

// --- fixture-based tests for the file-level merge functions -----------------
//
// Each fixture models exactly two "resources": A, present identically in
// both old and fresh (must survive untouched - this is what protects
// not-yet-ported components like system/discovery from being clobbered),
// and B, present in both but with different content in fresh (must be
// replaced by the fresh version - this is what lets a ported component's
// schema changes actually take effect).

func TestMergeInterfacesFile(t *testing.T) {
	mk := func(bVersion string) string {
		return "package store\n\n" +
			"import (\n\t\"context\"\n)\n\n" +
			"type (\n" +
			"\tStorer interface {\n" +
			"\t\tHealthcheck(context.Context) error\n\n" +
			"\t\tAStorer\n" +
			"\t\tBStorer\n" +
			"\t}\n\n" +
			"\tAStorer interface {\n" +
			"\t\tLookupAByID(ctx context.Context, id uint64) (*types.A, error)\n" +
			"\t}\n\n" +
			"\tBStorer interface {\n" +
			"\t\tLookupBBy" + bVersion + "(ctx context.Context, id uint64) (*types.B, error)\n" +
			"\t}\n" +
			")\n\n" +
			"func LookupAByID(ctx context.Context, s AStorer, id uint64) (*types.A, error) {\n" +
			"\treturn s.LookupAByID(ctx, id)\n" +
			"}\n\n" +
			"func LookupBBy" + bVersion + "(ctx context.Context, s BStorer, id uint64) (*types.B, error) {\n" +
			"\treturn s.LookupBBy" + bVersion + "(ctx, id)\n" +
			"}\n"
	}

	old := mk("ID")
	fresh := mk("Handle")

	got, err := mergeInterfacesFile(old, fresh)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(got, "LookupAByID(ctx context.Context, id uint64) (*types.A, error)") {
		t.Errorf("A's untouched interface method did not survive the merge:\n%s", got)
	}
	if contains(got, "LookupBByID") {
		t.Errorf("B's stale (old) method leaked into the merge output:\n%s", got)
	}
	if !contains(got, "LookupBByHandle") {
		t.Errorf("B's fresh method is missing from the merge output:\n%s", got)
	}
}

func TestMergeAuxTypesFile(t *testing.T) {
	mk := func(bField string) string {
		return "package rdbms\n\n" +
			"import (\n\t\"time\"\n)\n\n" +
			"type (\n" +
			"\tA struct {\n" +
			"\t\tID uint64 `db:\"id\"`\n" +
			"\t}\n\n" +
			"\tB struct {\n" +
			"\t\t" + bField + " string `db:\"" + bField + "\"`\n" +
			"\t}\n" +
			")\n\n" +
			"func (aux *A) decode() {}\n\n" +
			"func (aux *B) decode() {}\n" +
			"var _ = time.Now\n"
	}

	old := mk("Handle")
	fresh := mk("Name")

	got, err := mergeAuxTypesFile(old, fresh)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(got, `ID uint64 `+"`db:\"id\"`") {
		t.Errorf("A's untouched struct did not survive the merge:\n%s", got)
	}
	if contains(got, "Handle string") {
		t.Errorf("B's stale (old) field leaked into the merge output:\n%s", got)
	}
	if !contains(got, "Name string") {
		t.Errorf("B's fresh field is missing from the merge output:\n%s", got)
	}
}

func TestMergeQueriesFile(t *testing.T) {
	mk := func(bCol string) string {
		return "package rdbms\n\n" +
			"import (\n\t\"github.com/doug-martin/goqu/v9\"\n)\n\n" +
			"var (\n" +
			"\taTable = goqu.T(\"a\")\n" +
			"\taSelectQuery = goqu.From(aTable)\n\n" +
			"\tbTable = goqu.T(\"b\")\n" +
			"\tbSelectQuery = goqu.From(bTable).Select(\"" + bCol + "\")\n" +
			")\n"
	}

	old := mk("old_col")
	fresh := mk("new_col")

	got, err := mergeQueriesFile(old, fresh)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(got, `aTable = goqu.T("a")`) {
		t.Errorf("A's untouched query vars did not survive the merge:\n%s", got)
	}
	if contains(got, "old_col") {
		t.Errorf("B's stale (old) query leaked into the merge output:\n%s", got)
	}
	if !contains(got, "new_col") {
		t.Errorf("B's fresh query is missing from the merge output:\n%s", got)
	}
}

func TestMergeFiltersFile(t *testing.T) {
	mk := func(bBody string) string {
		return "package rdbms\n\n" +
			"import (\n\t\"context\"\n)\n\n" +
			"type (\n" +
			"\textendedFilters struct {\n" +
			"\t\t// Filter extensions for search/query functions\n" +
			"\t\ta func(context.Context) error\n\n" +
			"\t\tb func(context.Context) error\n" +
			"\t}\n" +
			")\n\n" +
			"func AFilter(ctx context.Context) error { return nil }\n\n" +
			"func BFilter(ctx context.Context) error { return " + bBody + " }\n"
	}

	old := mk("errOld")
	fresh := mk("errFresh")

	got, err := mergeFiltersFile(old, fresh)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(got, "a func(context.Context) error") {
		t.Errorf("A's untouched field did not survive the merge:\n%s", got)
	}
	if contains(got, "errOld") {
		t.Errorf("B's stale (old) func body leaked into the merge output:\n%s", got)
	}
	if !contains(got, "errFresh") {
		t.Errorf("B's fresh func body is missing from the merge output:\n%s", got)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

// --- self-merge idempotence against the real, current store files ----------
//
// Merging a file against an identical copy of itself must be a no-op. This
// exercises the full parse/merge/render path against real, full-scale
// generated files (not just the small hand-written fixtures above), without
// needing to hand-maintain large fixtures.

func TestSelfMergeIsNoopOnRealFiles(t *testing.T) {
	root := repoRootForTest(t)

	cases := []struct {
		path string
		fn   func(old, fresh string) (string, error)
	}{
		{"store/interfaces.gen.go", mergeInterfacesFile},
		{"store/adapters/rdbms/aux_types.gen.go", mergeAuxTypesFile},
		{"store/adapters/rdbms/queries.gen.go", mergeQueriesFile},
		{"store/adapters/rdbms/filters.gen.go", mergeFiltersFile},
	}

	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(root, c.path))
			if err != nil {
				t.Fatalf("reading %s: %v", c.path, err)
			}
			src := string(b)

			got, err := c.fn(src, src)
			if err != nil {
				t.Fatalf("self-merge failed: %v", err)
			}

			// formatting is applied by main.go after merging, not by the
			// merge functions themselves, so compare the merge functions'
			// own idempotence directly (gofmt-insensitive differences, if
			// any, are covered separately by the real `make codegen` run).
			if got != src {
				dbg := filepath.Join(os.TempDir(), filepath.Base(c.path)+".selfmerge")
				_ = os.WriteFile(dbg, []byte(got), 0o644)
				t.Errorf("self-merge of %s is not a no-op (diff length old=%d new=%d); wrote merged output to %s", c.path, len(src), len(got), dbg)
			}
		})
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// this package lives at server/codegen/def/cmd/mergestore
	return filepath.Join(wd, "..", "..", "..", "..")
}
