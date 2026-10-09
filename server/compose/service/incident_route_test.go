package service

import "testing"

func TestParseRecordID(t *testing.T) {
	cases := []struct {
		in   interface{}
		want uint64
	}{
		{uint64(42), 42},
		{int64(42), 42},
		{42, 42},
		{float64(42), 42},
		{"42", 42},
		{"42.0", 42},
		{"", 0},
		{nil, 0},
	}
	for _, c := range cases {
		if got := ParseRecordID(c.in); got != c.want {
			t.Errorf("ParseRecordID(%v)=%d want %d", c.in, got, c.want)
		}
	}
}

func TestStringifyID(t *testing.T) {
	if got := stringifyID(float64(34)); got != "34" {
		t.Errorf("stringifyID(34.0)=%q", got)
	}
	if got := stringifyID(" 7 "); got != "7" {
		t.Errorf("stringifyID spaces=%q", got)
	}
}
