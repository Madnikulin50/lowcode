package cmp

import (
	"strings"
	"testing"
)

func TestRegionsAreValidated(t *testing.T) {
	ok := `[{"side":"ref","doc":"d","page":2,"box":"0.1,0.1,0.5,0.4"},{"box":"0,0,1,0.1"}]`
	if rs, err := ParseRegions(ok); err != nil || len(rs) != 2 {
		t.Fatalf("%v %v", rs, err)
	}
	if rs, err := ParseRegions(""); rs != nil || err != nil {
		t.Errorf("no regions: %v %v", rs, err)
	}
	for _, bad := range []string{
		`[{"box":"0.5,0.5,0.1,0.9"}]`, // upside down
		`[{"box":"0,0,2,1"}]`,         // outside the page
		`[{"box":"nonsense"}]`,
		`[{"side":"left","box":"0,0,1,1"}]`,
		`[{"page":-1,"box":"0,0,1,1"}]`,
		`not json`,
	} {
		if _, err := ParseRegions(bad); err == nil {
			t.Errorf("accepted: %s", bad)
		}
	}
	if RegionsJSON(nil) != "" {
		t.Error("no regions, no value")
	}
}

func TestRegionAppliesToItsSideDocAndPage(t *testing.T) {
	r := Region{Side: SideRef, Doc: "pd", Page: 3, Box: "0,0,0.5,0.5"}
	if !r.appliesTo(SideRef, "pd", 3) || r.appliesTo(SideTarget, "pd", 3) || r.appliesTo(SideRef, "other", 3) || r.appliesTo(SideRef, "pd", 4) {
		t.Error("side, document and page must all match")
	}
	any := Region{Box: "0,0,1,1"}
	if !any.appliesTo(SideTarget, "x", 99) {
		t.Error("an empty side, document and page match everything")
	}
	b := Box{0.1, 0.1, 0.2, 0.2}
	if !excluded([]Region{any}, SideRef, "d", 1, b) || excluded([]Region{{Box: "0.5,0.5,1,1"}}, SideRef, "d", 1, b) {
		t.Error("a box is excluded when its center is in a region")
	}
}

func TestWithoutRegionsDoesNotTouchTheOriginalsOrTheTitleBlock(t *testing.T) {
	d := fmtDoc("d", SideRef, fmtPage{1, "", []string{"Строка вне области", "Строка в области"}}, fmtPage{2, "", []string{"Другая страница"}})
	// the second body line of page 1 sits at y≈0.13; cut the band 0.12..0.16, and the title block band as well
	rs := []Region{{Page: 1, Box: "0,0.12,1,0.16"}, {Page: 1, Box: "0,0.8,1,1"}}
	out, dropped := withoutRegions([]*Document{d}, SideRef, rs)
	if dropped != 1 {
		t.Fatalf("one body line is cut, the title block stays: %d", dropped)
	}
	count := func(doc *Document, pn int, sub string) int {
		n := 0
		for _, l := range doc.Pages[pn-1].Lines {
			if strings.Contains(l.Text, sub) {
				n++
			}
		}
		return n
	}
	if count(out[0], 1, "в области") != 0 || count(out[0], 1, "вне области") != 1 || count(out[0], 2, "Другая") != 1 {
		t.Errorf("the cut: %+v", out[0].Pages[0].Lines)
	}
	if count(d, 1, "в области") != 1 {
		t.Error("the document given must stay as it was")
	}
	stamps := 0
	for _, l := range out[0].Pages[0].Lines {
		if l.Stamp {
			stamps++
		}
	}
	if stamps == 0 {
		t.Error("title block lines are kept: the volume detection reads them")
	}
	if same, n := withoutRegions([]*Document{d}, SideRef, nil); n != 0 || same[0] != d {
		t.Error("no regions, nothing copied")
	}
}

func TestCompareLeavesAnExcludedAreaOut(t *testing.T) {
	mk := func(id, side, depth string) *Document {
		pages := five("а")
		pages[1].body = append([]string{"Глубина заложения фундамента " + depth + " м."}, pages[1].body...)
		return fmtDoc(id, side, pages...)
	}
	pd, rd := mk("pd", SideRef, "2,4"), mk("rd", SideTarget, "2,6")
	has := func(res *Result) bool {
		for _, f := range res.Findings {
			if strings.Contains(f.Description, "Глубина") {
				return true
			}
		}
		return false
	}
	base := Compare([]*Document{pd}, []*Document{rd}, Options{NoFormat: true, NoCross: true})
	if !has(base) {
		t.Fatalf("without an exclusion the changed number is found: %+v", base.Findings)
	}
	// the first body line of page 2 is at y≈0.1: exclude the top band on that page of both sides
	cut := Compare([]*Document{pd}, []*Document{rd}, Options{NoFormat: true, NoCross: true, Exclude: []Region{{Page: 2, Box: "0,0.05,1,0.115"}}})
	if has(cut) {
		t.Errorf("a difference inside an excluded area is not reported: %+v", cut.Findings)
	}
	if n := strings.Join(cut.Notes, "\n"); !strings.Contains(n, "Исключаемых областей: 1") {
		t.Errorf("the run says what was left out:\n%s", n)
	}
	// only on the reference side: the РД line is still there, the ПД one is gone → reported as a difference of another kind, never silently dropped
	one := Compare([]*Document{pd}, []*Document{rd}, Options{NoFormat: true, NoCross: true, Exclude: []Region{{Side: SideRef, Page: 2, Box: "0,0.05,1,0.115"}}})
	if !has(one) {
		t.Errorf("excluding one side only must leave the other side's line to be reported: %+v", one.Findings)
	}
}
