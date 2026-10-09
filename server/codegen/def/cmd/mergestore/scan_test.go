package main

import (
	"reflect"
	"testing"
)

func TestLineDepths(t *testing.T) {
	src := "a{\nb\n}c"
	// line 0 "a{" starts at depth 0, opens to depth 1
	// line 1 "b"  starts at depth 1
	// line 2 "}c" starts at depth 1, closes to depth 0
	want := []int{0, 1, 1}
	if got := lineDepths(src); !reflect.DeepEqual(got, want) {
		t.Errorf("lineDepths(%q) = %v, want %v", src, got, want)
	}
}

func TestLineDepthsIgnoresBracesInStringsAndComments(t *testing.T) {
	src := "func f() {\n" +
		"\ts := \"{ not a brace }\"\n" +
		"\t// } also not a brace\n" +
		"\treturn\n" +
		"}"
	depths := lineDepths(src)
	// the string literal and the line comment must not have affected depth:
	// every line from 1 through 3 should still read as depth 1, and the
	// final "}" should close back to 0.
	want := []int{0, 1, 1, 1, 1}
	if !reflect.DeepEqual(depths, want) {
		t.Errorf("lineDepths with braces in string/comment = %v, want %v", depths, want)
	}
}

func TestSplitTopLevelBlocks(t *testing.T) {
	src := "block one\nline two\n\nblock two\n\n\nblock three"
	got := splitTopLevelBlocks(src)
	want := []string{"block one\nline two", "block two", "block three"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitTopLevelBlocks = %#v, want %#v", got, want)
	}
}

func TestSplitTopLevelBlocksIgnoresBlankLinesInsideBraces(t *testing.T) {
	src := "func a() {\n\n\tx := 1\n}\n\nfunc b() {\n\ty := 2\n}"
	got := splitTopLevelBlocks(src)
	want := []string{"func a() {\n\n\tx := 1\n}", "func b() {\n\ty := 2\n}"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitTopLevelBlocks with blank line inside braces = %#v, want %#v", got, want)
	}
}
