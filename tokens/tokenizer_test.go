package tokens_test

import (
	"testing"

	"github.com/dwarvesf/go-sqlglot/tokens"
)

// helper used by all tokenizer tests
func tok(t *testing.T, sql string) []tokens.Token {
	t.Helper()
	cfg := tokens.DefaultConfig()
	got, err := tokens.Tokenize(sql, cfg)
	if err != nil {
		t.Fatalf("Tokenize(%q): %v", sql, err)
	}
	return got
}

func TestTokenizeEmpty(t *testing.T) {
	if got := tok(t, ""); len(got) != 0 {
		t.Errorf("empty: got %d tokens", len(got))
	}
}

func TestTokenizeWhitespace(t *testing.T) {
	if got := tok(t, "   \t\n  "); len(got) != 0 {
		t.Errorf("whitespace: got %d tokens", len(got))
	}
}

func TestTokenizeSinglePunct(t *testing.T) {
	cases := []struct {
		sql  string
		want tokens.TokenType
	}{
		{"(", tokens.LParen},
		{")", tokens.RParen},
		{",", tokens.Comma},
		{";", tokens.Semicolon},
		{"*", tokens.Star},
		{"+", tokens.Plus},
		{"-", tokens.Dash},
		{"=", tokens.Eq},
	}
	for _, tc := range cases {
		got := tok(t, tc.sql)
		if len(got) != 1 {
			t.Errorf("sql=%q: got %d tokens, want 1: %v", tc.sql, len(got), got)
			continue
		}
		if got[0].Type != tc.want {
			t.Errorf("sql=%q: type got %v, want %v", tc.sql, got[0].Type, tc.want)
		}
		if got[0].Text != tc.sql {
			t.Errorf("sql=%q: text got %q, want %q", tc.sql, got[0].Text, tc.sql)
		}
	}
}
