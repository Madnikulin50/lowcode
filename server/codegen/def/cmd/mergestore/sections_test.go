package main

import (
	"reflect"
	"testing"
)

func TestExtractParenBlock(t *testing.T) {
	src := "package foo\n\nimport (\n\t\"a\"\n\t\"b\"\n)\n\nrest\n"
	before, inner, after, err := extractParenBlock(src, "import (")
	if err != nil {
		t.Fatal(err)
	}
	if before != "package foo\n\nimport (\n" {
		t.Errorf("before = %q", before)
	}
	if inner != "\t\"a\"\n\t\"b\"" {
		t.Errorf("inner = %q", inner)
	}
	if after != ")\n\nrest\n" {
		t.Errorf("after = %q", after)
	}
}

func TestExtractParenBlockNotFound(t *testing.T) {
	if _, _, _, err := extractParenBlock("package foo\n", "import ("); err == nil {
		t.Fatal("expected an error when the opening line is missing")
	}
}

func TestSplitHeader(t *testing.T) {
	src := "package foo\n\nimport (\n\t\"a\"\n\n\t\"b\"\n)\n\nfunc f() {}\n"
	preamble, imports, rest, err := splitHeader(src)
	if err != nil {
		t.Fatal(err)
	}
	if preamble != "package foo\n\nimport (\n" {
		t.Errorf("preamble = %q", preamble)
	}
	wantImports := []string{"\"a\"", "\"b\""}
	if !reflect.DeepEqual(imports, wantImports) {
		t.Errorf("imports = %#v, want %#v", imports, wantImports)
	}
	if rest != "func f() {}\n" {
		t.Errorf("rest = %q", rest)
	}
}

func TestImportPath(t *testing.T) {
	cases := map[string]string{
		`"fmt"`:                 `"fmt"`,
		`sql "database/sql"`:    `"database/sql"`,
		`_ "some/blank/import"`: `"some/blank/import"`,
	}
	for in, want := range cases {
		if got := importPath(in); got != want {
			t.Errorf("importPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMergeImports(t *testing.T) {
	old := []string{`"a/b"`, `"c/d"`}
	fresh := []string{`"c/d/v2" "c/d"`, `"e/f"`}
	// fresh redeclares "c/d" with an alias - fresh must win on conflict.
	fresh = []string{`cd2 "c/d"`, `"e/f"`}

	got := mergeImports(old, fresh)
	want := []string{`"a/b"`, `cd2 "c/d"`, `"e/f"`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeImports = %#v, want %#v", got, want)
	}
}

func TestRenderImportBlock(t *testing.T) {
	got := renderImportBlock([]string{`"a"`, `"b"`})
	want := "\t\"a\"\n\t\"b\"\n)\n\n"
	if got != want {
		t.Errorf("renderImportBlock = %q, want %q", got, want)
	}
}
