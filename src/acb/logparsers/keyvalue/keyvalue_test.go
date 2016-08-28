package keyvalue_test

import (
	"acb/logparsers/keyvalue"
	"strings"
	"testing"
)

// Ensure the scanner can scan tokens correctly.
func TestScanner_Scan(t *testing.T) {
	var tests = []struct {
		s   string
		tok keyvalue.Token
		lit string
	}{
		// Special tokens (EOF, ILLEGAL, WS)
		{s: ``, tok: keyvalue.EOF},
		{s: `#`, tok: keyvalue.ILLEGAL, lit: `#`},
		{s: ` `, tok: keyvalue.WS, lit: " "},
		{s: "\t", tok: keyvalue.WS, lit: "\t"},
		{s: "\n", tok: keyvalue.WS, lit: "\n"},

		// Misc characters
		{s: `=`, tok: keyvalue.EQUAL, lit: "="},

		// Identifiers
		{s: `foo`, tok: keyvalue.IDENT, lit: `foo`},

		{s: `"a b c"`, tok: keyvalue.STRING, lit: `a b c`},
		{s: `"abc"`, tok: keyvalue.STRING, lit: `abc`},
		{s: `"a"`, tok: keyvalue.STRING, lit: `a`},
		{s: `""`, tok: keyvalue.STRING, lit: ``},
		{s: `"\""`, tok: keyvalue.STRING, lit: `"`},
	}

	for i, tt := range tests {
		s := keyvalue.NewScanner(strings.NewReader(tt.s))
		tok, lit := s.Scan()
		if tt.tok != tok {
			t.Errorf("%d. %q token mismatch: exp=%q got=%q <%q>", i, tt.s, tt.tok, tok, lit)
		} else if tt.lit != lit {
			t.Errorf("%d. %q literal mismatch: exp=%q got=%q", i, tt.s, tt.lit, lit)
		}
	}
}
