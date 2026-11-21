package rulesgo

import (
	"encoding/json"
	"testing"
)

func TestFlexibleID_StringAndNumber(t *testing.T) {
	const want uint64 = 509463708777775105
	var id flexibleID
	if err := json.Unmarshal([]byte(`"509463708777775105"`), &id); err != nil {
		t.Fatal(err)
	}
	if uint64(id) != want {
		t.Fatalf("string: got %d want %d", id, want)
	}
	id = 0
	if err := json.Unmarshal([]byte(`509463708777775105`), &id); err != nil {
		t.Fatal(err)
	}
	if uint64(id) != want {
		t.Fatalf("number: got %d want %d", id, want)
	}
}

func TestUint64FromAny_String(t *testing.T) {
	got := uint64FromAny("509463708777775105")
	if got != 509463708777775105 {
		t.Fatalf("got %d", got)
	}
}

func TestUint64FromAny_JSONNumber(t *testing.T) {
	got := uint64FromAny(json.Number("509463708777775105"))
	if got != 509463708777775105 {
		t.Fatalf("json.Number got %d", got)
	}
}

func TestUint64FromAny_ImpreciseFloat(t *testing.T) {
	var v interface{}
	if err := json.Unmarshal([]byte(`509463708777775105`), &v); err != nil {
		t.Fatal(err)
	}
	got := uint64FromAny(v)
	if got != 0 {
		t.Fatalf("imprecise float64 must not override chain IDs, got %d", got)
	}
}
