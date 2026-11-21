package idcheck

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSamples runs the parsers against the real sample package
// (registry + КС-2 + act PDFs). The samples are customer documents and are
// not committed — point IDCHECK_SAMPLES at the "ИД, реестр, КС-2" folder:
//
//	IDCHECK_SAMPLES="$HOME/Загрузки/Пример ПД и РД/ИД, реестр, КС-2" go test ./idcheck -run Samples -v
func TestSamples(t *testing.T) {
	dir := os.Getenv("IDCHECK_SAMPLES")
	if dir == "" {
		t.Skip("IDCHECK_SAMPLES not set")
	}
	regFiles, _ := filepath.Glob(filepath.Join(dir, "Реестр ИД", "*.xlsx"))
	ksFiles, _ := filepath.Glob(filepath.Join(dir, "КС", "*.xlsx"))
	if len(regFiles) == 0 || len(ksFiles) == 0 {
		t.Fatalf("registry/КС-2 xlsx not found under %s", dir)
	}
	regData, _ := os.ReadFile(regFiles[0])
	reg, err := ParseRegistry(regData)
	if err != nil {
		t.Fatal(err)
	}
	ksData, _ := os.ReadFile(ksFiles[0])
	ks, err := ParseKS2(ksData)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("registry: %d rows, %d acts, %d positions; КС-2 №%s: %d rows",
		len(reg.Rows), len(reg.Acts()), len(reg.Positions()), ks.Number, len(ks.Rows))

	pos := reg.Positions()["2.14.1.9"]
	if pos == nil || pos.QtyKS2 == nil || *pos.QtyKS2 != 26.3 {
		t.Fatalf("position 2.14.1.9 col 11: %+v", pos)
	}
	ksQty := map[string]float64{}
	for _, r := range ks.Rows {
		ksQty[r.Position] = r.Qty
	}
	if ksQty["2.14.1.9"] != 26.3 || ksQty["2.14.2.3"] != 88 {
		t.Fatalf("КС-2 quantities: 2.14.1.9=%v 2.14.2.3=%v", ksQty["2.14.1.9"], ksQty["2.14.2.3"])
	}

	fs := CheckRegistryConsistency(reg)
	found := false
	for _, f := range fs {
		if f.Position == "2.14.2.6" && f.Field == "doc_number" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected act number 3/2.14.1.6 under position 2.14.2.6 to be flagged; got %d findings", len(fs))
	}

	sum := ReconcileKS2(reg, ks)
	for _, s := range sum {
		if s.Status != "ok" {
			t.Logf("check 8: %s %s registry=%.4f ks2=%v status=%s", s.Position, s.Unit, s.RegistrySum, deref(s.KS2Qty), s.Status)
		}
	}

	acts, _ := filepath.Glob(filepath.Join(dir, "АОСР", "*", "*", "*.pdf"))
	more, _ := filepath.Glob(filepath.Join(dir, "АОСР", "*", "*.pdf"))
	acts = append(acts, more...)
	matched := 0
	for _, a := range acts {
		num, date := ParseActFileName(a)
		if row := MatchRegistryRow(reg, num, date); row != nil {
			matched++
		} else {
			t.Logf("no registry row for file %s (key %s, %s)", filepath.Base(a), num.Key(), FmtDate(date))
		}
	}
	t.Logf("act files matched to registry: %d/%d", matched, len(acts))
	if matched == 0 {
		t.Fatal("no act file matched a registry row")
	}
}

func deref(p *float64) any {
	if p == nil {
		return nil
	}
	return *p
}
