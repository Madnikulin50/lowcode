package ql

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// Ensure the scanner can scan tokens correctly.
func TestScanner_ScanSimple(t *testing.T) {
	var tests = []struct {
		s   string
		tok tokenCode
		lit string
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, tok: EOF},
		{s: `#`, tok: ILLEGAL, lit: `#`},
		{s: ` `, tok: WS, lit: " "},
		{s: "\t", tok: WS, lit: "\t"},
		{s: "\n", tok: WS, lit: "\n"},

		// Operators
		{s: `*`, tok: OPERATOR, lit: "*"},
		{s: `!=`, tok: OPERATOR, lit: "!="},
		{s: `<`, tok: OPERATOR, lit: "<"},
		{s: `>`, tok: OPERATOR, lit: ">"},
		{s: `>=`, tok: OPERATOR, lit: ">="},
		{s: `<>`, tok: OPERATOR, lit: "<>"},
		{s: `+`, tok: OPERATOR, lit: "+"},
		{s: `'fooo'`, tok: LSTRING, lit: "fooo"},
		{s: `'escaped \' quote'`, tok: LSTRING, lit: "escaped ' quote"},
		{s: `'double \\ escape'`, tok: LSTRING, lit: "double \\ escape"},
		{s: `12345`, tok: LNUMBER, lit: "12345"},

		// Identifiers
		{s: `foo`, tok: IDENT, lit: `foo`},
		{s: `Zx12_3U_-`, tok: IDENT, lit: `Zx12_3U_`},

		// Parenthesis
		{s: `(`, tok: PARENTHESIS_OPEN, lit: `(`},
		{s: `)`, tok: PARENTHESIS_CLOSE, lit: `)`},

		// Literals
		{s: `true`, tok: LBOOL, lit: `TRUE`},
	}

	for i, test := range tests {
		s := NewLexer(strings.NewReader(test.s))
		tok := s.Scan()
		if test.tok != tok.code {
			t.Errorf("%d. %q token mismatch: exp=%d got=%d <%q>", i, test.s, test.tok, tok.code, tok.literal)
		} else if test.lit != tok.literal {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, test.s, test.lit, tok.literal)
		}
	}
}

func TestScanner_ScanComplex(t *testing.T) {
	var tests = []struct {
		s      string
		tokens []tokenCode
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{`func(arg1, arg2)`,
			[]tokenCode{IDENT, PARENTHESIS_OPEN, IDENT, COMMA, WS, IDENT, PARENTHESIS_CLOSE}},
		{`arg1 * arg2`,
			[]tokenCode{IDENT, WS, OPERATOR, WS, IDENT}},
		{`date_format(created_at,'%Y')`,
			[]tokenCode{IDENT, PARENTHESIS_OPEN, IDENT, COMMA, LSTRING, PARENTHESIS_CLOSE}},
		{`foo LIKE 'abc%'`,
			[]tokenCode{IDENT, WS, OPERATOR, WS, LSTRING}},
		{`foo NOT LIKE 'abc%'`,
			[]tokenCode{IDENT, WS, OPERATOR, WS, OPERATOR, WS, LSTRING}},
		{`foo DESC`,
			[]tokenCode{IDENT, WS, KEYWORD}},
		{`year(now())-1`,
			[]tokenCode{IDENT, PARENTHESIS_OPEN, IDENT, PARENTHESIS_OPEN, PARENTHESIS_CLOSE, PARENTHESIS_CLOSE, OPERATOR, LNUMBER}},
		{`(device = 509925326227505153)`,
			[]tokenCode{PARENTHESIS_OPEN, IDENT, WS, OPERATOR, WS, LNUMBER, PARENTHESIS_CLOSE}},
	}

	for _, test := range tests {
		var tokens []Token
		s := NewLexer(strings.NewReader(test.s))
		for {
			tok := s.Scan()
			if tok.Is(EOF) {
				break
			}

			tokens = append(tokens, tok)
		}

		if len(tokens) != len(test.tokens) {
			t.Errorf("Collected tokens do not match (%v)", tokens)
		}

		for i := 0; i < len(tokens); i++ {
			if tokens[i].code != test.tokens[i] {
				t.Errorf("Input:     %s", test.s)
				t.Errorf("Expected:  %v", test.tokens)
				t.Errorf("Collected: %v", tokens)
				break
			}
		}
	}
}

func TestLexerNumberFollowedByParenDoesNotHang(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		s := NewLexer(strings.NewReader("509925326227505153)"))
		tok := s.Scan()
		if tok.code != LNUMBER || tok.literal != "509925326227505153" {
			done <- fmt.Errorf("number token code=%d lit=%q", tok.code, tok.literal)
			return
		}
		tok = s.Scan()
		if tok.code != PARENTHESIS_CLOSE {
			done <- fmt.Errorf("paren token code=%d lit=%q", tok.code, tok.literal)
			return
		}
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("lexer hung on number followed by ')'")
	}
}

func TestParseRecordIDFilterInParens(t *testing.T) {
	done := make(chan error, 1)
	go func() {
		n, err := NewParser().Parse("(device = 509925326227505153)")
		if err != nil {
			done <- err
			return
		}
		if n == nil {
			done <- fmt.Errorf("nil AST")
			return
		}
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("parser hung on (device = <id>)")
	}
}
